package internal

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config mirrors .repolint.yml. All fields optional.
type Config struct {
	ExpectedLicense string            `yaml:"expected_license"` // e.g. "Apache-2.0"
	Severity        map[string]string `yaml:"severity"`
	Disabled        map[string]bool   `yaml:"-"`
	Disable         []string          `yaml:"disable"`
	Remote          bool              `yaml:"remote"` // also run gh-metadata checks
	ExtraFiles      []string          `yaml:"require_files"`
}

// DefaultConfig returns the defaults.
func DefaultConfig() Config {
	return Config{Disable: []string{}}
}

// LoadConfig reads .repolint.yml when present.
func LoadConfig(root, configPath string) (Config, error) {
	cfg := DefaultConfig()
	p := configPath
	if p == "" {
		p = ".repolint.yml"
	}
	if !filepath.IsAbs(p) {
		p = filepath.Join(root, p)
	}
	b, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) && configPath == "" {
			cfg.Disabled = map[string]bool{}
			return cfg, nil
		}
		return cfg, fmt.Errorf("cannot read %s: %w", p, err)
	}
	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return cfg, fmt.Errorf("malformed .repolint.yml: %w", err)
	}
	cfg.Disabled = map[string]bool{}
	for _, c := range cfg.Disable {
		cfg.Disabled[c] = true
	}
	return cfg, nil
}
