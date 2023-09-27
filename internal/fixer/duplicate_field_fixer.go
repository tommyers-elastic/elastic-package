// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package fixer

import (
	"fmt"
	"io"
	"regexp"
	"os"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/elastic/elastic-package/internal/fields"
)

type DuplicateFieldFixer struct {
	re         *regexp.Regexp
	field      string
	filePath   string
	fixedFiles []string
}

func NewDuplicateFieldFixer() *DuplicateFieldFixer {
	return &DuplicateFieldFixer{
		re: regexp.MustCompile(`field "(.*?)" is defined multiple times for data stream ".+?", found in: (.*?\.ya?ml)`),
	}
}

func (lf *DuplicateFieldFixer) Type() string {
	return "duplicate field"
}

func (lf *DuplicateFieldFixer) Detect(e error) bool {
	match := lf.re.FindStringSubmatch(e.Error())
	if len(match) == 3 {
		lf.field = match[1]
		lf.filePath = match[2]
		return true
	}

	return false
}

func (lf *DuplicateFieldFixer) Fix() error {
	fileToFix := lf.filePath
	for _, fixedFile := range lf.fixedFiles {
		if fileToFix == fixedFile {
			// we have already deduplicated this file
			return nil
		}
	}

	r, err := os.Open(fileToFix)
	if err != nil {
		return err
	}

	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	var n yaml.Node
	if err := yaml.Unmarshal(b, &n); err != nil {
		return fmt.Errorf("could not parse fields: %w", err)
	}

	fd := fields.FieldDefinitions{}
	if err := fd.UnmarshalYAML(n.Content[0]); err != nil {
		return fmt.Errorf("could not parse fields: %w", err)
	}

	deduplicatedFields := deduplicate(fd, "", []string{})

	newFields, err := yaml.Marshal(deduplicatedFields)
	if err != nil {
		return fmt.Errorf("could not marshal fields to YAML: %w", err)
	}
	
	if err := overwriteFile(fileToFix, newFields); err != nil {
		return fmt.Errorf("error fixing file (%s): %w", fileToFix, err)
	}

	lf.fixedFiles = append(lf.fixedFiles, fileToFix)
	return nil
}

// deduplicate returns a copy of the input fields with duplicates removed based on their fully qualified field name
func deduplicate(inputFields fields.FieldDefinitions, fieldPrefix string, fullyQualifiedNames []string) fields.FieldDefinitions {
	var deduplicatedFields fields.FieldDefinitions

	for _, f := range inputFields {
		fullyQualifiedName := fmt.Sprintf("%s.%s", fieldPrefix, f.Name)

		add := true
		for _, name := range fullyQualifiedNames {
			if name == fullyQualifiedName {
				// field already added, do not add again
				add = false
				break
			}
		}
		
		if add {
			if f.Type == "group" {
				f.Fields = deduplicate(f.Fields, fullyQualifiedName, fullyQualifiedNames)
			}
			deduplicatedFields = append(deduplicatedFields, f)
			fullyQualifiedNames = append(fullyQualifiedNames, fullyQualifiedName)
		}
	}

	return deduplicatedFields
}
