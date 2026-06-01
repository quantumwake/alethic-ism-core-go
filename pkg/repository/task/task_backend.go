package task

import (
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/quantumwake/alethic-ism-core-go/pkg/repository"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// errNoTask is an internal sentinel: the claim transaction found nothing to do.
var errNoTask = errors.New("no task available")

type BackendStorage struct {
	*repository.Access
}

func NewBackend(dsn string) *BackendStorage {
	return &BackendStorage{
		Access: repository.NewDataAccess(dsn),
	}
}

// NewBackendFromAccess reuses an existing connection pool (so a service that
// already holds a *repository.Access doesn't open a second pool).
func NewBackendFromAccess(access *repository.Access) *BackendStorage {
	return &BackendStorage{Access: access}
}

// AutoMigrate creates the task table if it does not exist.
func (tb *BackendStorage) AutoMigrate() error {
	return tb.DB.AutoMigrate(&Task{})
}

// Create enqueues a new task. ID defaults to a new uuid and Status to Pending
// when unset.
func (tb *BackendStorage) Create(t *Task) (*Task, error) {
	if t == nil {
		return nil, fmt.Errorf("task cannot be nil")
	}
	if t.ID == "" {
		t.ID = uuid.NewString()
	}
	if t.Status == "" {
		t.Status = StatusPending
	}
	if err := tb.DB.Create(t).Error; err != nil {
		return nil, fmt.Errorf("error creating task: %w", err)
	}
	return t, nil
}

// Get returns a task by id.
func (tb *BackendStorage) Get(id string) (*Task, error) {
	var t Task
	if err := tb.DB.Where("id = ?", id).First(&t).Error; err != nil {
		return nil, err
	}
	return &t, nil
}

// FindByProject returns tasks for a project, newest first, optionally filtered
// by status.
func (tb *BackendStorage) FindByProject(projectID string, statuses ...Status) ([]Task, error) {
	return tb.find("project_id = ?", projectID, statuses)
}

// FindByAccount returns tasks for an account, newest first, optionally filtered
// by status.
func (tb *BackendStorage) FindByAccount(accountID string, statuses ...Status) ([]Task, error) {
	return tb.find("account_id = ?", accountID, statuses)
}

func (tb *BackendStorage) find(where string, arg any, statuses []Status) ([]Task, error) {
	var tasks []Task
	q := tb.DB.Where(where, arg)
	if len(statuses) > 0 {
		q = q.Where("status IN ?", statuses)
	}
	if err := q.Order("created_at DESC").Find(&tasks).Error; err != nil {
		return nil, fmt.Errorf("error finding tasks: %w", err)
	}
	return tasks, nil
}

// Claim atomically dequeues the oldest Pending task (optionally restricted to
// the given types), marks it Running, stamps a lease owned by ownerID that
// expires after leaseFor, and returns it. Returns (nil, nil) when the queue is
// empty. Safe to call concurrently from many workers/pods — FOR UPDATE SKIP
// LOCKED guarantees each row goes to exactly one claimant.
func (tb *BackendStorage) Claim(ownerID string, leaseFor time.Duration, types ...Type) (*Task, error) {
	now := time.Now().UTC()
	expires := now.Add(leaseFor)

	var t Task
	err := tb.DB.Transaction(func(tx *gorm.DB) error {
		sub := tx.Model(&Task{}).
			Select("id").
			Where("status = ?", StatusPending)
		if len(types) > 0 {
			sub = sub.Where("type IN ?", types)
		}
		sub = sub.Order("created_at").
			Limit(1).
			Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"})

		res := tx.Model(&t).
			Clauses(clause.Returning{}).
			Where("id = (?)", sub).
			Updates(map[string]any{
				"status":           StatusRunning,
				"lease_owner":      ownerID,
				"lease_expires_at": expires,
				"started_at":       gorm.Expr("COALESCE(started_at, ?)", now),
				"attempts":         gorm.Expr("attempts + 1"),
				"updated_at":       now,
			})
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return errNoTask
		}
		return nil
	})
	if errors.Is(err, errNoTask) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("claim task: %w", err)
	}
	return &t, nil
}

// Heartbeat extends the lease and (optionally) updates the opaque progress
// document. It only applies while the caller still owns the lease, so a worker
// whose task was reclaimed stops affecting it.
func (tb *BackendStorage) Heartbeat(id, ownerID string, leaseFor time.Duration, progress string) error {
	now := time.Now().UTC()
	updates := map[string]any{
		"lease_expires_at": now.Add(leaseFor),
		"updated_at":       now,
	}
	if progress != "" {
		updates["progress"] = progress
	}
	res := tb.DB.Model(&Task{}).
		Where("id = ? AND lease_owner = ? AND status = ?", id, ownerID, StatusRunning).
		Updates(updates)
	if res.Error != nil {
		return fmt.Errorf("heartbeat task: %w", res.Error)
	}
	return nil
}

// Complete marks a task Succeeded, records the opaque result and clears the
// lease.
func (tb *BackendStorage) Complete(id, result string) error {
	return tb.finish(id, StatusSucceeded, result, "")
}

// Fail marks a task Failed and records the error message.
func (tb *BackendStorage) Fail(id, errMsg string) error {
	return tb.finish(id, StatusFailed, "", errMsg)
}

func (tb *BackendStorage) finish(id string, status Status, result, errMsg string) error {
	now := time.Now().UTC()
	updates := map[string]any{
		"status":           status,
		"finished_at":      now,
		"updated_at":       now,
		"lease_owner":      "",
		"lease_expires_at": nil,
	}
	if result != "" {
		updates["result"] = result
	}
	if errMsg != "" {
		updates["error"] = errMsg
	}
	// Only finalize a still-running task: a task canceled mid-run must stay
	// canceled rather than be overwritten by a late Complete/Fail.
	if err := tb.DB.Model(&Task{}).Where("id = ? AND status = ?", id, StatusRunning).Updates(updates).Error; err != nil {
		return fmt.Errorf("finish task: %w", err)
	}
	return nil
}

// ReclaimStale returns running tasks whose lease has expired (a worker died) to
// Pending so another worker can retry them. Returns the number reclaimed; call
// on startup and periodically.
func (tb *BackendStorage) ReclaimStale() (int64, error) {
	now := time.Now().UTC()
	res := tb.DB.Model(&Task{}).
		Where("status = ? AND lease_expires_at IS NOT NULL AND lease_expires_at < ?", StatusRunning, now).
		Updates(map[string]any{
			"status":           StatusPending,
			"lease_owner":      "",
			"lease_expires_at": nil,
			"updated_at":       now,
		})
	if res.Error != nil {
		return 0, fmt.Errorf("reclaim stale tasks: %w", res.Error)
	}
	return res.RowsAffected, nil
}

// Cancel marks a non-terminal task Canceled. A running worker observes this via
// its next Heartbeat (which no longer matches status=running) and should stop.
func (tb *BackendStorage) Cancel(id string) error {
	now := time.Now().UTC()
	res := tb.DB.Model(&Task{}).
		Where("id = ? AND status IN ?", id, []Status{StatusPending, StatusRunning}).
		Updates(map[string]any{
			"status":           StatusCanceled,
			"finished_at":      now,
			"updated_at":       now,
			"lease_owner":      "",
			"lease_expires_at": nil,
		})
	if res.Error != nil {
		return fmt.Errorf("cancel task: %w", res.Error)
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("task %s not cancelable", id)
	}
	return nil
}
