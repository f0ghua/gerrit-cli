package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Profile holds connection settings for a single Gerrit instance.
type Profile struct {
	Server       string `yaml:"server"`
	Username     string `yaml:"username"`
	Password     string `yaml:"password,omitempty"`
	Project      string `yaml:"project,omitempty"`
	NoAuthPrefix bool   `yaml:"no_auth_prefix,omitempty"`
}

// Config is kept as an alias for backward compatibility.
type Config = Profile

// MultiConfig holds named profiles and a default profile name.
type MultiConfig struct {
	Default  string              `yaml:"default"`
	Profiles map[string]*Profile `yaml:"profiles"`
}

// ResolveProfile returns the profile for the given name.
// If name is empty, the default profile is used.
func (mc *MultiConfig) ResolveProfile(name string) (*Profile, error) {
	if name == "" {
		name = mc.Default
	}
	if name == "" {
		return nil, fmt.Errorf("no profile specified and no default set")
	}
	p, ok := mc.Profiles[name]
	if !ok {
		return nil, fmt.Errorf("profile %q not found", name)
	}
	return p, nil
}

func ConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".gerrit")
}

func ConfigFilePath() string {
	return filepath.Join(ConfigDir(), ".config.yml")
}

func Exists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// LoadMulti reads the config file. If it contains the new multi-profile
// format (has a "profiles" key), it returns it directly. Otherwise it
// treats the file as a legacy flat config and wraps it as a single
// "default" profile.
func LoadMulti() (*MultiConfig, error) {
	data, err := os.ReadFile(ConfigFilePath())
	if err != nil {
		return nil, err
	}

	// Probe for multi-profile format by checking for "profiles" key.
	var probe struct {
		Profiles map[string]*Profile `yaml:"profiles"`
	}
	if err := yaml.Unmarshal(data, &probe); err != nil {
		return nil, err
	}
	if probe.Profiles != nil {
		var mc MultiConfig
		if err := yaml.Unmarshal(data, &mc); err != nil {
			return nil, err
		}
		return &mc, nil
	}

	// Legacy flat format — wrap as single "default" profile.
	var p Profile
	if err := yaml.Unmarshal(data, &p); err != nil {
		return nil, err
	}
	return &MultiConfig{
		Default:  "default",
		Profiles: map[string]*Profile{"default": &p},
	}, nil
}

// SaveMulti writes the multi-profile config to disk.
func SaveMulti(mc *MultiConfig) error {
	if err := os.MkdirAll(ConfigDir(), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(mc)
	if err != nil {
		return err
	}
	return os.WriteFile(ConfigFilePath(), data, 0o600)
}

// Load reads a legacy flat config. Kept for backward compatibility.
func Load() (*Config, error) {
	data, err := os.ReadFile(ConfigFilePath())
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// Save writes a legacy flat config. Kept for backward compatibility.
func Save(cfg *Config) error {
	if err := os.MkdirAll(ConfigDir(), 0o755); err != nil {
		return err
	}
	data, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	return os.WriteFile(ConfigFilePath(), data, 0o600)
}
