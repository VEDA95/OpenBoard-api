package main

import (
	"fmt"
	"github.com/VEDA95/OpenBoard-API/internal/api/http/routes"
	"github.com/VEDA95/OpenBoard-API/internal/config"
	"github.com/VEDA95/OpenBoard-API/internal/db"
	"github.com/gofiber/fiber/v2"
	"log"
)

func main() {
	var conf config.ServerConfig

	if err := config.ParseConfig[config.ServerConfig](&conf); err != nil {
		log.Panic(err)
	}

	if err := db.InitialDBConnection(); err != nil {
		log.Panic(err)
	}

	app := fiber.New()

	defer db.DBInstance.Close()
	app.Get("/", routes.HelloWorld)

	if err := app.Listen(fmt.Sprintf("%s:%s", conf.Host, conf.Port)); err != nil {
		log.Panic(err)
	}
}
