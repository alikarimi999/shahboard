package event

import (
	"fmt"
)

// Topic represents an event topic with structured fields.
type Topic struct {
	domain   Domain // High-level category (e.g., "game", "user")
	action   Action
	resource string // Specific entity (e.g., "game123", "player456")
}

// NewTopic creates a new topic instance.
func NewTopic(domain Domain, action Action, resource string) Topic {
	return Topic{
		domain:   domain,
		action:   action,
		resource: resource,
	}
}

func (t Topic) Domain() Domain   { return t.domain }
func (t Topic) Resource() string { return t.resource }

func (t Topic) String() string {
	if t.resource == "" && t.action == "" {
		return string(t.domain)
	}
	if t.resource == "" {
		return fmt.Sprintf("%s.%s", t.domain, t.action)
	}
	if t.action == "" {
		return fmt.Sprintf("%s..%s", t.domain, t.resource)
	}
	return fmt.Sprintf("%s.%s.%s", t.domain, t.action, t.resource)
}

// WithResource creates a new topic with the specified resource.
func (t Topic) WithResource(resource string) Topic {
	return Topic{
		domain:   t.domain,
		action:   t.action,
		resource: resource,
	}
}

func (t Topic) Match(filter Topic) bool {
	if filter.domain == "" && filter.action == ActionAny && filter.resource == "" {
		return true
	}

	if t.domain != filter.domain {
		return false
	}
	if filter.action != ActionAny && t.action != filter.action {
		return false
	}
	if filter.resource != "" && t.resource != filter.resource {
		return false
	}
	return true
}
