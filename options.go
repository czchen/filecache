package filecache

import (
	"time"
)

type Option func(*options)

type options struct {
	timeToLive      time.Duration
	cleanerInterval time.Duration
}

func getDefaultOptions() options {
	return options{
		timeToLive:      time.Duration(1 * time.Hour),
		cleanerInterval: time.Duration(10 * time.Minute),
	}
}

func WithTTL(ttl time.Duration) func(*options) {
	return func(opts *options) {
		opts.timeToLive = ttl
	}
}

func WithCleanerInterval(interval time.Duration) func(*options) {
	return func(opts *options) {
		opts.cleanerInterval = interval
	}
}
