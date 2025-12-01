package main

import (
	"log"

	"github.com/bsv-blockchain/go-paymail"
)

func main() {
	// Load the client
	client, err := paymail.NewClient()
	if err != nil {
		log.Fatalf("error loading client: %s", err.Error())
	}

	// Get the capabilities
	var capabilities *paymail.CapabilitiesResponse
	if capabilities, err = client.GetCapabilities("moneybutton.com", paymail.DefaultPort); err != nil {
		log.Fatalf("error getting capabilities: %s", err.Error())
	}
	log.Printf("found capabilities: %d", len(capabilities.Capabilities))

	// Check if capabilities exist
	found := capabilities.Has(paymail.BRFCPki, paymail.BRFCPkiAlternate)
	log.Printf("capabilities found: %v", found)
}
