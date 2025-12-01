package main

import (
	"log"

	"github.com/bsv-blockchain/go-paymail"
)

func main() {
	// Load additional specification(s)
	additionalSpec := `[{"author": "andy (nChain)","id": "57dd1f54fc67","title": "BRFC Specifications","url": "http://bsvalias.org/01-02-brfc-id-assignment.html","version": "1"}]`
	_, err := paymail.NewClient(paymail.WithUserAgent(additionalSpec))
	if err != nil {
		log.Fatalf("error loading client: %s", err.Error())
	}

}
