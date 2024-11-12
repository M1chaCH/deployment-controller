package auth

import (
	"errors"
	"fmt"
	"github.com/M1chaCH/deployment-controller/framework"
	"github.com/M1chaCH/deployment-controller/framework/logs"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"time"
)

const (
	LoginStateLoggedIn          = "logged-in"
	LoginStateLoggedOut         = "logged-out"
	LoginStateTwofactorWaiting  = "two-factor-waiting"
	LoginStateOnboardingWaiting = "onboarding-waiting"
)

const idJwtCookieName = "MTAuth"
const idJwtContextKey = "idToken"
const updatedIdJwtContextKey = "activeIdToken"
const JwtListDelimiter = "&&"

type IdentityToken struct {
	Mail        string `json:"mail,omitempty"`
	Admin       bool   `json:"admin,omitempty"`
	OriginIp    string `json:"oip"`
	OriginAgent string `json:"agt"`
	LoginState  string `json:"lgs"`
	UserId      string `json:"uid,omitempty"`
	/*
		Used RegisteredClaims
		- Issuer -> ClientId
		- ExpiresAt -> ExpiresAt
		- IssuedAt -> IssuedAt
	*/
	jwt.RegisteredClaims
}

// RespondWithCookie creates a JSON response and appends the ID cookie. After this you can't apply any changes to the response in the context.
func RespondWithCookie(c *gin.Context, code int, obj any) {
	AppendJwtToken(c)
	c.JSON(code, obj)
}

func AbortWithCooke(c *gin.Context, code int, message string) {
	AppendJwtToken(c)
	c.AbortWithStatusJSON(code, gin.H{"message": message})
}

func GetCurrentIdentityToken(c *gin.Context) (IdentityToken, bool) {
	return getIdentityToken(c, updatedIdJwtContextKey)
}

func SetCurrentIdentityToken(c *gin.Context, token IdentityToken) {
	c.Set(updatedIdJwtContextKey, token)
}

// ToJwtString converts the data to a JwtToken with a configured secret
// does no app specific validation
func (data IdentityToken) ToJwtString() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS384, data)
	return token.SignedString(getSecret())
}

// parseIdentityToken parses the string to an IdentityToken
// fails if tokens has invalid format
// does no app specific validation
func parseIdentityToken(tokenString string) (IdentityToken, error) {
	var claims IdentityToken
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		return getSecret(), nil
	})
	if err != nil {
		logs.Warn(fmt.Sprintf("JWT could not be parsed: %v", err))
		return IdentityToken{}, errors.New("user unauthenticated")
	}
	if !token.Valid {
		logs.Warn("provided token ID is not valid!")
		return IdentityToken{}, errors.New("user unauthenticated")
	}

	expiresClaim, err := claims.GetExpirationTime()
	if err != nil || expiresClaim.Before(time.Now()) {
		logs.Warn("JWT has expired or the expiration date could not be parsed")
		return IdentityToken{}, errors.New("user unauthenticated")
	}

	if claims.IssuedAt.After(time.Now()) {
		logs.Warn("JWT token was issued in the future")
		return IdentityToken{}, errors.New("user unauthenticated")
	}

	return claims, nil
}

func createIdentityToken(
	clientId string,
	userId string,
	mail string,
	admin bool,
	loginState string,
	ip string,
	userAgent string,
) IdentityToken {
	return IdentityToken{
		Mail:        mail,
		Admin:       admin,
		OriginIp:    ip,
		OriginAgent: userAgent,
		LoginState:  loginState,
		UserId:      userId,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    clientId,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: getNewTokenExpirationTime(),
		},
	}
}

func getIdentityToken(c *gin.Context, key string) (IdentityToken, bool) {
	value, exists := c.Get(key)
	if !exists || value == nil {
		return IdentityToken{}, false
	}

	token, ok := value.(IdentityToken)
	if !ok {
		return IdentityToken{}, false
	}

	return token, true
}

func getSecret() []byte {
	config := framework.Config()
	if config.JWT.Secret == "" {
		logs.Warn("no JWT secret is provided, using default")
		config.JWT.Secret = "dGhpcyBzaG91bGQgbm90IGJlIHVzZWQ="
	}

	return []byte(config.JWT.Secret)
}

func getNewTokenExpirationTime() *jwt.NumericDate {
	tokenLifetime := time.Duration(framework.Config().JWT.Lifetime)
	if tokenLifetime <= 0 {
		tokenLifetime = 24
	}
	tokenLifetime *= time.Hour
	return jwt.NewNumericDate(time.Now().Add(tokenLifetime))
}
