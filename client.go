package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"

	"github.com/slack-go/slack"
)

type Options struct {
	Token    string
	DCookie  string
	DSCookie string
}

func newClient(token, dCookie, dsCookie string) (*slackClient, error) {
	if dCookie == "" {
		return nil, fmt.Errorf("d cookie flag is required")
	}
	jar, _ := cookiejar.New(nil)
	url, _ := url.Parse("https://slack.com")
	jar.SetCookies(url, []*http.Cookie{
		{
			Name:   "d",
			Value:  dCookie,
			Path:   "/",
			Domain: "slack.com",
		},
	})

	if dsCookie != "" {
		jar.SetCookies(url, []*http.Cookie{
			{
				Name:   "d-s",
				Value:  dsCookie,
				Path:   "/",
				Domain: "slack.com",
			},
		})
	}
	client := &http.Client{
		Jar: jar,
	}
	sc := slack.New(token,
		slack.OptionHTTPClient(client),
		slack.OptionDebug(true),
		slack.OptionLog(log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)))
	return &slackClient{sc}, nil
}

type slackClient struct {
	*slack.Client
}

func (s *slackClient) listConversations(ctx context.Context, types ...string) []string {
	var (
		result   []string
		channels []slack.Channel
		err      error
	)
	if len(types) == 0 {
		types = []string{"public_channel", "private_channel", "mpim", "im"}
	}
	params := &slack.GetConversationsParameters{
		Types: types,
	}
	for err == nil {
		for _, channel := range channels {
			result = append(result, channel.Name)
			j, _ := json.Marshal(channel)
			fmt.Println(string(j))
		}
		channels, params.Cursor, err = s.Client.GetConversations(params)
	}
	return result
}

// Dumps an entire conversation to stdout
func (s *slackClient) dumpConversation(ctx context.Context, conversationID string, limit int) error {
	params := &slack.GetConversationHistoryParameters{
		ChannelID:          conversationID,
		IncludeAllMetadata: true,
	}

	hasMore := true
	i := 0
	for hasMore {
		hist, err := s.Client.GetConversationHistoryContext(ctx, params)
		if err != nil {
			return err
		}
		// Loop that iterates over the messages
		for _, m := range hist.Messages {
			j, _ := json.Marshal(m)
			fmt.Println(string(j))
			i++
			if limit > 0 && i >= limit {
				return nil
			}
		}
		params.Cursor = hist.ResponseMetaData.NextCursor
		hasMore = hist.HasMore
	}
	return nil
}

func (s *slackClient) listUsers(ctx context.Context) ([]slack.User, error) {
	return s.Client.GetUsersContext(ctx)
}

func (s *slackClient) sendMessage(ctx context.Context, channelID, message string) error {
	a, b, err := s.Client.PostMessageContext(ctx, channelID, slack.MsgOptionText(message, false))
	fmt.Println(a, b, err)
	return err
}
