package main

import (
	"context"
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

func run(ctx context.Context, opts Options) error {
	client, err := newClient(opts.Token, opts.DCookie, opts.DSCookie)
	if err != nil {
		return err
	}
	client.listConversations(ctx)
	return nil
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
	return newClient(t, d, ds)
}
