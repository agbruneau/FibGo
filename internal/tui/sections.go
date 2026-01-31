package tui

// Section represents the focused section of the dashboard.
type Section int

const (
	SectionInput Section = iota
	SectionAlgorithms
	SectionResults
)

// String returns the string representation of the section.
func (s Section) String() string {
	switch s {
	case SectionInput:
		return "Input"
	case SectionAlgorithms:
		return "Algorithms"
	case SectionResults:
		return "Results"
	default:
		return "Unknown"
	}
}

// Next returns the next section in the cycle.
func (s Section) Next() Section {
	switch s {
	case SectionInput:
		return SectionAlgorithms
	case SectionAlgorithms:
		return SectionResults
	case SectionResults:
		return SectionInput
	default:
		return SectionInput
	}
}

// Prev returns the previous section in the cycle.
func (s Section) Prev() Section {
	switch s {
	case SectionInput:
		return SectionResults
	case SectionAlgorithms:
		return SectionInput
	case SectionResults:
		return SectionAlgorithms
	default:
		return SectionInput
	}
}

// AlgoStatus represents the status of an algorithm in the table.
type AlgoStatus int

const (
	StatusIdle AlgoStatus = iota
	StatusRunning
	StatusComplete
	StatusError
)

// String returns the string representation of the algorithm status.
func (s AlgoStatus) String() string {
	switch s {
	case StatusIdle:
		return "Idle"
	case StatusRunning:
		return "Running"
	case StatusComplete:
		return "Complete"
	case StatusError:
		return "Error"
	default:
		return "Unknown"
	}
}

// CalcMode represents the calculation mode.
type CalcMode int

const (
	ModeSingle CalcMode = iota
	ModeCompare
)
