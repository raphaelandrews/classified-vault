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

type Faction string

const (
	FactionMayorsOffice     Faction = "Mayor's Office"
	FactionWizardsTower     Faction = "Wizard's Tower"
	FactionJojaCorp         Faction = "Joja Corp"
	FactionAdventurersGuild Faction = "Adventurer's Guild"
	FactionHarveysClinic    Faction = "Harvey's Clinic"
	FactionCommunityCenter  Faction = "Community Center"
	FactionCarpentersShop   Faction = "Carpenter's Shop"
	FactionMuseum           Faction = "Museum"
	FactionBulletinBoard    Faction = "Bulletin Board"
	FactionQisOffice        Faction = "Mr. Qi's Office"
	FactionPierDocks        Faction = "Pier & Docks"
)
