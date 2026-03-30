// services/auth/internal/infra/config/config.go
package config

import (
	"context"
	"embed"
	"io/fs"

	pfconfig "github.com/piprim/mmw/platform/config"
	pfslog "github.com/piprim/mmw/platform/slog"
	"github.com/rotisserie/eris"
)

//go:embed configs/*.toml
var embeddedFS embed.FS

// getConfigFS returns the filesystem to use for reading config files.
// This is a variable so it can be mocked in tests.
var getConfigFS = func() fs.FS {
	return embeddedFS
}

type JWT struct {
	Secret string `env:"JWT_SECRET" json:"-"`
}

type Config struct {
	pfconfig.Base
	Database *pfconfig.Database `mapstructure:"database"`
	AppName  string             `env:"APP_NAME"`
	Server   *pfconfig.Server   `mapstructure:"server"`
	LogLevel pfslog.LogLevel    `mapstructure:"log-level"`
	JWT      JWT                `json:"-"`
}

var conf *Config

// Load reads the TOML files and automatically overrides them with Env Vars
// Load loads the configurations from embedded files:
// - configs/default.toml
// - configs/<APP_ENV>.toml if exist
// If envs is not nil, automatically overrides them with Env Vars.
func Load(ctx context.Context, envprefix string) (*Config, error) {
	if conf != nil {
		return conf, nil
	}

	conf := new(Config)

	configFS := getConfigFS()
	err := pfconfig.NewContext(ctx, configFS, envprefix).Fill(conf)
	if err != nil {
		return nil, eris.Wrap(err, "error filling config")
	}

	if conf.Database.Password == "" {
		return nil, eris.New(envprefix + "DB_PASSWORD environment variable is required")
	}

	if conf.JWT.Secret == "" {
		return nil, eris.New("jwt secret is empty")
	}

	return conf, nil
}
