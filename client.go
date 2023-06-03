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

	"github.com/redis/go-redis/v9"

	"github.com/slack-go/slack"
)

type Options struct {
	Token    string
	DCookie  string
	DSCookie string

	Verbose bool
}

func newClient(opts Options) (*slackClient, error) {
	if opts.DCookie == "" {
		return nil, fmt.Errorf("d cookie flag is required")
	}
	jar, _ := cookiejar.New(nil)
	url, _ := url.Parse("https://slack.com")
	jar.SetCookies(url, []*http.Cookie{
		{
			Name:   "d",
			Value:  opts.DCookie,
			Path:   "/",
			Domain: "slack.com",
		},
	})

	if opts.DSCookie != "" {
		jar.SetCookies(url, []*http.Cookie{
			{
				Name:   "d-s",
				Value:  opts.DSCookie,
				Path:   "/",
				Domain: "slack.com",
			},
		})
	}
	client := &http.Client{
		Jar: jar,
	}
	ctx := context.Background()
	sc := slack.New(opts.Token,
		slack.OptionHTTPClient(client),
		slack.OptionDebug(opts.Verbose),
		slack.OptionLog(log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)))

	rc := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "", // no password set
		DB:       0,  // use default DB
	})
	if err := rc.Ping(ctx).Err(); err != nil {
		panic(err)
	}
	return &slackClient{
		Client:              sc,
		redisClient:         rc,
		conversationHistory: make(map[string][]slack.Message),
	}, nil
}

type slackClient struct {
	*slack.Client

	conversationHistory map[string][]slack.Message

	redisClient *redis.Client
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
func (s *slackClient) dumpConversation(ctx context.Context, conversationID string, stream bool, limit int) ([]slack.Message, error) {
	result := []slack.Message{}
	params := &slack.GetConversationHistoryParameters{
		ChannelID:          conversationID,
		IncludeAllMetadata: true,
	}

	hasMore := true
	i := 0
	for hasMore {
		hist, err := s.Client.GetConversationHistoryContext(ctx, params)
		if err != nil {
			return nil, err
		}
		// Loop that iterates over the messages
		for _, m := range hist.Messages {
			if stream {
				j, _ := json.Marshal(m)
				fmt.Println(string(j))
			}
			result = append(result, m)
			i++
			if limit > 0 && i >= limit {
				return result, nil
			}
		}
		params.Cursor = hist.ResponseMetaData.NextCursor
		hasMore = hist.HasMore
	}
	return result, nil
}

func (s *slackClient) listUsers(ctx context.Context) ([]slack.User, error) {
	return s.Client.GetUsersContext(ctx)
}

func (s *slackClient) sendMessage(ctx context.Context, channelID, threadTS, message string) error {
	opts := []slack.MsgOption{
		slack.MsgOptionText(message, false),
	}
	if threadTS != "" {
		opts = append(opts, slack.MsgOptionTS(threadTS))
	}
	a, b, err := s.Client.PostMessageContext(ctx, channelID, opts...)
	fmt.Println(a, b, err)
	return err
}
