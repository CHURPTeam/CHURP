package main

import (
	"../clock"
	"flag"
)

func main() {
	metadataPath := flag.String("path", "/mpss/metadata", "Enter the metadata path")
	flag.Parse()

	clock, _ := clock.New(*metadataPath)
	clock.Connect()
	clock.ClientStartEpoch()
}
