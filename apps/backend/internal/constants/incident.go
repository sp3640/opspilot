package constants

// Incident Severity levels
const (
	SeverityP0 = "P0"
	SeverityP1 = "P1"
	SeverityP2 = "P2"
	SeverityP3 = "P3"
	SeverityP4 = "P4"
)

// Incident Status values
const (
	StatusOpen          = "OPEN"
	StatusInvestigating = "INVESTIGATING"
	StatusMitigating    = "MITIGATING"
	StatusResolved      = "RESOLVED"
)

// ValidSeverities is the slice of all valid severity values.
var ValidSeverities = []string{
	SeverityP0,
	SeverityP1,
	SeverityP2,
	SeverityP3,
	SeverityP4,
}

// ValidStatuses is the slice of all valid status values.
var ValidStatuses = []string{
	StatusOpen,
	StatusInvestigating,
	StatusMitigating,
	StatusResolved,
}

// IsValidStatus reports whether status is one of the incident lifecycle
// states. The single source of truth for incident status validity -
// callers should not re-list these values themselves.
func IsValidStatus(status string) bool {
	for _, valid := range ValidStatuses {
		if status == valid {
			return true
		}
	}

	return false
}
