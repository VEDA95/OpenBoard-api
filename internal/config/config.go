package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"os"
)

func ParseConfig[Conf any](conf *Conf) error {
	envType := os.Getenv("ENV_TYPE")

	if len(envType) == 0 {
		envType = "development"
	}

	if envType == "development" {
		if err := cleanenv.ReadConfig("./env/development.env", conf); err != nil {
			return err
		}
	}

	if envType == "production" {
		if err := cleanenv.ReadConfig("./env/production.env", conf); err != nil {
			return err
		}
	}

	if err := cleanenv.ReadConfig("./env/.env", conf); err != nil {
		return err
	}

	return nil
}
