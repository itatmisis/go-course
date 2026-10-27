package settings

import (
	"fmt"
	"net"
	"net/url"
	"strconv"

	"github.com/caarlos0/env/v11"
)

type Postgres struct {
	Database string `env:"DATABASE" envDefault:"db_main"`
	Host     string `env:"HOST"     envDefault:"localhost"`
	Port     int    `env:"PORT"     envDefault:"5432"`
	Username string `env:"USERNAME" envDefault:"db_main"`
	Password string `env:"PASSWORD" envDefault:"db_main"`
}

func (p Postgres) URL() string {
	databaseURL := url.URL{
		Scheme: "postgres",
		User:   url.UserPassword(p.Username, p.Password),
		Host:   net.JoinHostPort(p.Host, strconv.Itoa(p.Port)),
		Path:   p.Database,
	}

	return databaseURL.String()
}

type HTTP struct {
	Host    string `env:"HOST"     envDefault:"0.0.0.0"`
	Port    int    `env:"PORT"     envDefault:"8080"`
	BaseURL string `env:"BASE_URL" envDefault:"http://localhost:8080"`
}

func (h HTTP) Address() string {
	return net.JoinHostPort(h.Host, strconv.Itoa(h.Port))
}

type Settings struct {
	Postgres Postgres `envPrefix:"PG__"`
	HTTP     HTTP     `envPrefix:"HTTP__"`
}

func Load() (Settings, error) {
	appSettings, err := env.ParseAsWithOptions[Settings](env.Options{Prefix: "APP_"})
	if err != nil {
		return Settings{}, fmt.Errorf("parse env: %w", err)
	}

	return appSettings, nil
}
