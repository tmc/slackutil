package main

import (
	"context"
	"fmt"

	"github.com/slack-go/slack"
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
