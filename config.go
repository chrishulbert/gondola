package main

import (
	"errors"
	"github.com/BurntSushi/toml"
	"os"
	"os/user"
	"strings"
)

type Config struct {
	Root         string
	DebugSkipHLS bool // Skip conversion, this is good for speeding up dev/debugging.
}

func loadConfig() (Config, error) {
	// Find home.
	configFile := expandTilde("~/.gondola")

	// Check it exists.
	if _, err := os.Stat(configFile); os.IsNotExist(err) {
		return Config{}, errors.New("Your config file is missing: " + configFile)
	}

	// Parse it.
	var conf Config
	if _, err := toml.DecodeFile(configFile, &conf); err != nil {
		return Config{}, err
	}

	// Validate it.
	if conf.Root == "" {
		return Config{}, errors.New("'root' is missing from your config file. It should point to a root folder where your media is to be stored.")
	}

	return conf, nil
}

func expandTilde(path string) string {
	usr, _ := user.Current()
	homeDir := usr.HomeDir
	return strings.Replace(path, "~", homeDir, -1)
}
