package main

import (
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/d1str0/hpfeeds"
	"github.com/google/uuid"
)

const Version = "v0.0.7"

type App struct {
	Publish    chan []byte
	SeenIPLock sync.RWMutex
	SeenIP     map[string]bool
	SensorIP   string
	Config     *AppConfig
	SensorUUID *uuid.UUID
}

func main() {
	fmt.Println("///- Running Drupot")
	fmt.Printf("///- %s\n", Version)

	// All we take is a config file argument.
	var configFilename string
	flag.StringVar(&configFilename, "c", "config.toml", "load given config file")
	flag.Parse()

	fmt.Printf("//- Loading config file: %s\n", configFilename)
	c := loadConfig(configFilename)

	var app App
	app.SensorIP = "127.0.0.1" // Default will be overwritten if public IP set to fetch.
	app.Config = c
	app.SeenIP = make(map[string]bool)
	app.Publish = make(chan []byte)

	// TODO: Persist UUID. Maybe a command line flag to refresh or overwrite.
	uuid, err := uuid.NewUUID()
	if err != nil {
		log.Fatalf("Error generating UUID: %s\n", err.Error())
	}
	app.SensorUUID = &uuid

	if c.Hpfeeds.Enabled {

		hpc := c.Hpfeeds
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

		client.Publish(hpc.Channel, app.Publish)
	}

	if c.PublicIP.Enabled {
		ip, err := getPublicIP(c.PublicIP)
		if err != nil {
			log.Fatalf("Error getting public IP: %s\n", err.Error())
		}
		app.SensorIP = ip
	}

	// Load routes for the server
	mux := routes(app)

	addr := fmt.Sprintf(":%d", c.Drupal.Port)
	s := http.Server{
		Addr:         addr,
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		IdleTimeout:  120 * time.Second,
	}
	log.Fatal(s.ListenAndServe())

}

// getPublicIP goes through a list of URLs to
func getPublicIP(c *PublicIPConfig) (string, error) {
	var ip net.IP
	for _, site := range c.URLs {
		resp, err := http.Get(site)
		if err != nil {
			log.Print(err)
			continue
		}
		defer resp.Body.Close()
		var body []byte
		body, err = ioutil.ReadAll(resp.Body)
		if err != nil {
			return "", err
		}

		// Try and parse response into a valid IP.
		ip = net.ParseIP(string(body))
		// ip will be nil if the parsing fails.
		// if not nil, succesfully got back an IP and parsed it
		if ip != nil {
			return ip.String(), nil
		}
	}
	return "", errors.New("Unable to get public IP")
}
