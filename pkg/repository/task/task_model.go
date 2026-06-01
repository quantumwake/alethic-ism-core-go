// Package task is a generic, durable async-job store backed by Postgres. A row
// is enqueued as Pending and a worker dequeues it atomically via Claim
// (SELECT ... FOR UPDATE SKIP LOCKED), heartbeats a lease while it runs, then
// marks it Succeeded/Failed. It is deliberately type-agnostic: the Type field
// and the opaque JSON Progress/Result columns let any service (snapshot,
// download, clone, ...) reuse the same table and worker machinery.
package task

import "time"

// Status is the lifecycle state of a task.
type Status string

const (
	StatusPending   Status = "pending"
	StatusRunning   Status = "running"
	StatusSucceeded Status = "succeeded"
	StatusFailed    Status = "failed"
	StatusCanceled  Status = "canceled"
)

// Type identifies what kind of work a task represents. Values are defined by
// the producing service; the store treats it as an opaque, filterable label.
type Type string

const (
	// TypeSnapshot is a project → immutable snapshot publish (alethic-ism-publish-api).
	TypeSnapshot Type = "snapshot"
)

// Task is one unit of async work.
type Task struct {
	ID     string `gorm:"column:id;type:varchar(36);primaryKey" json:"id"`
	Type   Type   `gorm:"column:type;not null;index" json:"type"`
	Status Status `gorm:"column:status;not null;index" json:"status"`

	// AccountID / ProjectID scope a task to its owner and source project so the
	// UI can list "my tasks for this project".
	AccountID string `gorm:"column:account_id;type:varchar(36);index" json:"account_id,omitempty"`
	ProjectID string `gorm:"column:project_id;type:varchar(36);index" json:"project_id,omitempty"`

	// ReferenceID points at the entity the task produces or operates on
	// (e.g. the share_id for a snapshot).
	ReferenceID string `gorm:"column:reference_id" json:"reference_id,omitempty"`

	// Progress and Result are opaque per-type JSON documents (the store never
	// interprets them). Error holds a human-readable failure message.
	Progress string `gorm:"column:progress;type:text" json:"progress,omitempty"`
	Result   string `gorm:"column:result;type:text" json:"result,omitempty"`
	Error    string `gorm:"column:error;type:text" json:"error,omitempty"`

	// Attempts counts how many times the task has been claimed (incremented by
	// Claim), so a redelivered-after-crash task can be detected.
	Attempts int `gorm:"column:attempts;not null;default:0" json:"attempts"`

	// LeaseOwner / LeaseExpiresAt implement at-most-one-active-worker. A running
	// task whose lease has expired is reclaimed back to Pending.
	LeaseOwner     string     `gorm:"column:lease_owner" json:"-"`
	LeaseExpiresAt *time.Time `gorm:"column:lease_expires_at" json:"-"`

	CreatedAt  time.Time  `gorm:"column:created_at" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"column:updated_at" json:"updated_at"`
	StartedAt  *time.Time `gorm:"column:started_at" json:"started_at,omitempty"`
	FinishedAt *time.Time `gorm:"column:finished_at" json:"finished_at,omitempty"`
}

func (Task) TableName() string {
	return "task"
}

// Terminal reports whether the task has reached a final state.
func (t *Task) Terminal() bool {
	return t.Status == StatusSucceeded || t.Status == StatusFailed || t.Status == StatusCanceled
}
