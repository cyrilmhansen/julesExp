package main

import (
	"encoding/json"
	"fmt"
	"os"
	// time package is not needed here anymore as structs are defined in main.go
)

// saveData marshals AccountData to JSON and writes it to filePath
func saveData(data AccountData, filePath string) error {
	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling data: %w", err)
	}
	err = os.WriteFile(filePath, jsonData, 0644)
	if err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}
	return nil
}

// loadData reads JSON data from filePath and unmarshals it into AccountData
func loadData(filePath string) (AccountData, error) {
	var data AccountData
	fileData, err := os.ReadFile(filePath)
	if err != nil {
		return data, fmt.Errorf("error reading file: %w", err)
	}
	err = json.Unmarshal(fileData, &data)
	if err != nil {
		return data, fmt.Errorf("error unmarshaling data: %w", err)
	}
	return data, nil
}
