package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

func main() {
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func newClientFromFlags(cmd *cobra.Command) (*slackClient, error) {
	d, err := cmd.Flags().GetString("d-cookie")
	if err != nil {
		return nil, err
	}
	ds, err := cmd.Flags().GetString("ds-cookie")
	if err != nil {
		return nil, err
	}
	t, err := cmd.Flags().GetString("token")
	if err != nil {
		return nil, err
	}
	verbose, _ := cmd.Flags().GetBool("verbose")
	// TODO: check for empty strings
	// TODO: check for properly escaped d-cookie
	return newClient(Options{Token: t, DCookie: d, DSCookie: ds, Verbose: verbose})
}
