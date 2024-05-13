package config

import "fmt"

type Config struct {
	PostgresHost          string `env:"POSTGRES_HOST" envDefault:"localhost"`
	PostgresUser          string `env:"POSTGRES_USER"         envDefault:"bashrun"`
	PostgresPassword      string `env:"POSTGRES_PASSWORD"     envDefault:"bashrun"`
	PostgresDB            string `env:"POSTGRES_DB"           envDefault:"bashrun"`
	PostgresPort          int    `env:"POSTGRES_PORT"         envDefault:"5432"`
	ServicePort           int    `env:"SERVICE_PORT"          envDefault:"8080"`
	ServiceHost           string `env:"SERVICE_HOST"          envDefault:"0.0.0.0"`
	MigrationsPath        string `env:"MIGRATIONS_PATH"       envDefault:"migrations"`
	LogFilePath           string `env:"LOG_FILE_PATH"         envDefault:"logfile.log"`
	MaxConcurrentCommands int64  `env:"MAX_CONCURRENT_COMMANDS" envDefault:"100"`
}

func (c *Config) DSN() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
		c.PostgresUser,
		c.PostgresPassword,
		c.PostgresHost,
		c.PostgresPort,
		c.PostgresDB,
	)
}
