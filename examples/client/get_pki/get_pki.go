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
	// This is required first to get the corresponding PKI endpoint url
	var capabilities *paymail.CapabilitiesResponse
	if capabilities, err = client.GetCapabilities("moneybutton.com", paymail.DefaultPort); err != nil {
		log.Fatalf("error getting capabilities: %s", err.Error())
	}
	log.Printf("found capabilities: %d", len(capabilities.Capabilities))

	// Extract the PKI URL from the capabilities response
	pkiURL := capabilities.GetString(paymail.BRFCPki, paymail.BRFCPkiAlternate)

	// Get the actual PKI
	var pki *paymail.PKIResponse
	if pki, err = client.GetPKI(pkiURL, "mrz", "moneybutton.com"); err != nil {
		log.Fatalf("error getting pki: %s", err.Error())
	}
	log.Printf("found pki: %v", pki)
}
