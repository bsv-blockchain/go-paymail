package main

import (
	"context"
	"log"
	"net"

	"github.com/bsv-blockchain/go-paymail"
)

func main() {
	// Load the client
	client, err := paymail.NewClient()
	if err != nil {
		log.Fatalf("error loading client: %s", err.Error())
	}

	// Get the SRV record
	var srv *net.SRV
	srv, err = client.GetSRVRecord(paymail.DefaultServiceName, paymail.DefaultProtocol, "moneybutton.com")
	if err != nil {
		log.Fatalf("error getting SRV record: %s", err.Error())
	}

	// Found record!
	log.Printf("found SRV record: %v", srv)

	// Validate the record (1 instead of 10, moneybutton deviated from the defaults)
	err = client.ValidateSRVRecord(
		context.Background(), srv, paymail.DefaultPort, 1, paymail.DefaultWeight,
	)
	if err != nil {
		log.Fatalf("failed validating SRV record: %s", err.Error())
	}
}
