// Copyright (c) Kyle Huggins
// SPDX-License-Identifier: BSD-3-Clause

// Package config provides XDG-compliant configuration loading for TaskKit.
package config

import (
	_ "embed"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/hugginsio/taskkit/urgency"
	"gopkg.in/yaml.v3"
)

//go:embed default.yaml
var defaultConfigYAML []byte

// ErrNotFound is returned when no configuration file can be located.
var ErrNotFound = errors.New("config: file not found")

// DefaultReports are the built-in report definitions, sourced from default.yaml.
// User-defined reports in the config file override entries with the same name.
var DefaultReports = defaults().Reports

// Config holds the TaskKit configuration.
type Config struct {
	Database string            `yaml:"database"`
	Urgency  urgency.Weights   `yaml:"urgency"`
	Reports  map[string]string `yaml:"reports"`
}

// DefaultPath returns the path where Load will look for a config file:
// $TASKKIT_CONFIG if set, otherwise $XDG_CONFIG_HOME/taskkit/config.yaml.
func DefaultPath() string {
	if path := os.Getenv("TASKKIT_CONFIG"); path != "" {
		return path
	}

	return filepath.Join(xdg.ConfigHome, "taskkit", "config.yaml")
}

// CreateDefault writes the embedded default configuration to DefaultPath,
// creating parent directories as needed. Returns the path written.
// Returns an error if the file already exists.
func CreateDefault() (string, error) {
	path := DefaultPath()
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return "", fmt.Errorf("config: mkdir: %w", err)
	}

	f, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o600)
	if err != nil {
		return "", fmt.Errorf("config: create %s: %w", path, err)
	}

	defer f.Close()

	if _, err := f.Write(defaultConfigYAML); err != nil {
		return "", fmt.Errorf("config: write %s: %w", path, err)
	}

	return path, nil
}

// Load resolves the configuration file path and loads it.
// It checks $TASKKIT_CONFIG first, then $XDG_CONFIG_HOME/taskkit/config.yaml.
// Returns ErrNotFound if neither path exists.
func Load() (*Config, error) {
	return LoadFrom(DefaultPath())
}

// LoadFrom loads a Config from the given file path.
// Returns ErrNotFound if the file does not exist.
func LoadFrom(path string) (*Config, error) {
	f, err := os.Open(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, path)
	}

	if err != nil {
		return nil, fmt.Errorf("config: open %s: %w", path, err)
	}

	defer f.Close()

	var user Config
	if err := yaml.NewDecoder(f).Decode(&user); err != nil {
		return nil, fmt.Errorf("config: decode %s: %w", path, err)
	}

	cfg := defaults()
	cfg.overlay(&user)
	return cfg, nil
}

// defaults returns a Config populated from the embedded default.yaml.
func defaults() *Config {
	var cfg Config
	_ = yaml.Unmarshal(defaultConfigYAML, &cfg)
	return &cfg
}

// overlay applies non-zero values from user on top of cfg (the defaults).
func (cfg *Config) overlay(user *Config) {
	if user.Database != "" {
		cfg.Database = user.Database
	}

	// Urgency: non-zero user values override defaults field by field.
	u := user.Urgency
	if u.Deadline != 0 {
		cfg.Urgency.Deadline = u.Deadline
	}

	if u.Scheduled != 0 {
		cfg.Urgency.Scheduled = u.Scheduled
	}

	if u.Age != 0 {
		cfg.Urgency.Age = u.Age
	}

	if u.AgeNorm != 0 {
		cfg.Urgency.AgeNorm = u.AgeNorm
	}

	if u.Tags != 0 {
		cfg.Urgency.Tags = u.Tags
	}

	if u.Waiting != 0 {
		cfg.Urgency.Waiting = u.Waiting
	}

	if u.Blocked != 0 {
		cfg.Urgency.Blocked = u.Blocked
	}

	if u.Blocking != 0 {
		cfg.Urgency.Blocking = u.Blocking
	}

	// Reports: user entries override, defaults fill the rest.
	maps.Copy(cfg.Reports, user.Reports)
}
