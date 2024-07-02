package main

import (
	"fmt"
	"github.com/M1chaCH/deployment-controller/pages"
	"github.com/gin-gonic/gin"
	"os"
)

/*
TODO:
Middlewares
1. Error, handle errors and log critical stuff to elastic, rollback transaction if exists in context
2. Identity, add IdentityId to context (maybe even add Id to localstorage if determined)
3. Transaction, add transaction to context, commit transaction if success
4. Auth, check if endpoint needs validation and check JWT Token

CORS!!
- only allow request from michu-tech.com...

Endpoints
- Auth (TwoFacture!, Login, LoggedInUserData, IsLoggedIn)
- Admin (User X Pages, remove user, block user, change password)
- Health (mini stats for health)
- Contact

Login
- Mail TwoFacture for new Agent
- Access granted for agent and ip
- E-Mail X Password / E-Mail X Code via Mail

Mail
- send mails for login
- send mails for TwoFacture token
- send mails if issues were detected (stats or errors)
- send mails for contact requests (limit rate)

Identity
- if user logged in -> userId or similar
- else -> Agent X IP -> store ID in agent -> use stored ID, so agent / ip can change
- track locations from IP (maybe even from nginx plugin)

Other
- REBUILD Frontend with SvelteKit (SSR=false)
	- redesign home (pages with access large others small)
- OAuth2.0 for other apps?
	- would be cool, this way I only have to do this shit once
- report system metrics to elastic (ram, cpu, temperature, storage, ...)
- nginx logs to elastic
- specific rate limiting?
- login with google, apple, github...?
*/

func main() {
	host := os.Getenv("APP_HOST")
	port := os.Getenv("APP_PORT")

	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})
	pages.Init(router)

	err := router.Run(fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		fmt.Println("could not start Webserver", err)
	}
}
