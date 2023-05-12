package main

import (
	"context"
	"encoding/json"
	"fmt"
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
	return &slackClient{slack.New(token, slack.OptionHTTPClient(client))}, nil
}

type slackClient struct {
	*slack.Client
}

func (s *slackClient) listConversations(ctx context.Context) []string {
	var (
		result   []string
		channels []slack.Channel
		err      error
	)
	params := &slack.GetConversationsParameters{
		Types: []string{"public_channel", "private_channel", "mpim", "im"},
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

func (s *slackClient) dumpConversation(ctx context.Context, conversationID string) []string {
	var (
		result   []string
		channels []slack.Channel
	)
	params := &slack.GetConversationHistoryParameters{
		ChannelID:          conversationID,
		IncludeAllMetadata: true,
	}
	hasMore := true
	for hasMore {
		for _, channel := range channels {
			result = append(result, channel.Name)
			j, _ := json.Marshal(channel)
			fmt.Println(string(j))
		}
		hist, err := s.Client.GetConversationHistoryContext(ctx, params)
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			return result
		}
		params.Cursor = hist.ResponseMetaData.NextCursor
	}
	return result
}

func (s *slackClient) listUsers(ctx context.Context) ([]slack.User, error) {
	return s.Client.GetUsersContext(ctx)
}
