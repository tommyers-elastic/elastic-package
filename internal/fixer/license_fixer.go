// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package fixer

import (
	"fmt"
	"io"
	"regexp"
	"os"
	
	"gopkg.in/yaml.v3"
)

type LicenceFixer struct {
	re *regexp.Regexp
	manifestFilePath string
}

func NewLicenseFixer() *LicenceFixer {
	return &LicenceFixer{
		re: regexp.MustCompile(`file "(.*?/manifest\.ya?ml)" is invalid: field .+?: Additional property license is not allowed`),
	}
}

func (lf *LicenceFixer) Type() string {
	return "license"
}

func (lf *LicenceFixer) Detect(e error) bool {	
	match := lf.re.FindStringSubmatch(e.Error())
	if len(match) == 2 {
		lf.manifestFilePath = match[1]
		return true
	}
	
	return false
}

func (lf *LicenceFixer) Fix() error {
	r, err := os.Open(lf.manifestFilePath)
	if err != nil {
		return err
	}

	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	var manifest map[string]interface{}
	if err := yaml.Unmarshal(b, &manifest); err != nil {
		return fmt.Errorf("could not parse manifest: %w", err)
	}
	
	licenseValue := manifest["license"].(string)
	delete(manifest, "license")
	
	conditions := manifest["conditions"].(map[string]interface{})
	conditions["elastic"] = map[string]string{"subscription": licenseValue}
	
	newManifest, err := yaml.Marshal(manifest)
	if err != nil {
		return fmt.Errorf("could not marshal manifest to YAML: %w", err)
	}
	
	if err := overwriteFile(lf.manifestFilePath, newManifest); err != nil {
		return fmt.Errorf("error fixing file (%s): %w", lf.manifestFilePath, err)
	}
	return nil
}