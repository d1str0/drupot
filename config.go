package main

import (
	"log"

	"github.com/BurntSushi/toml"
)

type AppConfig struct {
	Drupal   *DrupalConfig
	Hpfeeds  *HpfeedsConfig
	PublicIP *PublicIPConfig `toml:"fetch_public_ip"`
}

// DrupalConfig provides configuration for how to host the Drupal web app
// portion of the honeypot.
// [drupal]
type DrupalConfig struct {
	Port              int
	ChangelogFilepath string `toml:"changelog_filepath"`
}

// HpfeedsConfig provides configuration for connecting to an hpfeeds broker
// server and credentials for publishing data.
// [hpfeeds]
type HpfeedsConfig struct {
	Enabled bool
	Host    string
	Port    int
	Ident   string
	Auth    string
	Channel string
	Meta    string
}

// [fetch_public_ip]
type PublicIPConfig struct {
	Enabled bool
	URLs    []string
}

func loadConfig(filename string) *AppConfig {
	var c AppConfig
	_, err := toml.DecodeFile(filename, &c)
	if err != nil {
		log.Fatalf("Unable to parse config file: %s\n", err.Error())
	}
	return &c
}
