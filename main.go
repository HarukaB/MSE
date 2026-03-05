package main

import (
	"fmt"
	"mse/converter"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "to-json":
		if len(os.Args) < 4 {
			fmt.Fprintf(os.Stderr, "Usage: mse to-json <input.sav> <output.json>\n")
			os.Exit(1)
		}
		if err := converter.SAVToJSON(os.Args[2], os.Args[3]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	case "to-sav":
		if len(os.Args) < 4 {
			fmt.Fprintf(os.Stderr, "Usage: mse to-sav <input.json> <output.sav>\n")
			os.Exit(1)
		}
		if err := converter.JSONToSAV(os.Args[2], os.Args[3]); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintf(os.Stderr, "Muse Dash SAV Serializer\n\n")
	fmt.Fprintf(os.Stderr, "Usage:\n")
	fmt.Fprintf(os.Stderr, "  mse to-json <input.sav> <output.json>   Convert SAV to JSON\n")
	fmt.Fprintf(os.Stderr, "  mse to-sav  <input.json> <output.sav>   Convert JSON to SAV\n")
}
