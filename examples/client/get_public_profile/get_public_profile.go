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
	// This is required first to get the corresponding PublicProfile endpoint url
	var capabilities *paymail.CapabilitiesResponse
	if capabilities, err = client.GetCapabilities("moneybutton.com", paymail.DefaultPort); err != nil {
		log.Fatalf("error getting capabilities: %s", err.Error())
	}
	log.Printf("found capabilities: %d", len(capabilities.Capabilities))

	// Extract the PublicProfile URL from the capabilities response
	publicProfileURL := capabilities.GetString(paymail.BRFCPublicProfile, "")

	// Get the public profile
	var profile *paymail.PublicProfileResponse
	if profile, err = client.GetPublicProfile(publicProfileURL, "mrz", "moneybutton.com"); err != nil {
		log.Fatalf("error getting profile: %s", err.Error())
	}
	log.Printf("found profile: %s : %s", profile.Name, profile.Avatar)
}
