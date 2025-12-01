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

	// Extract the PIKE Invite URL from the capabilities response
	pikeInviteURL := capabilities.ExtractPikeInviteURL()
	if pikeInviteURL == "" {
		log.Fatalf("PIKE invite capability not found")
	}
	log.Printf("found PIKE Invite URL: %s", pikeInviteURL)

	// Prepare the contact request payload
	request := &paymail.PikeContactRequestPayload{
		FullName: "John Doe",
		Paymail:  "johndoe@example.com",
	}

	// Send the contact request using the invite URL
	var response *paymail.PikeContactRequestResponse
	if response, err = client.AddInviteRequest(pikeInviteURL, "alias", "domain.tld", request); err != nil {
		log.Fatalf("error sending invite request: %s", err.Error())
	}
	log.Printf("invite request response: %v", response)
}
