package processor

// StateDirection represents the direction of the processor <> state (the state is an input to the processor, or an output to the processor).
type StateDirection string

const (
	DirectionInput  StateDirection = "INPUT"
	DirectionOutput StateDirection = "OUTPUT"
)
