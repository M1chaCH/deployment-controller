package main

import (
	"fmt"
	"github.com/M1chaCH/deployment-controller/auth"
	"github.com/M1chaCH/deployment-controller/data/clients"
	"github.com/M1chaCH/deployment-controller/data/pageaccess"
	"github.com/M1chaCH/deployment-controller/framework"
	"github.com/M1chaCH/deployment-controller/framework/config"
	"github.com/M1chaCH/deployment-controller/framework/logs"
	"github.com/M1chaCH/deployment-controller/location"
	"github.com/M1chaCH/deployment-controller/mail"
	"github.com/M1chaCH/deployment-controller/rest"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.elastic.co/apm/module/apmgin/v2"
	"strings"
)

/*
TODO:
Middlewares
1. Error, handle errors and logs critical stuff to elastic, rollback transaction if exists in context
2. ✔ Identity, add IdentityId to context (maybe even add Id to localstorage if determined)
3. ✔ Transaction, add transaction to context, commit transaction if success
4. ✔ Auth, check if endpoint needs validation and check JWT Token

CORS
- ✔ only allow request from michu-tech.com...

Clean Configs
- ✔ create yml file
- ✔ properly read and parse file
- ✔ https://github.com/go-yaml/yaml

Endpoints
- ✔ Auth (TwoFactor!, Login, LoggedInUserData, IsLoggedIn)
- ✔ Admin (user X Pages, remove user, block user, change password)
- Health (mini stats for health)
- ✔ Contact

Login
- ✔ Mail TwoFacture for new Agent
- ✔ Access granted for agent and ip
- ✔ E-Mail X Password / E-Mail X Code via Auth App

Mail
- send mails for login
- send mails if issues were detected (stats or errors)
- ✔ send mails for contact requests (limit rate)

Identity
- if user logged in -> userId or similar
- else -> Agent X IP -> store Id in agent -> use stored Id, so agent / ip can change
- track location from IP (maybe even from nginx plugin)

Docker

Other
- REBUILD Frontend with SvelteKit (SSR=false)
	- redesign home (pages with access large others small)
- OAuth2.0 for other apps?
	- would be cool, this way I only have to do this shit once
- report system metrics to elastic (ram, cpu, temperature, storage, ...) (https://www.elastic.co/guide/en/elasticsearch/reference/current/docs.html)
- nginx logs to elastic
- specific rate limiting?
- login with google, apple, github...?
*/

func main() {
	cnf := config.Config()
	host := cnf.Host
	port := cnf.Port

	logs.InitLogging()

	auth.MakeSureAdminExists()
	location.InitScheduledLocationCheck()
	initCaches()

	engine := gin.New()
	engine.Use(apmgin.Middleware(engine))
	engine.Use(cors.New(cors.Config{
		AllowOrigins:     strings.Split(cnf.Cors.Origins, ","),
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Length", "Content-Type", "Cookie", "Set-Cookie"},
		AllowCredentials: true,
	}))

	engine.Use(framework.TransactionMiddleware())
	engine.Use(auth.IdentityJwtMiddleware())
	engine.Use(auth.AuthenticationMiddleware())

	initRoutes(engine)

	err := engine.Run(fmt.Sprintf("%s:%s", host, port))
	if err != nil {
		logs.Error(nil, "could not start server: %v", err)
	}
}

func initCaches() {
	mail.InitTemplates()
	location.InitCache()
	clients.InitCache()
	pageaccess.InitCache()
}

func initRoutes(router *gin.Engine) {
	openEndpoints := router.Group("/open")
	rest.InitOpenEndpoints(openEndpoints)

	adminEndpoints := router.Group("/admin")
	adminEndpoints.Use(auth.AdminAuthorisationMiddleware())
	rest.InitAdminEndpoints(adminEndpoints)

	auth.InitAuthRequest(router)
}
