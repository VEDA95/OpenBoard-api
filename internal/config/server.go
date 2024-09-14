package config

type ServerConfig struct {
	Host  string `env:"HOST"`
	Port  string `env:"PORT"`
	DBUrl string `env:"DB_URL"`
}
