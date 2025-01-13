package event

import (
	"fmt"
	"strings"
)

// Topic represents an event topic with structured fields.
type Topic struct {
	domain   Domain // High-level category (e.g., "game", "user")
	resource string // Specific entity (e.g., "game123", "player456")
}

// NewTopic creates a new topic instance.
func NewTopic(domain Domain, resource string) Topic {
	return Topic{
		domain:   domain,
		resource: resource,
	}
}

func (t Topic) Domain() Domain   { return t.domain }
func (t Topic) Resource() string { return t.resource }

func (t Topic) String() string {
	if t.resource != "" {
		return fmt.Sprintf("%s.%s", t.domain, t.resource)
	}
	return string(t.domain)
}

// WithResource creates a new topic with the specified resource.
func (t Topic) WithResource(resource string) Topic {
	return Topic{
		domain:   t.domain,
		resource: resource,
	}
}

func (t Topic) Matches(filter string) bool {
	filterParts := strings.Split(filter, ".")
	topicParts := strings.Split(t.String(), ".")

	// If lengths differ, they can't match
	if len(filterParts) != len(topicParts) {
		return false
	}

	for i, part := range filterParts {
		if part == "*" || part == topicParts[i] || strings.HasPrefix(topicParts[i], "{") && strings.HasSuffix(topicParts[i], "}") {
			continue
		}
		return false
	}
	return true
}

// DecodeTopic parses a string into a Topic struct.
func DecodeTopic(topicStr string) (Topic, error) {
	parts := strings.Split(topicStr, ".")

	switch len(parts) {
	case 1: // Format: domain
		return Topic{
			domain: Domain(parts[0]),
		}, nil
	case 2: // Format: domain.resource
		return Topic{
			domain:   Domain(parts[0]),
			resource: parts[1],
		}, nil
	default:
		return Topic{}, fmt.Errorf("invalid topic format: %s", topicStr)
	}
}
