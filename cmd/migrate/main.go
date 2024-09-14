package main

import (
	"fmt"
	"github.com/VEDA95/OpenBoard-API/internal/config"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"log"
	"os"
	"os/exec"
	"strconv"
)

func main() {
	var conf config.MigrationConfig
	argsWithoutProgram := os.Args[1:]

	if len(argsWithoutProgram) > 2 {
		log.Panic("Too many arguments...")
	}

	if len(argsWithoutProgram) < 1 {
		log.Panic("Please provide one of the following arguments: init, up, down, step <count>")
	}

	if err := config.ParseConfig[config.MigrationConfig](&conf); err != nil {
		log.Panic(err)
	}

	mainAction := argsWithoutProgram[0]

	if mainAction == "init" {
		if len(argsWithoutProgram) < 2 {
			log.Panic("A name needs to be provided for the migration being created")
		}

		cmd := exec.Command("migrate", "create", "-ext", "sql", "-dir", "./migrations", "-seq", argsWithoutProgram[1])
		err := cmd.Run()

		if err != nil {
			log.Panic(err)
		}

		fmt.Println(fmt.Sprintf(`Migration files for migration: "%s" were successfully created!`, argsWithoutProgram[1]))
		return
	}

	migration, err := migrate.New("file://./migrations", conf.DBUrl)

	if err != nil {
		log.Panic(err)
	}

	if mainAction == "up" {
		err := migration.Up()

		if err != nil {
			log.Panic(err)
		}

		fmt.Println("UP migration completed!")
		return
	}

	if mainAction == "down" {
		err := migration.Down()

		if err != nil {
			log.Panic(err)
		}

		fmt.Println("DOWN migration completed!")
		return
	}

	if mainAction == "step" {
		if len(argsWithoutProgram) < 2 {
			log.Panic("Step count needs to be provided")
		}

		stepCount, err := strconv.ParseInt(argsWithoutProgram[1], 0, 64)

		if err != nil {
			log.Panic("Unable to parse the value provided for the step count")
		}

		err = migration.Steps(int(stepCount))

		if err != nil {
			log.Panic(err)
		}

		fmt.Println("STEP migration completed!")
		return
	}
}
