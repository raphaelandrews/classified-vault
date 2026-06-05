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
		return "TOWN NOTICE"
	case TierCouncil:
		return "GUILD SEALED"
	case TierGuild:
		return "COUNCIL SEALED"
	case TierCorporate:
		return "VAULT SEALED"
	case TierArcane:
		return "ARCANE SEALED"
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
	RoleMayor     Role = "Mayor"
	RoleKeeper    Role = "Director"
	RoleVillager  Role = "Member"
	RoleAssociate Role = "Visitor"
)

func MaxClearanceForRole(role Role) ClearanceLevel {
	switch role {
	case RoleMayor:
		return TierJunimo
	case RoleKeeper:
		return TierCorporate
	case RoleVillager:
		return TierGuild
	case RoleAssociate:
		return TierPublic
	default:
		return TierPublic
	}
}

type DepartmentRole struct {
	Name      string
	Clearance ClearanceLevel
	IsLead    bool
}

var DepartmentRoles = map[Department][]string{
	DepartmentMayorsOffice:     {"Mayor", "Secretary", "Clerk", "Director", "Member", "Visitor"},
	DepartmentWizardsTower:     {"Archmage", "Enchanter", "Apprentice", "Director", "Member", "Visitor"},
	DepartmentHarveysClinic:    {"Doctor", "Nurse", "Medic", "Director", "Member", "Visitor"},
	DepartmentAdventurersGuild: {"Guildmaster", "Adventurer", "Scout", "Director", "Member", "Visitor"},
	DepartmentMuseum:           {"Curator", "Archivist", "Docent", "Director", "Member", "Visitor"},
	DepartmentJojaCorp:         {"Manager", "Supervisor", "Clerk", "Director", "Member", "Visitor"},
	DepartmentCarpentersShop:   {"Master Builder", "Carpenter", "Apprentice", "Director", "Member", "Visitor"},
	DepartmentCommunityCenter:  {"Coordinator", "Organizer", "Volunteer", "Director", "Member", "Visitor"},
	DepartmentBulletinBoard:    {"Editor", "Scribe", "Crier", "Director", "Member", "Visitor"},
	DepartmentQisOffice:        {"Agent", "Operative", "Courier", "Director", "Member", "Visitor"},
	DepartmentPierDocks:        {"Harbormaster", "Sailor", "Fisher", "Director", "Member", "Visitor"},
	DepartmentRovingTrader:     {"Merchant", "Trader", "Hawker", "Director", "Member", "Visitor"},
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
	DepartmentRovingTrader     Department = "Roving Trader"
)

var AllDepartments = []Department{
	DepartmentMayorsOffice,
	DepartmentWizardsTower,
	DepartmentHarveysClinic,
	DepartmentAdventurersGuild,
	DepartmentMuseum,
	DepartmentJojaCorp,
	DepartmentCarpentersShop,
	DepartmentCommunityCenter,
	DepartmentBulletinBoard,
	DepartmentQisOffice,
	DepartmentPierDocks,
	DepartmentRovingTrader,
}
