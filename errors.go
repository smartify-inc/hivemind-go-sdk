package hivemind

import (
	"errors"
	"fmt"
)

type HivemindError struct {
	Type       string `json:"type"`
	Title      string `json:"title"`
	Status     int    `json:"status"`
	Detail     string `json:"detail"`
	Instance   string `json:"instance,omitempty"`
	HivemindID string `json:"hivemind_id,omitempty"`
	RequestID  string `json:"request_id,omitempty"`
}

func (e *HivemindError) Error() string {
	return fmt.Sprintf("%s: %s (status %d)", e.Title, e.Detail, e.Status)
}

func IsNotFound(err error) bool {
	var he *HivemindError
	return errors.As(err, &he) && he.Status == 404
}

func IsUnauthorized(err error) bool {
	var he *HivemindError
	return errors.As(err, &he) && he.Status == 401
}

func IsForbidden(err error) bool {
	var he *HivemindError
	return errors.As(err, &he) && he.Status == 403
}

func IsRateLimited(err error) bool {
	var he *HivemindError
	return errors.As(err, &he) && he.Status == 429
}

func IsConflict(err error) bool {
	var he *HivemindError
	return errors.As(err, &he) && he.Status == 409
}

func IsServerError(err error) bool {
	var he *HivemindError
	return errors.As(err, &he) && he.Status >= 500
}
