package processor

// Status represents the possible statuses of a processor and, the processor <> state association.
type Status string

// Enum-like constants for ProcessorStatus
const (
	Created   Status = "CREATED"
	Route     Status = "ROUTE"
	Routed    Status = "ROUTED"
	Queued    Status = "QUEUED"
	Running   Status = "RUNNING"
	Terminate Status = "TERMINATE"
	Stopped   Status = "STOPPED"
	Completed Status = "COMPLETED"
	Failed    Status = "FAILED"
)
