package notifier

import (
	"github.com/stretchr/testify/mock"
)

type MockNotifier struct {
	mock.Mock
}

// SendNotification sends a notification to Slack.
func (n *MockNotifier) SendNotification(message string) error {
	args := n.Called(message)
	return args.Error(0)
}
