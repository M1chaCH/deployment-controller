package auth

import (
	"errors"
	"fmt"
	"github.com/M1chaCH/deployment-controller/framework"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"strings"
	"time"
)

const (
	LoginStateLoggedIn         = "logged-in"
	LoginStateLoggedOut        = "logged-out"
	LoginStateTwofactorWaiting = "two-factor-waiting"
)
const PrivatePagesDelimiter = ","

const idJwtCookieName = "Bearer"
const idJwtContextKey = "idToken"
const updatedIdJwtContextKey = "activeIdToken"

type IdentityToken struct {
	Mail         string `json:"mail,omitempty"`
	Admin        bool   `json:"admin,omitempty"`
	PrivatePages string `json:"pp,omitempty"`
	OriginIp     string `json:"oip"`
	OriginAgent  string `json:"agt"`
	LoginState   string `json:"lgs"`
	/*
		Used RegisteredClaims
		- Issuer -> ClientId
		- Subject -> UserId
		- ExpiresAt -> ExpiresAt
		- IssuedAt -> IssuedAt
	*/
	jwt.RegisteredClaims
}

// ToJwtString converts the data to a JwtToken with a configured secret
// does no app specific validation
func (data IdentityToken) ToJwtString() (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS384, data)
	secret := getSecret()
	return token.SignedString([]byte(secret))
}

// IdentityJwtMiddleware gets and parses the Identity cookie and adds it to the context as a struct
// does no app specific validation
func IdentityJwtMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenFailed := false
		tokenString, err := c.Cookie(idJwtCookieName)
		if err != nil {
			tokenFailed = true
		}

		token, err := parseIdentityToken(tokenString)
		if err != nil {
			tokenFailed = true
		}

		if tokenFailed {
			c.Set(idJwtContextKey, nil)
		} else {
			c.Set(idJwtContextKey, token)
		}

		c.Next()

		newToken, ok := getIdentityToken(c, updatedIdJwtContextKey)
		if !ok {
			fmt.Println("ERROR: found request with invalid updated ID token")
			return
		}

		config := framework.Config()
		newTokenString, err := newToken.ToJwtString()
		if err != nil {
			fmt.Printf("ERROR: could not parse ID token, %v\n", err)
			return
		}

		// keep the cookie as long as possible, don't want to lose the clientId
		c.SetCookie(idJwtCookieName, newTokenString, 400, "/", config.JWT.Domain, true, true)
	}
}

func createIdentityToken(
	clientId string,
	userId string,
	mail string,
	admin bool,
	loginState string,
	privatePages []string,
	ip string,
	userAgent string,
) IdentityToken {
	tokenLifetime := time.Duration(framework.Config().JWT.Lifetime)
	if tokenLifetime <= 0 {
		tokenLifetime = 24
	}
	tokenLifetime *= time.Hour

	return IdentityToken{
		Mail:         mail,
		Admin:        admin,
		PrivatePages: strings.Join(privatePages, PrivatePagesDelimiter),
		OriginIp:     ip,
		OriginAgent:  userAgent,
		LoginState:   loginState,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    clientId,
			Subject:   userId,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(tokenLifetime)),
		},
	}
}

func GetCurrentIdentityToken(c *gin.Context) (IdentityToken, bool) {
	return getIdentityToken(c, updatedIdJwtContextKey)
}

func getIdentityToken(c *gin.Context, key string) (IdentityToken, bool) {
	value, exists := c.Get(key)
	if !exists {
		return IdentityToken{}, false
	}

	token, ok := value.(IdentityToken)
	if !ok {
		fmt.Println("ERROR: found request with invalid ID token")
		return IdentityToken{}, false
	}

	return token, true
}

// parseIdentityToken parses the string to an IdentityToken
// fails if token has invalid format
// does no app specific validation
func parseIdentityToken(tokenString string) (*IdentityToken, error) {
	token, err := jwt.ParseWithClaims(tokenString, &IdentityToken{}, func(token *jwt.Token) (interface{}, error) {
		return getSecret(), nil
	})

	if err != nil {
		fmt.Printf("JWT could not be parsed: %v\n", err)
		return nil, errors.New("user unauthenticated")
	}

	if claims, ok := token.Claims.(*IdentityToken); ok && token.Valid {
		return claims, nil
	} else {
		return nil, errors.New("user unauthorized")
	}
}

func getSecret() string {
	config := framework.Config()
	if config.JWT.Secret == "" {
		fmt.Println("WARN: no JWT secret is provided, using default")
		config.JWT.Secret = "dGhpcyBzaG91bGQgbm90IGJlIHVzZWQ="
	}

	return config.JWT.Secret
}
