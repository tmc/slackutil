package main

import (
	"context"
	"encoding/json"
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

func init() {
	RootCmd.PersistentFlags().StringP("d-cookie", "d", "", "'d' cookie value")
	RootCmd.PersistentFlags().StringP("ds-cookie", "s", "", "'d-s' cookie value")
	RootCmd.PersistentFlags().StringP("token", "t", "", "slack token (see readme)")
	RootCmd.MarkPersistentFlagRequired("d-cookie")
	RootCmd.MarkPersistentFlagRequired("ds-cookie")
	RootCmd.MarkPersistentFlagRequired("token")

	RootCmd.AddCommand(ListConversationsCmd)
	RootCmd.AddCommand(ListUsersCmd)
	RootCmd.AddCommand(ListDMsCmd)
	RootCmd.AddCommand(DumpConversationCmd)
	RootCmd.AddCommand(SendMessageCmd)
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

var RootCmd = &cobra.Command{
	Use:   "slackdump",
	Short: "slackdump",
	Long:  `slackdump`,
	RunE: func(cmd *cobra.Command, args []string) error {
		cmd.Usage()
		return nil
	},
}

var ListConversationsCmd = &cobra.Command{
	Use:   "list-conversations",
	Short: "list-conversations",
	Long:  `list-conversations`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := newClientFromFlags(cmd)
		if err != nil {
			return err
		}
		c.listConversations(ctx)
		return nil
	},
}

var ListDMsCmd = &cobra.Command{
	Use:   "list-dms",
	Short: "list-dms",
	Long:  `list-dms`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := newClientFromFlags(cmd)
		if err != nil {
			return err
		}
		c.listConversations(ctx, "im")
		return nil
	},
}

var DumpConversationCmd = &cobra.Command{
	Use:   "dump-conversation",
	Short: "dump-conversation",
	Long:  `dump-conversation`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := newClientFromFlags(cmd)
		if err != nil {
			return err
		}
		c.dumpConversation(ctx, args[0])
		return nil
	},
}

var ListUsersCmd = &cobra.Command{
	Use:   "list-users",
	Short: "list-users",
	Long:  `list-users`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := newClientFromFlags(cmd)
		if err != nil {
			return err
		}
		users, err := c.listUsers(ctx)
		if err != nil {
			return err
		}
		for _, u := range users {
			j, _ := json.Marshal(u)
			fmt.Println(string(j))
		}
		return nil
	},
}

var SendMessageCmd = &cobra.Command{
	Use:   "send-msg",
	Short: "send-msg",
	Long:  `send-msg`,
	Args:  cobra.ExactArgs(2),
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := newClientFromFlags(cmd)
		if err != nil {
			return err
		}
		return c.sendMessage(ctx, args[0], args[1])
	},
}
