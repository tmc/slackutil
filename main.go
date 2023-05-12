package main

import (
	"context"
	"flag"
	"fmt"
	"os"
)

var (
	flagDCookie  = flag.String("d", "", "'d' cookie value")
	flagDSCookie = flag.String("d-s", "", "'d-s' cookie value")
	flagToken    = flag.String("token", "", "slack token")
)

func main() {
	flag.Parse()
	ctx := context.Background()
	if err := run(ctx, Options{
		DCookie:  *flagDCookie,
		DSCookie: *flagDSCookie,
		Token:    *flagToken,
	}); err != nil {
		fmt.Fprintln(os.Stderr, err)
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
