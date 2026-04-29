// Package work implements the local-first domain store for the work CLI.
package work

import "time"

const (
	// DefaultStoreDir is the repository-relative store path used by the CLI.
	DefaultStoreDir = ".work"
)

type InboxStatus string

const (
	InboxStatusOpen     InboxStatus = "open"
	InboxStatusAccepted InboxStatus = "accepted"
)

type WorkStatus string

const (
	WorkStatusReady     WorkStatus = "ready"
	WorkStatusActive    WorkStatus = "active"
	WorkStatusBlocked   WorkStatus = "blocked"
	WorkStatusDone      WorkStatus = "done"
	WorkStatusCancelled WorkStatus = "cancelled"
)

type EventType string

const (
	EventWorkCreated   EventType = "work.created"
	EventInboxAccepted EventType = "inbox.accepted"
)

// InboxItem is a captured piece of untriaged work.
type InboxItem struct {
	ID         string            `yaml:"id" json:"id"`
	Title      string            `yaml:"title" json:"title"`
	Body       string            `yaml:"body,omitempty" json:"body,omitempty"`
	Source     string            `yaml:"source,omitempty" json:"source,omitempty"`
	Status     InboxStatus       `yaml:"status" json:"status"`
	Labels     []string          `yaml:"labels,omitempty" json:"labels,omitempty"`
	Metadata   map[string]string `yaml:"metadata,omitempty" json:"metadata,omitempty"`
	AcceptedAs string            `yaml:"accepted_as,omitempty" json:"accepted_as,omitempty"`
	CreatedAt  time.Time         `yaml:"created_at" json:"created_at"`
	UpdatedAt  time.Time         `yaml:"updated_at" json:"updated_at"`
	AcceptedAt *time.Time        `yaml:"accepted_at,omitempty" json:"accepted_at,omitempty"`
}

// InboxItemInput describes an inbox item to create.
type InboxItemInput struct {
	Title    string
	Body     string
	Source   string
	Labels   []string
	Metadata map[string]string
}

// WorkItem is the durable unit tracked by the work CLI.
type WorkItem struct {
	ID                 string            `yaml:"id" json:"id"`
	Title              string            `yaml:"title" json:"title"`
	Description        string            `yaml:"description,omitempty" json:"description,omitempty"`
	Status             WorkStatus        `yaml:"status" json:"status"`
	Priority           string            `yaml:"priority,omitempty" json:"priority,omitempty"`
	Area               string            `yaml:"area,omitempty" json:"area,omitempty"`
	Labels             []string          `yaml:"labels,omitempty" json:"labels,omitempty"`
	AcceptanceCriteria []string          `yaml:"acceptance_criteria,omitempty" json:"acceptance_criteria,omitempty"`
	Relations          []Relation        `yaml:"relations,omitempty" json:"relations,omitempty"`
	SourceInboxID      string            `yaml:"source_inbox_id,omitempty" json:"source_inbox_id,omitempty"`
	Metadata           map[string]string `yaml:"metadata,omitempty" json:"metadata,omitempty"`
	CreatedAt          time.Time         `yaml:"created_at" json:"created_at"`
	UpdatedAt          time.Time         `yaml:"updated_at" json:"updated_at"`
}

// WorkItemInput describes a work item to create.
type WorkItemInput struct {
	Title              string
	Description        string
	Status             WorkStatus
	Priority           string
	Area               string
	Labels             []string
	AcceptanceCriteria []string
	Relations          []Relation
	SourceInboxID      string
	Metadata           map[string]string
}

// AcceptInboxOptions controls how an inbox item becomes a work item.
type AcceptInboxOptions struct {
	Title              string
	Description        string
	Status             WorkStatus
	Priority           string
	Area               string
	Labels             []string
	AcceptanceCriteria []string
	Relations          []Relation
	Metadata           map[string]string
}

// Relation links one work item to another durable object.
type Relation struct {
	Type     string `yaml:"type" json:"type"`
	TargetID string `yaml:"target_id" json:"target_id"`
}

// Event is an append-only activity record for a work item.
type Event struct {
	ID         string         `json:"id"`
	WorkItemID string         `json:"work_item_id"`
	Type       EventType      `json:"type"`
	At         time.Time      `json:"at"`
	Message    string         `json:"message,omitempty"`
	Data       map[string]any `json:"data,omitempty"`
}

// WorkItemFilter is used by ListWorkItems and View definitions.
type WorkItemFilter struct {
	IDs      []string     `yaml:"ids,omitempty" json:"ids,omitempty"`
	Statuses []WorkStatus `yaml:"statuses,omitempty" json:"statuses,omitempty"`
	Areas    []string     `yaml:"areas,omitempty" json:"areas,omitempty"`
	Labels   []string     `yaml:"labels,omitempty" json:"labels,omitempty"`
	Text     string       `yaml:"text,omitempty" json:"text,omitempty"`
}

// View is a named saved filter over work items.
type View struct {
	ID          string         `yaml:"id" json:"id"`
	Name        string         `yaml:"name" json:"name"`
	Description string         `yaml:"description,omitempty" json:"description,omitempty"`
	Filter      WorkItemFilter `yaml:"filter,omitempty" json:"filter,omitempty"`
}

// ViewResult is the materialized item list for a saved view.
type ViewResult struct {
	View  View       `json:"view"`
	Items []WorkItem `json:"items"`
}
