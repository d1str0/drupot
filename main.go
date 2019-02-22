package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/BurntSushi/toml"
	"github.com/d1str0/hpfeeds"
)

const Version = "v0.0.1"

type Config struct {
	Port                int `toml:"HttpServerPort"`
	DrupalChangelogFile string
	HpfeedsMeta         string
	HpfConfig           *HpfConfig
}

// Config for Hpfeeds publishing
type HpfConfig struct {
	Host    string
	Port    int
	Ident   string
	Auth    string
	Channel string
}

func main() {
	fmt.Println("///- Running Drupot")
	fmt.Println("//- %s", Version)

	// Load config file
	var configFilename string
	var logFilename string
	flag.StringVar(&configFilename, "c", "config.toml", "load given config file")
	flag.StringVar(&logFilename, "l", "", "log file to write to")

	flag.Parse()

	fmt.Printf("//- Loading config file: %s\n", configFilename)
	c := loadConfig(configFilename)
	if c.HpfConfig != nil {
		hpc := c.HpfConfig
		fmt.Printf("/- Connecting to hpfeeds server: %s\n", hpc.Host)
		fmt.Printf("/-\tPort: %d\n", hpc.Port)
		fmt.Printf("/-\tIdent: %s\n", hpc.Ident)
		fmt.Printf("/-\tAuth: %s\n", hpc.Auth)
		fmt.Printf("/-\tChannel: %s\n", hpc.Channel)

		client := hpfeeds.NewClient(hpc.Host, hpc.Port, hpc.Ident, hpc.Auth)
		err := client.Connect()
		if err != nil {
			log.Fatalf("Error connecting to hpfeeds server: %s\n", err.Error())
		}

		publish := make(chan []byte)
		client.Publish(hpc.Channel, publish)
		publish <- []byte("test")
	}

	if logFilename != "" {
		fmt.Printf("//- Logging to file: %s\n", logFilename)
	}

}

func loadConfig(filename string) *Config {
	var c Config
	_, err := toml.DecodeFile(filename, &c)
	if err != nil {
		log.Fatalf("Unable to parse config file: %s\n", err.Error())
	}
	return &c
}
