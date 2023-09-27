// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package fixer

import (
	"fmt"
	"os"
)

func overwriteFile(fileToFix string, data []byte) error {
	f, err := os.Create(fileToFix)
	if err != nil {
		return fmt.Errorf("could not create file: %w", err)
	}

	if _, err = f.Write(data); err != nil {
		return fmt.Errorf("could not write data: %w", err)
	}

	return nil
}