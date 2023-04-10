package notifier

import (
	"fmt"
	"github.com/slack-go/slack"
)

type SlackNotifier struct {
	client    *slack.Client
	channelID string
}

// NewSlackNotifier creates a new SlackNotifier.
func NewSlackNotifier(token string, channelID string) *SlackNotifier {
	return &SlackNotifier{
		client:    slack.New(token),
		channelID: channelID,
	}
}

// SendNotification sends a notification to Slack.
func (n *SlackNotifier) SendNotification(message string) error {
	_, _, err := n.client.PostMessage(n.channelID, slack.MsgOptionText(message, false))
	if err != nil {
		return fmt.Errorf("error sending notification to Slack: %s", err.Error())
	}
	return nil
}
