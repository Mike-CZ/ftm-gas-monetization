package notifier

// Notifier is an interface for sending notifications.
type Notifier interface {
	SendNotification(message string) error
}
