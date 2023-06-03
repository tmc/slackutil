package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/slack-go/slack"
	"github.com/tmc/langchaingo/llms"
	"github.com/tmc/langchaingo/llms/openai"
)

func (s *slackClient) autoRespond(ctx context.Context) error {
	rtm := s.Client.NewRTM()
	go rtm.ManageConnection()
	for {
		select {
		case msg := <-rtm.IncomingEvents:
			if err := s.handleRTMEvent(ctx, rtm, &msg); err != nil {
				return err
			}
		case <-ctx.Done():
			return nil
		}
	}
	return nil
}

func (s *slackClient) handleRTMEvent(ctx context.Context, rtm *slack.RTM, msg *slack.RTMEvent) error {
	switch ev := msg.Data.(type) {
	case *slack.MessageEvent:
		if ev.Channel != "D03N9E9RYTS" {
			return nil
		}
		rtm.SendMessage(rtm.NewTypingMessage(ev.Channel))
		j, _ := json.Marshal(ev)
		fmt.Println("NEW MESSAGE:", string(j))

		fmt.Println("getting history")
		h, err := s.dumpConversation(ctx, "D03N9E9RYTS", 50)
		if err != nil {
			return err
		}
		rtm.SendMessage(rtm.NewTypingMessage(ev.Channel))
		fmt.Println("done getting history")
		messages := []openai.ChatMessage{
			{Role: "system", Content: `You are a witty and sarcastic.
Make fun of the user, say something funny, be mean towards them. Try to outwit them.

If you're not confident about your answer return 'NOTHING'.`},
		}

		myID := "U04A8LSNZ"
		if ev.User == myID {
			return nil
		}
		s.conversationHistory["D03N9E9RYTS"] = h
		// walk over conversation history in reverse and build up chat messages:
		for i := len(h) - 1; i >= 0; i-- {
			role := "user"
			if h[i].User == myID {
				role = "assistant"
			}
			messages = append(messages, openai.ChatMessage{
				Role: role, Content: h[i].Text, Name: h[i].User,
			})
		}
		openaiChatClient, err := openai.New()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(messages)
		completion, err := openaiChatClient.Chat(ctx, messages, llms.WithModel("gpt-4"))

		fmt.Println(completion, err)

		bannedPhrases := []string{
			"NOTHING",
			"As of August 2021",
			"As a large language model",
		}
		banned := false
		for _, phrase := range bannedPhrases {
			if strings.Contains(strings.ToLower(completion.Content), strings.ToLower(phrase)) {
				banned = true
			}
		}
		if banned {
			fmt.Println("BANNED PHRASE DETECTED:", completion.Content)
			fmt.Println(s.sendMessage(ctx, ev.Channel, "uhhh, dunno dude"))
		} else {
			fmt.Println(s.sendMessage(ctx, ev.Channel, completion.Content))
		}

		// case *slack.TypingEvent:
		// 	fmt.Printf("User %q is typing.\n", ev.User)
	case *slack.ConnectingEvent:
	case *slack.ConnectedEvent:
	case *slack.HelloEvent:
	case *slack.LatencyReport:
	case *slack.UserTypingEvent:
		rtm.SendMessage(rtm.NewTypingMessage(ev.Channel))
	default:
		// Ignore other events..
		fmt.Printf("Unexpected event: %T: %v\n", msg.Data, msg.Data)
	}
	return nil
}
