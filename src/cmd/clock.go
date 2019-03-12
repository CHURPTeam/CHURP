package main

import "flag"
import "github.com/bl4ck5un/ChuRP/src/networking/clock"

func main() {
	metadataPath := flag.String("path", "/mpss/metadata", "Enter the metadata path")
	flag.Parse()

	clock, _ := clock.New(*metadataPath)
	clock.Connect()
	clock.ClientStartEpoch()
}
