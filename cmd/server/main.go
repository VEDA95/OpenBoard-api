package main

import (
	"fmt"
	"github.com/VEDA95/OpenBoard-API/internal/api/http/routes"
	"github.com/VEDA95/OpenBoard-API/internal/api/http/validators"
	"github.com/VEDA95/OpenBoard-API/internal/auth"
	"github.com/VEDA95/OpenBoard-API/internal/config"
	"github.com/VEDA95/OpenBoard-API/internal/db"
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

	app := fiber.New(fiber.Config{
		ErrorHandler: util.ErrorHandler,
	})
	providerEntries := []auth.ProviderRegistrationEntry{
		{
			Key:    "local",
			Name:   "local",
			Config: make(map[string]interface{}),
		},
	}

	settings.InitializeSettingsInstance([]string{"auth"})
	auth.InitializeProvidersInstance(providerEntries)
	validators.InitializeValidatorInstance()
	app.Get("/", routes.HelloWorld)

	authGroup := app.Group("/auth")

	authGroup.Post("/login", routes.LocalLogin)
	authGroup.Get("/@me", routes.Me)

	userGroup := app.Group("/api/users")

	userGroup.Get("/", routes.ShowUsers)
	userGroup.Post("/", routes.CreateUser)
	userGroup.Get("/:id", routes.ShowUser)
	userGroup.Put("/:id", routes.UpdateUser)
	userGroup.Delete("/:id", routes.DeleteUser)

	if err := app.Listen(fmt.Sprintf("%s:%s", conf.Host, conf.Port)); err != nil {
		log.Panic(err)
	}
}
