package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/slack-go/slack"
	"github.com/tmc/langchaingo/llms/openai"
	"github.com/tmc/langchaingo/schema"
)

func (s *slackClient) autoRespond(ctx context.Context) error {
	rtm := s.Client.NewRTM()
	go rtm.ManageConnection()
	go s.waitForIncomingMessages(ctx, rtm)
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
}

var channelsToRespondIn = map[string]bool{
	"C05AN541ZB8": true,
	// "C05ART3FPEH": true,
	"D05AUNB87JP": true,
}

func (s *slackClient) waitForIncomingMessages(ctx context.Context, rtm *slack.RTM) {
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-rtm.IncomingEvents:
			if err := s.handleRTMEvent(ctx, rtm, &msg); err != nil {
				log.Println(err)
			}
		}
	}
}

func (s *slackClient) handleRTMEvent(ctx context.Context, rtm *slack.RTM, msg *slack.RTMEvent) error {
	switch ev := msg.Data.(type) {
	case *slack.MessageEvent:
		fmt.Println("NEW MESSAGE:", ev.Channel)

		ok := channelsToRespondIn[ev.Channel]
		if !ok {
			fmt.Println("ignoring message from channel:", ev.Channel)
			return nil
		}
		rtm.SendMessage(rtm.NewTypingMessage(ev.Channel))
		j, _ := json.Marshal(ev)
		fmt.Println("NEW MESSAGE:", string(j))

		fmt.Println("getting history")
		h, err := s.dumpConversation(ctx, "C05AN541ZB8", 50)
		if err != nil {
			return err
		}
		rtm.SendMessage(rtm.NewTypingMessage(ev.Channel))
		fmt.Println("done getting history")

		messages := []schema.ChatMessage{
			schema.SystemChatMessage{
				Text: `You are MLOpsGPT. A helpful slack assistant that helps answer questsions.

If you're not confident about your answer return 'NOTHING'.`},
		}

		myID := "U05ADH71NT1"
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
			if role == "user" {
				messages = append(messages, schema.HumanChatMessage{
					Text: h[i].Text,
				})
			} else {
				messages = append(messages, schema.AIChatMessage{
					Text: h[i].Text,
				})
			}
			// messages = append(messages, openai.ChatMessage{
			// 	Role: role, Content: h[i].Text, Name: h[i].User,
			// })
		}
		openaiChatClient, err := openai.New(openai.WithModel("gpt-4"))
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(messages)
		completion, err := openaiChatClient.Chat(ctx, messages)

		fmt.Println(completion, err)

		bannedPhrases := []string{
			"NOTHING",
			"As of August 2021",
			"As a large language model",
		}
		banned := false
		content := completion.Message.Text
		for _, phrase := range bannedPhrases {
			if strings.Contains(strings.ToLower(content), strings.ToLower(phrase)) {
				banned = true
			}
		}
		if banned {
			fmt.Println("BANNED PHRASE DETECTED:", content)
		} else {
			fmt.Println(s.sendMessage(ctx, ev.Channel, content))
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
