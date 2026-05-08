package main

import (
	"encoding/json"

	"github.com/goccy/go-yaml"
)

func marshalIndentedJSON(value any) ([]byte, error) {
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

func marshalYAMLViaJSON(value any) ([]byte, error) {
	jsonBytes, err := marshalIndentedJSON(value)
	if err != nil {
		return nil, err
	}
	return yaml.JSONToYAML(jsonBytes)
}
