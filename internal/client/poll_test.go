package client

import (
	"testing"
	"time"
)

// shortenPoll lowers a polling interval for the duration of a test so wait
// loops that need several attempts do not sleep for their production interval.
func shortenPoll(t *testing.T, interval *time.Duration) {
	t.Helper()

	original := *interval
	*interval = time.Millisecond
	t.Cleanup(func() { *interval = original })
}
