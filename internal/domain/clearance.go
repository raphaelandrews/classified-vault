package domain

type ClearanceLevel int

const (
	ClearancePublic       ClearanceLevel = 0
	ClearanceRestricted   ClearanceLevel = 1
	ClearanceConfidential ClearanceLevel = 2
	ClearanceSecret       ClearanceLevel = 3
	ClearanceTopSecret    ClearanceLevel = 4
)

func (c ClearanceLevel) String() string {
	switch c {
	case ClearancePublic:
		return "PUBLIC"
	case ClearanceRestricted:
		return "RESTRICTED"
	case ClearanceConfidential:
		return "CONFIDENTIAL"
	case ClearanceSecret:
		return "SECRET"
	case ClearanceTopSecret:
		return "TOP SECRET"
	default:
		return "UNKNOWN"
	}
}

func (c ClearanceLevel) Label() string {
	labels := map[ClearanceLevel]string{
		ClearancePublic:       "PUBLIC",
		ClearanceRestricted:   "RESTRICTED",
		ClearanceConfidential: "CONFIDENTIAL",
		ClearanceSecret:       "SECRET",
		ClearanceTopSecret:    "TOP SECRET",
	}
	return labels[c]
}

type Role string

const (
	RoleAdmin   Role = "admin"
	RoleAnalyst Role = "analyst"
	RoleViewer  Role = "viewer"
	RoleIntern  Role = "intern"
)

func MaxClearanceForRole(role Role) ClearanceLevel {
	switch role {
	case RoleAdmin:
		return ClearanceTopSecret
	case RoleAnalyst:
		return ClearanceConfidential
	case RoleViewer:
		return ClearanceRestricted
	case RoleIntern:
		return ClearancePublic
	default:
		return ClearancePublic
	}
}
