package api

import (
	"os"
	"time"

	"github.com/spf13/viper"

	"github.com/fog/gerrit-cli/pkg/gerrit"
)

var gerritClient *gerrit.Client

func Client(cfg gerrit.Config) *gerrit.Client {
	if gerritClient != nil {
		return gerritClient
	}
	if cfg.Server == "" {
		cfg.Server = viper.GetString("server")
	}
	if cfg.Username == "" {
		cfg.Username = viper.GetString("username")
	}
	if cfg.Password == "" {
		cfg.Password = viper.GetString("password")
	}
	if cfg.Password == "" {
		cfg.Password = os.Getenv("GERRIT_HTTP_PASSWORD")
	}
	if !cfg.NoAuthPrefix {
		cfg.NoAuthPrefix = viper.GetBool("no_auth_prefix")
	}
	gerritClient = gerrit.NewClient(cfg, gerrit.WithTimeout(15*time.Second))
	return gerritClient
}

func DefaultClient(debug bool) *gerrit.Client {
	return Client(gerrit.Config{Debug: debug})
}
