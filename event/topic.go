package event

import (
	"fmt"
)

const (
	ResourceAny = "*"
	DomainAny   = "*"
)

// Topic represents an event topic with structured fields.
type Topic struct {
	domain   Domain // High-level category (e.g., "game", "user")
	action   Action
	resource string // Specific entity (e.g., "game123", "player456")
}

// NewTopic creates a new topic instance.
func NewTopic(domain Domain, action Action) Topic {
	return Topic{
		domain:   domain,
		action:   action,
		resource: ResourceAny,
	}
}

func (t Topic) Domain() Domain   { return t.domain }
func (t Topic) Action() Action   { return t.action }
func (t Topic) Resource() string { return t.resource }

func (t Topic) String() string {
	return fmt.Sprintf("%s.%s.%s", t.domain, t.action, t.resource)
}

// SetResource creates a new topic with the specified resource.
func (t Topic) SetResource(resource string) Topic {
	return Topic{
		domain:   t.domain,
		action:   t.action,
		resource: resource,
	}
}

func (t Topic) Match(filter Topic) bool {
	if filter.domain == DomainAny && filter.action == ActionAny && filter.resource == ResourceAny {
		return true
	}

	if t.domain != filter.domain {
		return false
	}
	if filter.action != ActionAny && t.action != filter.action {
		return false
	}
	if filter.resource != ResourceAny && t.resource != filter.resource {
		return false
	}
	return true
}
