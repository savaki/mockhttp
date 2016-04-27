package mockhttp

import "time"

type NotifyFunc func(code int, method, endpoint string, elapsed time.Duration)

func (fn NotifyFunc) Notify(code int, method, endpoint string, elapsed time.Duration) {
	fn(code, method, endpoint, elapsed)
}

type Notifier interface {
	Notify(code int, method, endpoint string, elapsed time.Duration)
}

type nopNotifier struct {
}

func (n nopNotifier) Notify(code int, method, endpoint string, elapsed time.Duration) {
}
