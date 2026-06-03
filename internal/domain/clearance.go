package domain

type ClearanceLevel int

const (
	TierPublic    ClearanceLevel = 0
	TierCouncil   ClearanceLevel = 1
	TierGuild     ClearanceLevel = 2
	TierCorporate ClearanceLevel = 3
	TierArcane    ClearanceLevel = 4
	TierJunimo    ClearanceLevel = 5
)

func (c ClearanceLevel) String() string {
	switch c {
	case TierPublic:
		return "PUBLIC NOTICE"
	case TierCouncil:
		return "COUNCIL EYES ONLY"
	case TierGuild:
		return "GUILD BUSINESS"
	case TierCorporate:
		return "CORPORATE ACCESS"
	case TierArcane:
		return "ARCANE KNOWLEDGE"
	case TierJunimo:
		return "JUNIMO SCRIPT"
	default:
		return "UNKNOWN"
	}
}

func (c ClearanceLevel) Label() string {
	return c.String()
}

type Role string

const (
	RoleMayor     Role = "mayor"
	RoleKeeper    Role = "keeper"
	RoleVillager  Role = "villager"
	RoleAssociate Role = "associate"
)

func MaxClearanceForRole(role Role) ClearanceLevel {
	switch role {
	case RoleMayor:
		return TierJunimo
	case RoleKeeper:
		return TierArcane
	case RoleVillager:
		return TierCouncil
	case RoleAssociate:
		return TierPublic
	default:
		return TierPublic
	}
}

type Department string

const (
	DeptPublic                 Department = "public"
	DepartmentMayorsOffice     Department = "Mayor's Office"
	DepartmentWizardsTower     Department = "Wizard's Tower"
	DepartmentJojaCorp         Department = "Joja Corp"
	DepartmentAdventurersGuild Department = "Adventurer's Guild"
	DepartmentHarveysClinic    Department = "Harvey's Clinic"
	DepartmentCommunityCenter  Department = "Community Center"
	DepartmentCarpentersShop   Department = "Carpenter's Shop"
	DepartmentMuseum           Department = "Museum"
	DepartmentBulletinBoard    Department = "Bulletin Board"
	DepartmentQisOffice        Department = "Mr. Qi's Office"
	DepartmentPierDocks        Department = "Pier & Docks"
)
