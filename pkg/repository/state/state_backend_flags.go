package state

// StateLoadFlags defines what parts of a state to load
type StateLoadFlags uint

const (
	StateLoadBasic                StateLoadFlags = 1 << iota // Just the state itself
	StateLoadColumns                                         // Include columns
	StateLoadData                                            // Include data rows
	StateLoadConfigKeyDefinitions                            // Include key definitions
	StateLoadConfigAttributes                                // Include config data

	StateLoadFull       StateLoadFlags = StateLoadBasic | StateLoadColumns | StateLoadData | StateLoadConfigKeyDefinitions | StateLoadConfigAttributes
	StateLoadFullNoData StateLoadFlags = StateLoadBasic | StateLoadColumns | StateLoadConfigKeyDefinitions | StateLoadConfigAttributes
)
