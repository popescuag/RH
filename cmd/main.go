package main

import (
	"log"
	"os"
	"time"

	"github.com/popescuag/RH/internal/pkg/delta"
	"github.com/popescuag/RH/internal/pkg/signature"
	"github.com/popescuag/RH/internal/pkg/validator"
)

func main() {
	err := validator.ValidateInputParams(os.Args[1:])

	if err != nil {
		log.Printf("Invalid command parameters. Error was %v", err)
		os.Exit(1)
	}

	startTime := time.Now()
	if os.Args[1] == validator.SIGNATURE_CMD {
		err = signature.Compute(os.Args[2], os.Args[3])
	} else {
		err = delta.Compute(os.Args[2], os.Args[3], os.Args[4])
	}

	if err != nil {
		log.Printf("Command has failed. Error was %v", err)
		os.Exit(1)
	}

	log.Printf("Command completed in %v", time.Since(startTime))
	os.Exit(0)
}
