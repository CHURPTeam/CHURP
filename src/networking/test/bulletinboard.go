package main

import (
	"../bulletinboard"
	"flag"
)

func main() {
	cnt := flag.Int("c", 2, "Enter number of nodes")
	degree := flag.Int("d", 1, "Enter the polynomial degree")
	metadataPath := flag.String("path", "/mpss/metadata", "Enter the metadata path")
	aws := flag.Bool("aws", false, "if test on real aws")
	flag.Parse()

	bb, _ := bulletinboard.New(*degree, *cnt, *metadataPath)
	bb.Serve(*aws)
}
