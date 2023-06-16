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
	pubsub := s.redisClient.Subscribe(ctx, "outgoing-messages")
	defer pubsub.Close()
	ch := pubsub.Channel()
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-ch:

			pl := msg.Payload
			fromPy := &msgFromython{}
			json.Unmarshal([]byte(pl), fromPy)
			fmt.Println("got message from pubsub:", msg)
			j, _ := json.Marshal(fromPy)
			fmt.Println("got message from pubsub (json decoded)", string(j))

			m := rtm.NewOutgoingMessage(fromPy.Message, fromPy.Channel, slack.RTMsgOptionTS(fromPy.ThreadTS))
			fmt.Println("SENDING MESSAGE VIA RTM:", m)
			rtm.SendMessage(m)
		}
	}
}

type msgToPython struct {
	Channel   string `json:"channel"`
	MessageID string `json:"message_id"`
	ThreadTS  string `json:"thread_ts"`
	Message   string `json:"message"`
}

type msgFromython struct {
	Channel   string `json:"channel"`
	MessageID string `json:"message_id"`
	ThreadTS  string `json:"thread_ts"`
	Message   string `json:"new-message"`
}

func (s *slackClient) handleRTMEvent(ctx context.Context, rtm *slack.RTM, msg *slack.RTMEvent) error {
	switch ev := msg.Data.(type) {
	case *slack.MessageEvent:
		s.handleRTMMessageEvent(ctx, rtm, ev)
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

func (s *slackClient) handleRTMMessageEvent(ctx context.Context, rtm *slack.RTM, ev *slack.MessageEvent) error {
	fmt.Println("NEW MESSAGE:", ev.User, ev.Channel)

	ok := channelsToRespondIn[ev.Channel]
	if !ok {
		fmt.Println("ignoring message from channel:", ev.Channel)
		return nil
	}
	myID := "U05ADH71NT1"
	if ev.User == myID {
		fmt.Println("ignoring message from self:", ev.Channel)
		return nil
	}
	if ev.Msg.Type == "message_changed" || ev.Msg.Type == "message_deleted" {
		fmt.Println("ignoring message change:", ev.Msg.Type)
		return nil
	}
	if ev.Msg.SubType == "message_deleted" || ev.Msg.SubType == "message_changed" {
		fmt.Println("ignoring deleted or changed message")
		return nil
	}

	if ev.Msg.SubType == "message_replied" {
		fmt.Println("ignoring message reply")
		return nil
	}

	j, _ := json.Marshal(ev)
	fmt.Println("not ignoring message:", ev.User, ev.Msg.Type, ev.Msg.SubType, string(j))

	forPy := &msgToPython{
		Channel:   ev.Channel,
		MessageID: ev.Msg.ClientMsgID,
		ThreadTS:  ev.Msg.ThreadTimestamp,
		Message:   ev.Msg.Text,
	}
	if forPy.ThreadTS == "" {
		forPy.ThreadTS = ev.Msg.Timestamp
	}
	j, _ = json.Marshal(forPy)
	s.redisClient.Publish(ctx, "incoming-messages", j)

	rtm.SendMessage(rtm.NewTypingMessage(ev.Channel))

	fmt.Println("getting history")
	h, err := s.dumpConversation(ctx, ev.Channel, false, 50)
	if err != nil {
		return err
	}
	rtm.SendMessage(rtm.NewTypingMessage(ev.Channel))
	fmt.Println("done getting history")

	messages := []schema.ChatMessage{
		schema.SystemChatMessage{
			Text: `You are MLOpsGPT. A helpful slack assistant that helps answer questsions.
Respond only in 2-3 sentences.

If you're not confident about your answer return 'NOTHING'.`},
	}

	return nil // TODO: foo

	s.conversationHistory[ev.Channel] = h
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
		ts := ""
		if ev.ThreadTimestamp != "" {
			ts = ev.ThreadTimestamp
		}
		if ts == "" {
			ts = ev.Timestamp
		}
		fmt.Println(s.sendMessage(ctx, ev.Channel, ts, content))
	}
	return nil
}
