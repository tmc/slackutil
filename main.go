package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"

	"github.com/slack-go/slack"
)

var (
	flagDCookie  = flag.String("d", "", "'d' cookie value")
	flagDSCookie = flag.String("d-s", "", "'d-s' cookie value")
	flagToken    = flag.String("token", "", "slack token")
)

func main() {
	flag.Parse()
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
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

func run() error {
	client, err := newClient(*flagToken, *flagDCookie, *flagDSCookie)
	if err != nil {
		return err
	}
	ctx := context.Background()
	conversations := client.listConversations(ctx)
	_ = conversations
	// for _, c := range conversations {
	// 	fmt.Printf("%s\n", c)
	// }
	return nil
}

func (c *slackClient) listConversations(ctx context.Context) []string {
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
			fmt.Println(channel.Name)
		}
		channels, params.Cursor, err = c.Client.GetConversations(params)
	}
	return result
}

func getSlackToken(cookie string) (string, error) {
	//workspaces := []*Workspace{}
	fmt.Printf("cookie: '%v'", cookie)
	unesc, err := url.QueryUnescape(cookie)
	if err != nil {
		return "", err
	}
	fmt.Printf("cookie: '%v'", unesc)
	cookie = url.QueryEscape(unesc)
	fmt.Printf("cookie: '%v'", cookie)
	req, err := http.NewRequest("GET", "https://xxx-foo.slack.com", nil)
	// set
	req.Header.Set("Cookie", fmt.Sprintf("d=%v", cookie))
	r, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer r.Body.Close()
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}
