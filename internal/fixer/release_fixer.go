// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package fixer

import (
	"regexp"
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
	return nil
}