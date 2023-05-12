package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"

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

func (c *slackClient) listConversations(ctx context.Context) []string {
	var (
		result   []string
		channels []slack.Channel
		err      error
	)
	params := &slack.GetConversationsParameters{
		Types: []string{"im"},
		// Types: []string{"public_channel", "private_channel", "mpim", "im"},
	}
	for err == nil {
		for _, channel := range channels {
			result = append(result, channel.Name)
			j, _ := json.Marshal(channel)
			fmt.Println(string(j))
		}
		channels, params.Cursor, err = c.Client.GetConversations(params)
	}
	return result
}
