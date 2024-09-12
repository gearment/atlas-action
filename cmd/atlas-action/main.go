// Copyright 2021-present The Atlas Authors. All rights reserved.
// This source code is licensed under the Apache 2.0 license found
// in the LICENSE file in the root directory of this source tree.

package main

import (
	"context"
	"errors"
	"fmt"
	"os"

	"ariga.io/atlas-action/atlasaction"
	"ariga.io/atlas-go-sdk/atlasexec"
	"github.com/alecthomas/kong"
)

var (
	// version holds atlas-action version. When built with cloud packages should be set by build flag, e.g.
	// "-X 'main.version=v0.1.2'"
	version string
	// commit holds the git commit hash. When built with cloud packages should be set by build flag, e.g.
	// "-X 'main.commit=abcdef1234'"
	commit string = "dev"
)

func main() {
	act, err := atlasaction.New(os.Getenv, os.Stdout)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to run action in the current environment: %s\n", err)
		os.Exit(1)
	}
	act.Atlas, err = atlasexec.NewClient("", "atlas")
	if err != nil {
		act.Fatalf("Failed to create client: %s", err)
	}
	act.Version = version
	cli := kong.Parse(
		&RunActionCmd{},
		kong.BindTo(context.Background(), (*context.Context)(nil)),
	)
	if err := cli.Run(act); err != nil {
		if uerr := errors.Unwrap(err); uerr != nil {
			err = uerr
		}
		act.Fatalf(err.Error())
	}
}

// VersionFlag is a flag type that can be used to display a version number, stored in the "version" variable.
type VersionFlag bool

// BeforeReset writes the version variable and terminates with a 0 exit status.
func (v VersionFlag) BeforeReset(app *kong.Kong) error {
	_, err := fmt.Fprintf(app.Stdout, "%s-%s\n", version, commit)
	app.Exit(0)
	return err
}

// RunActionCmd is a command to run one of the Atlas GitHub Actions.
type RunActionCmd struct {
	Action  string      `help:"Command to run" required:""`
	Version VersionFlag `help:"Prints the version and exits"`
}

func (r *RunActionCmd) Run(ctx context.Context, a *atlasaction.Actions) error {
	_ = os.Setenv("ATLAS_ACTION_COMMAND", r.Action)
	defer func() {
		_ = os.Unsetenv("ATLAS_ACTION_COMMAND")
	}()
	return a.Run(ctx, r.Action)
}
