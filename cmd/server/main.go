package main

import (
	"fmt"
	"github.com/VEDA95/OpenBoard-API/internal/api/http/routes"
	"github.com/VEDA95/OpenBoard-API/internal/api/http/validators"
	"github.com/VEDA95/OpenBoard-API/internal/auth"
	"github.com/VEDA95/OpenBoard-API/internal/config"
	"github.com/VEDA95/OpenBoard-API/internal/db"
	"github.com/VEDA95/OpenBoard-API/internal/settings"
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

	app := fiber.New()
	providerEntries := []auth.ProviderRegistrationEntry{
		{
			Key:    "local",
			Name:   "local",
			Config: make(map[string]interface{}),
		},
	}

	defer db.Instance.Close()
	settings.InitializeSettingsInstance([]string{"auth"})
	auth.InitializeProvidersInstance(providerEntries)
	validators.InitializeValidatorInstance()
	app.Get("/", routes.HelloWorld)
	app.Post("/auth/login", routes.LocalLogin)

	if err := app.Listen(fmt.Sprintf("%s:%s", conf.Host, conf.Port)); err != nil {
		log.Panic(err)
	}
}
