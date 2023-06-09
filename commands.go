package main

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"
)

func init() {
	RootCmd.PersistentFlags().BoolP("verbose", "v", false, "verbose mode")
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
	RootCmd.AddCommand(AutoRespondCmd)

	DumpConversationCmd.Flags().IntP("limit", "l", 0, "limit number of messages to dump")
	RootCmd.PersistentFlags().BoolP("streaming", "S", false, "streaming mode")
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
	Short: "List conversations",
	Long:  "List conversations",
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
	Short: "Dump conversation history",
	Long:  "Dump conversation history",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		streaming, _ := cmd.Flags().GetBool("streaming")
		ctx := context.Background()
		limit, _ := cmd.Flags().GetInt("limit")
		c, err := newClientFromFlags(cmd)
		if err != nil {
			return err
		}
		hist, err := c.dumpConversation(ctx, args[0], streaming, limit)
		if err != nil {
			return err
		}
		if streaming {
			return nil
		}

		for _, m := range hist {
			j, _ := json.Marshal(m)
			fmt.Println(string(j))
		}
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
			return fmt.Errorf("error creating client: %w", err)
		}
		users, err := c.listUsers(ctx)
		if err != nil {
			return fmt.Errorf("error listing users: %w", err)
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
		return c.sendMessage(ctx, args[0], "", args[1])
	},
}

var AutoRespondCmd = &cobra.Command{
	Use:   "auto-respond",
	Short: "auto-respond",
	Long:  `auto-respond`,
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx := context.Background()
		c, err := newClientFromFlags(cmd)
		if err != nil {
			return err
		}
		return c.autoRespond(ctx)
	},
}
