package main

import "flag"
import "github.com/bl4ck5un/ChuRP/src/networking/nodes"

func main() {
	label := flag.Int("l", 1, "Enter node label")
	counter := flag.Int("c", 1, "Enter number of nodes")
	degree := flag.Int("d", 1, "Enter the polynomial degree")
	metadataPath := flag.String("path", "/mpss/metadata", "Enter the metadata path")
	aws := flag.Bool("aws", false, "if test on real aws")
	flag.Parse()

	n, _ := nodes.New(*degree, *label, *counter, *metadataPath)
	n.Serve(*aws)
}
