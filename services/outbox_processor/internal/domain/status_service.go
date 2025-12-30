package domain

// StatusService updates the notification statuses by utilizing a
// repository.
type StatusService interface {
	// Listen listens to updates on the given channel, and batches repository
	// updates accordingly.
	Listen(updates <-chan StatusUpdate)
}
