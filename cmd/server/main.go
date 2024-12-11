package main

import (
	"fmt"
	"github.com/VEDA95/OpenBoard-API/internal/api/http/middleware"
	"github.com/VEDA95/OpenBoard-API/internal/api/http/routes"
	"github.com/VEDA95/OpenBoard-API/internal/api/http/validators"
	"github.com/VEDA95/OpenBoard-API/internal/auth"
	"github.com/VEDA95/OpenBoard-API/internal/config"
	"github.com/VEDA95/OpenBoard-API/internal/db"
	"github.com/VEDA95/OpenBoard-API/internal/email"
	"github.com/VEDA95/OpenBoard-API/internal/settings"
	"github.com/VEDA95/OpenBoard-API/internal/util"
	"github.com/gofiber/fiber/v2"
	"log"
)

func main() {
	var conf config.ServerConfig

	if err := config.ParseConfig[config.ServerConfig](&conf); err != nil {
		log.Panic(err)
	}

	if err := db.InitializeDBInstance(); err != nil {
		log.Panic(err)
	}

	providerEntries := []auth.ProviderRegistrationEntry{
		{
			Key:    "local",
			Name:   "local",
			Config: make(map[string]interface{}),
		},
	}

	settings.InitializeSettingsInstance([]string{"auth", "notification"})
	auth.InitializeProvidersInstance(providerEntries)
	auth.InitializeMultiAuthMethodStore()
	validators.InitializeValidatorInstance()

	if err := email.InitializeMailClient(); err != nil {
		log.Panic(err)
	}

	app := fiber.New(fiber.Config{
		ErrorHandler: util.ErrorHandler,
	})

	app.Get("/", routes.HelloWorld)

	apiGroup := app.Group("/api")
	authGroup := app.Group("/auth")
	userGroup := apiGroup.Group("/users")
	mfaGroup := authGroup.Group("/mfa")
	registerGroup := mfaGroup.Group("/register")
	challengeGroup := mfaGroup.Group("/challenge")

	apiGroup.Get("/settings/:name", routes.SettingsGET)
	apiGroup.Put("/settings/:name", routes.SettingsPUT)
	userGroup.Use(middleware.AuthenticationMiddleware)
	userGroup.Get("/", routes.ShowUsers)
	userGroup.Post("/", routes.CreateUser)
	userGroup.Get("/:id", routes.ShowUser)
	userGroup.Put("/:id", routes.UpdateUser)
	userGroup.Delete("/:id", routes.DeleteUser)
	userGroup.Get("/@me", routes.Me)
	userGroup.Post("/@me/password/unlock", routes.PasswordResetUnlock)
	userGroup.Post("/@me/password/reset", routes.PasswordReset)
	authGroup.Post("/login", routes.LocalLogin)
	authGroup.Post("/refresh", routes.LocalRefresh)
	authGroup.Post("/logout", routes.LocalLogout)
	mfaGroup.Get("/methods", routes.GETMFAMethods)
	mfaGroup.Post("/methods", routes.SelectMFAMethod)
	registerGroup.Post("/webauthn/create", routes.CreateWebAuthnAuthMethodStart)
	registerGroup.Post("/webauthn/verify", routes.CreateWebAuthnAuthMethodEnd)
	registerGroup.Post("/otp/create", routes.CreateOTPAuthMethodStart)
	registerGroup.Post("/otp/verify", routes.CreateOTPAuthMethodEnd)
	registerGroup.Post("/authenticator/create", routes.CreateAuthenticatorAuthMethodStart)
	registerGroup.Post("/authenticator/verify", routes.CreateAuthenticatorAuthMethodEnd)
	challengeGroup.Post("/webauthn/create", routes.CreateWebAuthnChallengeStart)
	challengeGroup.Post("/webauthn/verify", routes.CreateWebAuthnChallengeEnd)
	challengeGroup.Post("/otp/create", routes.CreateOTPChallengeStart)
	challengeGroup.Post("/otp/verify", routes.CreateOTPChallengeEnd)
	challengeGroup.Post("/authenticator/create", routes.CreateAuthenticatorChallengeStart)
	challengeGroup.Post("/authenticator/verify", routes.CreateAuthenticatorChallengeEnd)

	if err := app.Listen(fmt.Sprintf("%s:%s", conf.Host, conf.Port)); err != nil {
		log.Panic(err)
	}
}
