package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/helper/logging"
	api "github.com/waisbrot/tp-link-api/lib"
)

const invalidCreds = `
No valid credentials found for TP-Link provider.
`

// Client returns a new TP-Link client
func Client(username, password, host string) (*api.Client, error) {
	c = api.Config{
		Host:     host,
		Username: username,
		Password: password,
	}
	if c.Username == "" || c.Password == "" || c.Host == "" {
		return nil, fmt.Errorf(invalidCreds)
	}

	config := &Config{
		HTTPClient: httpClient,
		Username:   c.Username,
		Password:   c.Password,
		Host:       c.Host,
	}

	client, err := api.NewClient(config)
	if err != nil {
		return nil, err
	}

	log.Printf("[INFO] TPLink client configured")

	return client, nil
}
