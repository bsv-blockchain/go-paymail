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
	if capabilities, err = client.GetCapabilities("example.com", paymail.DefaultPort); err != nil {
		log.Fatalf("error getting capabilities: %s", err.Error())
	}
	log.Printf("found capabilities: %d", len(capabilities.Capabilities))

	// Extract the PIKE Outputs URL from the capabilities response
	pikeOutputsURL := capabilities.ExtractPikeOutputsURL()
	if pikeOutputsURL == "" {
		log.Fatalf("PIKE outputs capability not found")
	}
	log.Printf("found PIKE Outputs URL: %s", pikeOutputsURL)

	// Prepare the payload
	alias := "examplealias"
	domain := "example.com"
	payload := &paymail.PikePaymentOutputsPayload{
		SenderPaymail: "joedoe@example.com",
		Amount:        1000, // Example amount in satoshis
	}

	// Get the outputs template from PIKE
	var outputs *paymail.PikePaymentOutputsResponse
	if outputs, err = client.GetOutputsTemplate(pikeOutputsURL, alias, domain, payload); err != nil {
		log.Fatalf("error getting outputs template: %s", err.Error())
	}
	log.Printf("found outputs template: %v", outputs)
}
