package converter

import (
	"encoding/json"
	"fmt"
	"mse/odin"
	"os"
)

// SAVToJSON reads a .sav file and converts it to JSON, writing to outPath.
func SAVToJSON(savPath, jsonPath string) error {
	data, err := os.ReadFile(savPath)
	if err != nil {
		return fmt.Errorf("reading sav file: %w", err)
	}

	reader := odin.NewBinaryDataReader(data)
	tree, err := reader.ReadTree()
	if err != nil {
		return fmt.Errorf("parsing sav: %w", err)
	}

	jsonData, err := json.MarshalIndent(tree, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling json: %w", err)
	}

	if err := os.WriteFile(jsonPath, jsonData, 0644); err != nil {
		return fmt.Errorf("writing json: %w", err)
	}

	fmt.Printf("Converted %s -> %s (%d bytes -> %d bytes)\n", savPath, jsonPath, len(data), len(jsonData))
	return nil
}

// JSONToSAV reads a JSON file and converts it back to .sav binary format.
func JSONToSAV(jsonPath, savPath string) error {
	jsonData, err := os.ReadFile(jsonPath)
	if err != nil {
		return fmt.Errorf("reading json file: %w", err)
	}

	tree := &odin.Node{}
	if err := json.Unmarshal(jsonData, tree); err != nil {
		return fmt.Errorf("parsing json: %w", err)
	}

	writer := odin.NewBinaryDataWriter()
	if err := writer.WriteTree(tree); err != nil {
		return fmt.Errorf("writing binary: %w", err)
	}

	savData := writer.Bytes()
	if err := os.WriteFile(savPath, savData, 0644); err != nil {
		return fmt.Errorf("writing sav: %w", err)
	}

	fmt.Printf("Converted %s -> %s (%d bytes -> %d bytes)\n", jsonPath, savPath, len(jsonData), len(savData))
	return nil
}
