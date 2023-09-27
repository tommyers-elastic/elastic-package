// Copyright Elasticsearch B.V. and/or licensed to Elasticsearch B.V. under one
// or more contributor license agreements. Licensed under the Elastic License;
// you may not use this file except in compliance with the Elastic License.

package cmd

import (
	"errors"
	"fmt"

	"github.com/spf13/cobra"
	
	"github.com/elastic/elastic-package/internal/cobraext"
	"github.com/elastic/elastic-package/internal/fixer"
	"github.com/elastic/elastic-package/internal/packages"
	
	"github.com/tommyers-elastic/package-spec/v2/code/go/pkg/validator"
)

const fixLongDescription = `Use this command to fix common issues with package formatting and upgrades.

Issues fixed by this command include:
 - field ...: Additional property ... is not allowed
 - field ... is not normalized as expected: expected array, found ...
`

func setupFixCommand() *cobraext.Command {
	cmd := &cobra.Command{
		Use:   "fix",
		Short: "Fix the package",
		Long:  fixLongDescription,
		Args:  cobra.NoArgs,
		RunE:  fixCommandAction,
		}
		
		cmd.Flags().StringP(cobraext.PackageRootFlagName, cobraext.PackageRootFlagShorthand, "", cobraext.PackageRootFlagDescription)
		cmd.Flags().Bool(cobraext.DryRunFlagName, false, cobraext.DryRunFlagDescription)

	return cobraext.NewCommand(cmd, cobraext.ContextPackage)
}

func fixCommandAction(cmd *cobra.Command, args []string) error {
	cmd.Println("Fixing the package..")
	
	packageRoot, err := cmd.Flags().GetString(cobraext.PackageRootFlagName)
	if err != nil {
		return cobraext.FlagParsingError(err, cobraext.PackageRootFlagName)
	}

	if packageRoot == "" {
		var found bool
		packageRoot, found, err = packages.FindPackageRoot()
		if err != nil {
			return fmt.Errorf("locating package root failed: %w", err)
		}

		if !found {
			return errors.New("package root not found")
		}
	}
	
	dryRun, err := cmd.Flags().GetBool(cobraext.DryRunFlagName)
	if err != nil {
		return cobraext.FlagParsingError(err, cobraext.DryRunFlagName)
	}
	
	fixers := []fixer.Fixer{
		fixer.NewLicenseFixer(),
		fixer.NewReleaseFixer(),
	}

	err = validator.ValidateFromPath(packageRoot)
	if err != nil {
		validatorErrors := err.(validator.ValidationErrors)
		for _, e := range validatorErrors {
			fmt.Printf("\nAttempting to fix error: %v\n", e)
			detected := false
			
			for _, f := range fixers {
				if f.Detect(e) {
					fmt.Printf("🔨 Fixing using %s fixer\n", f.Type())
					detected = true
					
					if dryRun {
						fmt.Println("⏭️  Skipping fix in dry-run mode")
					} else if err := f.Fix(); err != nil {
						fmt.Printf("⛔️ Could not fix due to error: %v\n", err)
					} else {
						fmt.Println("✅ Fixed!")
					}
					break
				}
			}
			
			if !detected {
				fmt.Println("😭 No fixer found for this error; could not fix")			
			}
		}
	}

	return nil
}

