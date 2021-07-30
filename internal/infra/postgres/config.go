package postgres

import (
	"fmt"
	"time"
)

type Config struct {
	Host     string `long:"host" env:"HOST" description:"Postgres host" required:"yes"`
	Port     int    `long:"port" env:"PORT" description:"Postgres port" required:"yes"`
	User     string `long:"user" env:"USER" description:"Postgres user" required:"yes"`
	Password string `long:"password" env:"PASSWORD" description:"Postgres password" required:"yes"`
	Name     string `long:"name" env:"NAME" description:"Postgres name" required:"yes"`

	MaxOpenConns    int           `long:"max-open-conns" env:"MAX_OPEN_CONNS" default:"25" description:"maximum of open database connections"`
	MaxIdleConns    int           `long:"max-idle-conns" env:"MAX_IDLE_CONNS" default:"10" description:"maximum of idle database connections"`
	ConnMaxLifeTime time.Duration `long:"conn-max-life-time" env:"CONN_MAX_LIFE_TIME" default:"5m" description:"database max connection life time"`

	MigrationsSourceURL string `long:"migrations-source-url" env:"MIGRATIONS_SOURCE_URL" default:"file://migrations"`
}

func (c *Config) ConnectionString() string {
	uri := fmt.Sprintf("host=%s port=%d user=%s dbname=%s password=%s sslmode=disable",
		c.Host, c.Port,
		c.User, c.Name,
		c.Password,
	)

	return uri
}
