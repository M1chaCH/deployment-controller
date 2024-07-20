package auth

import (
	"github.com/gin-gonic/gin"
)

func Init(router *gin.Engine) {
	router.GET("/proxy/:uri", getAuthRequest) // uri is the uri from NGINX: original request "https://michu-tech.com/test/more?query='hereere'" -> uri:"/test/more" (request_uri would include query params)
}

/*
if shit goes bad in nginx:

https://stackoverflow.com/questions/55751365/nginx-auth-request-with-cookie
OR
# Perform the auth request
auth_request /auth;

# Add the cookie based on the auth response
auth_request_set $auth_cookie $upstream_http_set_cookie;
add_header Set-Cookie $auth_cookie;
*/
func getAuthRequest(c *gin.Context) {
	uri := c.Param("uri")

	var hasClientId bool
	var clientId string
	jwtString, err := c.Cookie(identityCookie)
	if err != nil { // cookie was not found
		hasClientId = false
	} else {
		token, err := parseIdentityToken(jwtString)
		if err != nil {
			hasClientId = false
		}
	}

	/*
		TODO:
		- check if clientId exists in request
			- if true: update Devices
		    - if false: create clientNewClientId
		- check if page requires auth (info is stored in DB, DB has start of URI)
	*/
}
