// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package fixer

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

type ReleaseFixer struct {
	re *regexp.Regexp
	filePath string
}

func NewReleaseFixer() *ReleaseFixer {
	return &ReleaseFixer{
		re: regexp.MustCompile(`file "(.*?)" is invalid: field .+?: Additional property release is not allowed`),
	}
}

func (lf *ReleaseFixer) Type() string {
	return "release"
}

func (lf *ReleaseFixer) Detect(e error) bool {	
	match := lf.re.FindStringSubmatch(e.Error())
	if len(match) == 2 {
		lf.filePath = match[1]
		return true
	}

	return false
}

func (lf *ReleaseFixer) Fix() error {
	// if the release field is in a package-level manifest file, remove it and validate the package version
	if strings.HasSuffix(lf.filePath, "manifest.yaml") || strings.HasSuffix(lf.filePath, "manifest.yml") {
		manifest, err := parseManifest(lf.filePath)
		if err != nil {
			return fmt.Errorf("error parsing manifest: %w", err)
		}

		// if the release field is not "ga", the package version should reflect that
		if strings.ToLower(manifest["release"].(string)) != "ga" {
			version := manifest["version"].(string)
			hasPreReleaseTag := strings.Contains(version, "-")
			hasPreReleaseVersion := strings.Split(version, ".")[0] == "0"
			if !(hasPreReleaseTag || hasPreReleaseVersion) {
				return fmt.Errorf("package version indicates GA release, but release field is not GA")
			}
		}

		delete(manifest, "release")

		newManifest, err := yaml.Marshal(manifest)
		if err != nil {
			return fmt.Errorf("could not marshal manifest to YAML: %w", err)
		}

		if err := overwriteFile(lf.filePath, newManifest); err != nil {
			return fmt.Errorf("error fixing file (%s): %w", lf.filePath, err)
		}
	} else {
		// assuming the release field is in a field mapping group
		// we could parse the fields and remove the field etc.
		// but it's way simpler here to just remove the offending line

		f, err := os.Open(lf.filePath)
		if err != nil {
			return err
		}

		var lines []string
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}

		f.Close()

		if scanner.Err() != nil {
			return err
		}

		f, err = os.Create(lf.filePath)
		if err != nil {
			return err
		}
		defer f.Close()

		w := bufio.NewWriter(f)
		for _, line := range lines {
			if !strings.Contains(line, "release: ") {
				_, err := fmt.Fprintln(w, line)
				if err != nil {
					return err
				}
			}
		}

		return w.Flush()
	}

	return nil
}