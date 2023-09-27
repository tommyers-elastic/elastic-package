// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package fixer

import (
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v3"
)

func overwriteFile(fileToFix string, data []byte) error {
	f, err := os.Create(fileToFix)
	if err != nil {
		return fmt.Errorf("could not create file: %w", err)
	}
	defer f.Close()

	if _, err = f.Write(data); err != nil {
		return fmt.Errorf("could not write data: %w", err)
	}

	return nil
}

func parseManifest(manifestFile string) (map[string]interface{}, error) {
	f, err := os.Open(manifestFile)
	if err != nil {
		return nil, fmt.Errorf("could not open file (%s): %w", manifestFile, err)
	}
	defer f.Close()

	b, err := io.ReadAll(f)
	if err != nil {
		return nil, fmt.Errorf("could not read file: %w", err)
	}

	var manifest map[string]interface{}
	if err := yaml.Unmarshal(b, &manifest); err != nil {
		return nil, fmt.Errorf("could not unmarshal YAML: %w", err)
	}

	return manifest, nil
}