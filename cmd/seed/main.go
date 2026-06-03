package main

import (
	"fmt"
	"log"
	"os"

	"classified-vault/config"
	"classified-vault/internal/auth"
	"classified-vault/internal/domain"
	"classified-vault/internal/repository"
	"github.com/google/uuid"
)

func main() {
	cfg := config.Load()
	dbPath := cfg.DatabasePath
	if v := os.Getenv("DATABASE_PATH"); v != "" {
		dbPath = v
	}

	db, err := repository.Connect(dbPath)
	if err != nil {
		log.Fatalf("connect: %v", err)
	}
	defer db.Close()

	if err := repository.RunMigrations(db); err != nil {
		log.Fatalf("migrations: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	docRepo := repository.NewDocumentRepository(db)

	type villager struct {
		username string
		email    string
		role     domain.Role
		faction  domain.Faction
		tier     domain.ClearanceLevel
		password string
	}

	villagers := []villager{
		{"lewis", "lewis@pelican.valley", domain.RoleMayor, domain.FactionMayorsOffice, domain.TierJunimo, "mayor123"},
		{"marnie", "marnie@pelican.valley", domain.RoleMayor, domain.FactionMayorsOffice, domain.TierArcane, "deputy123"},
		{"rasmodius", "rasmodius@pelican.valley", domain.RoleKeeper, domain.FactionWizardsTower, domain.TierArcane, "wizard123"},
		{"morris", "morris@pelican.valley", domain.RoleKeeper, domain.FactionJojaCorp, domain.TierCorporate, "joja123"},
		{"marlon", "marlon@pelican.valley", domain.RoleKeeper, domain.FactionAdventurersGuild, domain.TierGuild, "guild123"},
		{"gil", "gil@pelican.valley", domain.RoleVillager, domain.FactionAdventurersGuild, domain.TierCouncil, "guild123"},
		{"harvey", "harvey@pelican.valley", domain.RoleKeeper, domain.FactionHarveysClinic, domain.TierGuild, "clinic123"},
		{"junimo", "junimo@pelican.valley", domain.RoleVillager, domain.FactionCommunityCenter, domain.TierCouncil, "bundle123"},
		{"robin", "robin@pelican.valley", domain.RoleVillager, domain.FactionCarpentersShop, domain.TierCouncil, "build123"},
		{"gunther", "gunther@pelican.valley", domain.RoleAssociate, domain.FactionMuseum, domain.TierPublic, "museum123"},
		{"gus", "gus@pelican.valley", domain.RoleAssociate, domain.FactionBulletinBoard, domain.TierPublic, "saloon123"},
		{"qi", "qi@pelican.valley", domain.RoleKeeper, domain.FactionQisOffice, domain.TierJunimo, "qichallenge"},
		{"willy", "willy@pelican.valley", domain.RoleVillager, domain.FactionPierDocks, domain.TierCouncil, "fishmaster"},
		{"krobus", "krobus@pelican.valley", domain.RoleAssociate, domain.FactionPierDocks, domain.TierPublic, "voidshadow"},
	}

	for _, v := range villagers {
		existing, _ := userRepo.FindByUsername(v.username)
		if existing != nil {
			fmt.Printf("skip existing villager: %s\n", v.username)
			continue
		}
		h, _ := auth.HashPassword(v.password)
		user := &domain.User{
			ID:           "usr_" + uuid.New().String()[:12],
			Username:     v.username,
			Email:        v.email,
			Role:         v.role,
			Clearance:    v.tier,
			Faction:      v.faction,
			Active:       true,
			PasswordHash: h,
		}
		if err := userRepo.Create(user); err != nil {
			log.Printf("create villager %s: %v", v.username, err)
		} else {
			fmt.Printf("registered villager: %-12s [%-20s | %-10s | tier %d]\n", v.username, v.faction, v.role, v.tier)
		}
	}

	mayor, _ := userRepo.FindByUsername("lewis")
	if mayor == nil {
		log.Fatal("mayor not found")
	}

	type scroll struct {
		title   string
		content string
		tier    domain.ClearanceLevel
		faction domain.Faction
		folder  string
		tags    []string
		refs    []string
		author  string
	}

	scrolls := []scroll{
		// =========================================================================
		// PUBLIC NOTICES — Bulletin Board & Public
		// =========================================================================

		{
			"Spring Egg Festival Announcement",
			"The annual Egg Festival will be held on the 13th of Spring in the town square. All villagers are invited to participate in the egg hunt. Mayor Lewis will present the grand prize — a Straw Hat — to the winner.",
			domain.TierPublic, domain.FactionBulletinBoard, "",
			[]string{"festival", "spring"}, nil, "gus",
		},
		{
			"Governor's Visit — Luau Preparation",
			"His Excellency the Governor will attend this year's Luau Potluck on Summer 11th. Each villager is expected to contribute one ingredient to the communal soup. The quality of the soup will reflect upon our entire community.",
			domain.TierPublic, domain.FactionBulletinBoard, "",
			[]string{"festival", "summer", "governor"}, nil, "gus",
		},
		{
			"Train Passing Through — Safety Advisory",
			"A freight train is scheduled to pass through the Railroad on Spring 17th between 9:00 AM and 6:00 PM. Villagers are advised to stay clear of the tracks during this time. Falling items may include coal, stone, or geode minerals.",
			domain.TierPublic, domain.FactionBulletinBoard, "",
			[]string{"safety", "infrastructure"}, nil, "gus",
		},

		// =========================================================================
		// MAYOR'S OFFICE — Town governance (tier 1–5)
		// =========================================================================

		{
			"Junimo Language Primer",
			"Partial translation of the ancient Junimo script discovered in the Community Center. The Junimos communicate through bundles of intention — each offering represents a promise. Warning: text resonates with forest energy.",
			domain.TierJunimo, domain.FactionMayorsOffice, "",
			[]string{"ancient", "language"}, nil, "lewis",
		},
		{
			"Grandpa's Evaluation Criteria",
			"On the first day of Year 3, the spirit of Grandpa returns to evaluate the farm. Criteria include: total earnings, skill levels, friendship with villagers, marriage status, and completion of Community Center bundles. Ratings range from 0 to 4 candles lit upon the Shrine of Perfection.",
			domain.TierJunimo, domain.FactionMayorsOffice, "",
			[]string{"farm", "evaluation", "spirit"}, nil, "lewis",
		},
		{
			"Town Budget — Fiscal Year",
			"Fiscal allocation for the current year: Town Hall maintenance (12,000g), Festival organization (8,500g), Road and bridge repair (6,200g), Community Center restoration fund (4,000g), Emergency reserves (3,300g). Total: 34,000g.",
			domain.TierCorporate, domain.FactionMayorsOffice, "",
			[]string{"finance", "budget"}, nil, "lewis",
		},
		{
			"Pelican Town Tax Records",
			"Individual tax contributions for the preceding fiscal year. Pierre's General Store: 4,200g. JojaMart: 8,500g. Marnie's Ranch: 2,100g. Stardrop Saloon: 3,800g. All other residents below the taxable threshold.",
			domain.TierGuild, domain.FactionMayorsOffice, "",
			[]string{"finance", "tax"}, nil, "lewis",
		},
		{
			"Town Hall Renovation Plans",
			"Proposed renovation of Town Hall includes: new information kiosk for visitors, expanded museum wing, accessible ramp for elderly villagers, and a portrait gallery honoring town founders. Estimated completion: Winter Year 2.",
			domain.TierCouncil, domain.FactionMayorsOffice, "",
			[]string{"infrastructure", "renovation"}, nil, "lewis",
		},

		// =========================================================================
		// "Secret Notes" folder (Mayor's Office + Qi's Office — cross-faction)
		// =========================================================================

		{
			"Secret Note #19 — Solid Gold Lewis",
			"A peculiar statue discovered hidden behind Mayor Lewis's house in the backyard. The statue depicts Mayor Lewis in solid gold, bearing the inscription 'To our beloved Mayor.' Origin and funding source remain unknown. Mayor Lewis declined to comment when questioned.",
			domain.TierJunimo, domain.FactionMayorsOffice, "Secret Notes",
			[]string{"investigation", "blackmail", "statue"}, nil, "lewis",
		},
		{
			"Secret Note #23 — The Maple Syrup Bear",
			"Multiple villagers have reported a talking bear in the Secret Woods requesting Maple Syrup. The bear appears seasonal (Spring–Fall) and offers unique foraging knowledge in exchange for syrup. The Wizard confirms the bear is not an illusion. The bear's knowledge includes: salmonberry locations, blackberry locations, and a secret method to triple foraged berry yields.",
			domain.TierCorporate, domain.FactionQisOffice, "Secret Notes",
			[]string{"wildlife", "bear", "foraging"}, nil, "qi",
		},
		{
			"Secret Note #10 — The Qi Challenge",
			"You have been challenged to reach level 25 in the Skull Cavern. If you succeed, you will receive a substantial reward. Do not bring any items you are not willing to lose. The clock is ticking. — Mr. Qi",
			domain.TierCorporate, domain.FactionQisOffice, "Secret Notes",
			[]string{"skull-cavern", "challenge"}, nil, "qi",
		},

		// =========================================================================
		// WIZARD'S TOWER — Arcane knowledge
		// =========================================================================

		{
			"Warp Totem Authorization: Beach",
			"Rasmodius authorizes the crafting and distribution of Beach Warp Totems to registered farmers. Ingredients required: 1 Hardwood, 2 Coral, 10 Fiber. The totem activates a teleportation sequence upon contact with forest essence. Do NOT use during thunderstorms.",
			domain.TierArcane, domain.FactionWizardsTower, "",
			[]string{"arcane", "teleportation", "beach"}, nil, "rasmodius",
		},
		{
			"Warp Totem: Desert Authorization",
			"Authorization extended for Desert Warp Totem distribution. Recipe: 2 Hardwood, 1 Coconut, 4 Iridium Ore. The Calico Desert is accessible via the bus after vault bundle completion; totems provide alternative access for registered holders.",
			domain.TierCorporate, domain.FactionWizardsTower, "",
			[]string{"arcane", "desert", "teleportation"}, nil, "rasmodius",
		},
		{
			"Void Essence Containment Report",
			"The void essence levels in the Sewers have risen 17% since the Strange Capsule incident. Containment wards are holding, but additional void essence from Shadow Brutes in the Mines could breach the seal. Krobus has been cooperative in maintaining the barrier.",
			domain.TierCorporate, domain.FactionWizardsTower, "",
			[]string{"arcane", "dangerous", "void"}, nil, "rasmodius",
		},

		// =========================================================================
		// JOJA CORP — Corporate records
		// =========================================================================

		{
			"Joja Membership Expansion Plan",
			"Strategic plan to grow JojaMart membership to 75% of Pelican Town households within two seasons. Tactics: discount coupon mailers, gold-star produce aisle, membership loyalty rewards, and targeted advertising during festivals. Projected revenue increase: 40%.",
			domain.TierCorporate, domain.FactionJojaCorp, "",
			[]string{"business", "expansion", "marketing"}, nil, "morris",
		},
		{
			"Joja Warehouse Inventory and Logistics",
			"Current warehouse stock: 2,400 units Joja Cola, 1,800 units Joja Brand Flour, 950 units Joja Brand Sugar, 340 units Joja Brand Rice. Supply chain from Zuzu City operates on a bi-weekly truck schedule. Inventory variance of 2.1% acceptable.",
			domain.TierGuild, domain.FactionJojaCorp, "",
			[]string{"business", "inventory", "logistics"}, nil, "morris",
		},

		// =========================================================================
		// ADVENTURER'S GUILD — Monster & expedition reports
		// =========================================================================

		{
			"Mine Level 80 — Skeleton Infestation Report",
			"Multiple skeleton warriors reported on mine levels 71–79. Guild members are advised to equip bone swords or better before descending. The skeletal mages at level 80 can cast ranged shadow bolts. Earthquake on Spring 5th may have opened new caverns.",
			domain.TierGuild, domain.FactionAdventurersGuild, "",
			[]string{"monsters", "mine", "safety"}, nil, "marlon",
		},
		{
			"Skull Cavern Expedition: Level 100",
			"Expedition log from the most recent descent into Skull Cavern. Reached floor 103 before emergency totem extraction. Observed: 12 Iridium nodes, 4 Treasure Rooms, 3 Prismatic Shards recovered. Encountered multiple Serpents and one Mummy floor. Qi's evaluation pending.",
			domain.TierGuild, domain.FactionAdventurersGuild, "",
			[]string{"expedition", "skull-cavern", "iridium"}, nil, "marlon",
		},
		{
			"Secret Woods — Shadow Brute Census",
			"Quarterly census of Shadow Brute population in the Secret Woods. Count: 6 Brutes (down from 9 last quarter), 2 Shamans (stable). Hardwood stump regrowth correlates inversely with brute population. Recommend maintaining current patrol schedule.",
			domain.TierCouncil, domain.FactionAdventurersGuild, "",
			[]string{"monsters", "arcane", "forest"}, nil, "marlon",
		},

		// =========================================================================
		// HARVEY'S CLINIC — Medical records
		// =========================================================================

		{
			"Medical Record: Farmer's Energy Levels",
			"A confidential health assessment of the local farmer. Observed energy depletion from 270 units to 45 units after a full day of watering crops and mining. Dietary supplement: 3 Field Snacks and 1 Stardrop consumed in the past season. Recommend more high-energy meals.",
			domain.TierGuild, domain.FactionHarveysClinic, "",
			[]string{"medical", "farmer", "nutrition"}, nil, "harvey",
		},
		{
			"Strange Capsule Incident Report",
			"A mysterious capsule was discovered on the outskirts of the farm on Winter 28th. The capsule appears to be of non-terrestrial origin. After three days the capsule broke open; contents unknown. The Wizard has been consulted. Residual energy signature is faintly arcane.",
			domain.TierGuild, domain.FactionHarveysClinic, "",
			[]string{"medical", "arcane", "anomaly"}, nil, "harvey",
		},
		{
			"Flu Vaccine Distribution Plan",
			"Annual Green Flu vaccination program for Pelican Town residents. Priority groups: elderly (Evelyn, George), children (Vincent, Jas), and those with regular contact with animals (Marnie, the farmer). Vaccines stored in Clinic refrigerator at 4°C.",
			domain.TierCouncil, domain.FactionHarveysClinic, "",
			[]string{"medical", "vaccine", "public-health"}, nil, "harvey",
		},

		// =========================================================================
		// COMMUNITY CENTER — Bundles & festivals
		// =========================================================================

		{
			"Community Center Bundle Progress",
			"Current bundle completion status. Crafts Room: 4/6 complete (remaining: Winter Foraging, Construction). Pantry: 5/6 (remaining: Artisan Bundle). Fish Tank: 2/6. Bulletin Board: 3/5. Vault: 2/4. Boiler Room: 6/6 COMPLETE. Three Junimos observed in the main hall.",
			domain.TierCouncil, domain.FactionCommunityCenter, "",
			[]string{"bundles", "restoration"}, nil, "junimo",
		},
		{
			"Spirit's Eve Maze Layout",
			"Approved maze layout for the Spirit's Eve festival on Fall 27th. The hedgerow maze occupies the central town plaza. Golden Pumpkin prize location: dead-end corridor behind the skeleton statue. Secret path through the ??? area leads to a hidden treat basket.",
			domain.TierCouncil, domain.FactionCommunityCenter, "",
			[]string{"festival", "fall"}, nil, "junimo",
		},
		{
			"Forest Spirit Sighting Log",
			"Increased forest spirit activity reported near the abandoned Farm near Marnie's Ranch. Three villagers reported seeing glowing green figures after 10 PM. The Wizard confirms these are benevolent forest spirits — likely Junimos scouting for potential restoration projects.",
			domain.TierCouncil, domain.FactionCommunityCenter, "",
			[]string{"arcane", "forest", "spirits"}, nil, "junimo",
		},

		// =========================================================================
		// CARPENTER'S SHOP — Infrastructure & permits
		// =========================================================================

		{
			"Shortcut Key Locations — Town Infrastructure",
			"Carpenter Robin's official map of town shortcuts: Behind the Farm → Bus Stop (requires Hardwood bridge), Beach tide pools → Town (requires 300 Wood planks), Mountain Lake → Mines (requires 10 Stone steps), Secret Woods → Lower Forest (requires Hardwood fence passage).",
			domain.TierCouncil, domain.FactionCarpentersShop, "",
			[]string{"infrastructure", "map", "shortcuts"}, nil, "robin",
		},
		{
			"Farm Building Permit #42",
			"Building permit issued for the construction of a Deluxe Barn on the south field of Willow Lane Farm. Structure specifications: 7x4 tiles, capacity 12 animals, auto-feed system included. Foundation inspection passed. Materials: 25,000g, 550 Wood, 300 Stone.",
			domain.TierCouncil, domain.FactionCarpentersShop, "",
			[]string{"construction", "farm", "permit"}, nil, "robin",
		},
		{
			"Town Bridge Repair Assessment",
			"The wooden bridge connecting the Beach to the Tidal Pool area suffered storm damage on Summer 21st. Assessment: 14 planks require replacement, 3 support beams cracked. Estimated repair cost: 300 Wood. The bridge remains passable with caution.",
			domain.TierPublic, domain.FactionCarpentersShop, "",
			[]string{"infrastructure", "repair"}, nil, "robin",
		},

		// =========================================================================
		// MUSEUM — Artifacts & archives
		// =========================================================================

		{
			"Ancient Artifact Catalog",
			"Current catalog of recovered artifacts displayed in the Pelican Town Museum: 14 Dwarf Scrolls (I–IV, partial), 7 Ancient Dolls, 3 Elvish Jewelry pieces, 1 Ancient Sword, 2 Prehistoric Tools, 1 Golden Relic. The Dwarf has provided translation assistance for Scrolls I–III.",
			domain.TierPublic, domain.FactionMuseum, "",
			[]string{"artifacts", "catalog", "history"}, nil, "gunther",
		},
		{
			"Lost Books Recovery Status",
			"Twenty-one of twenty-seven Lost Books have been recovered and restored to the Museum library. Remaining six volumes are believed to be buried in artifact spots throughout the valley. Notable recovered texts: 'The Wizard and the Witch', 'Secrets of the Stardrop', and 'A Study on Yoba'.",
			domain.TierPublic, domain.FactionMuseum, "",
			[]string{"books", "recovery", "library"}, nil, "gunther",
		},

		// =========================================================================
		// FOLDER: "War of the Worlds Broadcast" — Orson Welles radio homage
		// =========================================================================

		{
			"EMERGENCY BROADCAST: Gotoro Invasion Fleet Detected",
			"[BROADCAST TRANSCRIPT — AIRED ON FALL 14TH, 8:00 PM]\n\nURGENT. This is the Pelican Town Emergency Broadcast System. At approximately 7:45 PM, large cylindrical vessels descended from the sky over Zuzu City. The objects, estimated at 30 meters in diameter, unleashed a devastating heat-ray upon the city center. Communication with the Gotoro Empire has been severed. The Ferngill Republic Army has been mobilized. All residents of Pelican Town are advised to: (1) Remain indoors. (2) Barricade all windows and doors. (3) Await further instructions. The Gotoro vessels appear to be advancing eastward toward Stardew Valley. Estimated arrival: 9:30 PM. This is NOT a drill. Repeat: this is NOT a drill.\n\n[END TRANSCRIPT]\n\nSEE ALSO: War of the Worlds folder — BROADCAST RETRACTION.",
			domain.TierPublic, domain.FactionBulletinBoard, "War of the Worlds Broadcast",
			[]string{"broadcast", "gotoro", "invasion", "emergency"}, nil, "gus",
		},
		{
			"BROADCAST RETRACTION: Invasion Report Was Fictional",
			"[RETRACTION — AIRED ON FALL 14TH, 10:30 PM]\n\nTHIS IS A RETRACTION. The earlier broadcast titled 'Gotoro Invasion Fleet Detected' was a dramatic adaptation of the classic radio play 'War of the Worlds' by H.G. Orson, performed by the Stardrop Saloon Theater Troupe. It was NOT a real emergency. We deeply apologize for any distress caused. At least 47 villagers evacuated to the Community Center basement. Mayor Lewis spent the entire hour hiding in a shipping bin. The Saloon will offer free mead tomorrow as compensation.\n\nA full investigation into the broadcast authorization lapse has been opened. The Bulletin Board regrets the oversight.\n\nORIGINAL BROADCAST: See folder 'War of the Worlds Broadcast' for the source scroll.\n\nSTATUS: Panic has subsided. Normal town operations resumed.",
			domain.TierCouncil, domain.FactionBulletinBoard, "War of the Worlds Broadcast",
			[]string{"retraction", "apology", "radio-drama"}, nil, "gus",
		},

		// =========================================================================
		// FOLDER: "X-Files: Qi's Investigations" — Mr. Qi's classified investigation records
		// =========================================================================

		{
			"Strange Capsule Incident — Qi Follow-Up Investigation",
			"[QI INVESTIGATION — TIER 5]\n\nCross-reference: Harvey's Clinic scroll 'Strange Capsule Incident Report'. The capsule that landed on the farm is NOT the first of its kind. I have catalogued 3 prior incidents across the Ferngill Republic over the past 80 years. In all cases, the capsule opens after 3 days and a dark creature emerges. The creature appears to be drawn to arcane energy signatures — particularly the Wizard's Tower and the Secret Woods.\n\nIMPORTANT FINDING: On the 28th day after emergence, the creature leaves behind a broken capsule and a small quantity of Iridium Ore. The creature itself is never seen again. Current hypothesis: these are scouts from a civilization seeking Iridium.\n\nLINKED RECORDS: Harvey's Clinic — Strange Capsule Incident Report (faction: Harvey's Clinic, tier: GUILD BUSINESS). Recommend Mayor override for cross-faction access.",
			domain.TierJunimo, domain.FactionQisOffice, "X-Files: Qi's Investigations",
			[]string{"qi", "alien", "capsule", "investigation"}, nil, "qi",
		},
		{
			"The Prismatic Entity — Skull Cavern Level 100+ Observations",
			"[QI INVESTIGATION — TIER 5]\n\nDuring expeditions beyond Level 100 of the Skull Cavern, I have observed a shimmering, multi-colored entity that appears randomly in treasure rooms. The entity speaks in riddles and grants a single 'Prismatic Shard' to those who answer correctly. I have collected 22 shards from this entity over a 3-year observation period.\n\nThe entity's energy signature matches NO known creature in any bestiary. Spectral analysis reveals temporal distortion around the entity's location — time flows approximately 3.7x slower in its presence. This explains why farmers report 'losing hours' in the deeper cavern levels.\n\nWORKING THEORY: The Prismatic Entity is a temporal custodian, not a hostile. It may be guarding something deeper in the mines. I will continue observation.\n\nCROSS-REF: See Adventurer's Guild scroll 'Skull Cavern Expedition: Level 100'. See Wizard's Tower scroll 'Void Essence Containment Report'.",
			domain.TierJunimo, domain.FactionQisOffice, "X-Files: Qi's Investigations",
			[]string{"qi", "prismatic", "entity", "skull-cavern", "temporal"}, nil, "qi",
		},
		{
			"Time Anomaly Report — Grandpa's Bed",
			"[QI INVESTIGATION — TIER 5]\n\nOn the first day of Year 3, at precisely 6:00 AM, a temporal distortion was detected at Willow Lane Farm. The farmer reported being visited in a dream by the spirit of their deceased grandfather, who evaluated the farm's progress over the past two years. The farmer then woke up at the exact same time (6:00 AM) with a fully restored Shrine of Perfection on the property.\n\nTemporal analysis reveals that the farmer's subjective experience lasted approximately 4 hours (the dream/evaluation), but objective time elapsed was zero. This is consistent with a Class-3 Temporal Insertion event.\n\nNOTE: The four candles on the Shrine of Perfection appear to be linked to the grandfather's evaluation criteria. Each candle represents a category of achievement. When all four are lit, the shrine produces a Statue of Perfection that generates Iridium Ore daily.\n\nCONCLUSION: Grandpa was more than a simple farmer. Investigate any connection to the Iridium Seam.",
			domain.TierJunimo, domain.FactionQisOffice, "X-Files: Qi's Investigations",
			[]string{"qi", "temporal", "grandpa", "shrine"}, nil, "qi",
		},
		{
			"The Wizard's True Identity — Background Investigation",
			"[QI INVESTIGATION — TIER 5]\n\nSubject: M. Rasmodius, self-styled 'Wizard' of Stardew Valley.\n\nBirth records for the Western Forest region show no M. Rasmodius prior to 40 years ago. However, Ferngill Republic Academy of Arcane Arts records from 40 years ago list a 'Magnus Rasmodius' who was expelled for unauthorized research into 'interdimensional communication.' Expulsion was classified, sealed by the Academy.\n\nRasmodius's ex-wife, known only as 'The Witch,' resides in the Witch's Swamp west of the Railroad. She maintains a collection of Void Chickens (see separate scroll: 'Void Chicken Origin Study') and a Dark Shrine with the power to erase memories of ex-spouses — a service used by at least 3 villagers.\n\nCurrent Rasmodius is NOT hostile but IS withholding information about the nature of the Arcane. The Wizard's Tower contains a basement level accessible only through a teleportation circle. Contents unknown.\n\nCREDIBILITY ASSESSMENT: 7/10. The Wizard serves the valley's interests, but his silence on certain topics is suspicious. Continue passive observation.",
			domain.TierJunimo, domain.FactionQisOffice, "X-Files: Qi's Investigations",
			[]string{"qi", "wizard", "investigation", "background"}, nil, "qi",
		},
		{
			"Weekly Paranormal Activity Log — Stardew Valley Region",
			"[QI INVESTIGATION — TIER 4]\n\nWeekly compilation of anomalous events across Stardew Valley:\n\nMON: Three Void Chickens escaped Krobus's sewer enclosure. Recovered near the Mountain Lake. No injuries.\n\nTUE: Forest spirits (Junimos) observed carrying a golden walnut toward the abandoned Farm. Purpose unknown.\n\nWED: Earthquake detected at 1:22 AM, magnitude 2.4. Epicenter: Mines Level 0 entrance. A boulder blocking the Railroad bath house was destroyed by the tremor. Coincidence?\n\nTHU: The Traveling Merchant's cart arrived with a 'rare' artifact — an Ancient Seed inside a clay pot. The Museum authenticated it. The cart's pig was wearing a fez.\n\nFRI: A mermaid was sighted on the Beach at midnight during rain. The Wizard confirms merpeople presence in the Gem Sea. She was singing a tune that translated roughly to 'plant blueberries in summer.'\n\nSAT: 3 Prismatic Shards recovered from Skull Cavern by a local farmer using only a stone pickaxe and 47 plates of sashimi. This farmer's luck stat defies statistical modeling.\n\nWEEKLY ASSESSMENT: Paranormal baseline remains elevated but stable. No hostile activity detected.",
			domain.TierArcane, domain.FactionQisOffice, "X-Files: Qi's Investigations",
			[]string{"qi", "paranormal", "weekly", "log"}, nil, "qi",
		},

		// =========================================================================
		// FOLDER: "The Legendary Fish" — Willy's catalog of mythical catches (Pier & Docks)
		// =========================================================================

		{
			"Legend — The Mountain Lake Monster",
			"[LEGENDARY FISH CATALOG — ENTRY #1]\n\nSpecies: Oncorhynchus legendaris. Common name: Legend.\nLocation: The Mountain Lake, near the submerged log. Spring only, during rain.\nMaximum recorded weight: 55 kg (121 lbs). Largest freshwater fish ever catalogued in the Ferngill Republic.\n\nThe Legend has been pursued by anglers for over 200 years. Willy's grandfather claimed to have hooked it in Spring of 1894, but the line snapped. The fish is said to be intelligent — it will only bite when the angler has Fishing Level 10 and is using an Iridium Rod.\n\nSTATUS: Still at large. This fish is older than Pelican Town itself. Treat with respect.",
			domain.TierCouncil, domain.FactionPierDocks, "The Legendary Fish",
			[]string{"fish", "legend", "mountain-lake"}, nil, "willy",
		},
		{
			"Glacierfish — The Southernmost Catch",
			"[LEGENDARY FISH CATALOG — ENTRY #2]\n\nSpecies: Salvelinus glacialis. Common name: Glacierfish.\nLocation: The southernmost tip of Cindersap Forest, Arrowhead Island. Winter only.\nMaximum recorded weight: 35 kg (77 lbs).\n\nThe Glacierfish thrives in near-freezing water and appears only during the coldest weeks of Winter. Legend says it once swam from the distant Northern Sea through an underground glacial tunnel into the Stardew Valley river system.\n\nAnglers report that the Glacierfish moves erratically underwater, making it one of the hardest fish to reel in. Its flesh is said to taste like 'frozen starlight' — though no one has successfully cooked one.\n\nSTATUS: Caught twice in recorded history. Both times released.",
			domain.TierCouncil, domain.FactionPierDocks, "The Legendary Fish",
			[]string{"fish", "glacierfish", "winter"}, nil, "willy",
		},
		{
			"Crimsonfish — The Eastern Pier Phantom",
			"[LEGENDARY FISH CATALOG — ENTRY #3]\n\nSpecies: Sebastes phantasma. Common name: Crimsonfish.\nLocation: The eastern pier on the Beach. Summer only.\nMaximum recorded weight: 25 kg (55 lbs).\n\nThe Crimsonfish earned its nickname 'Phantom' because it was believed to be a sailor's myth for over a century. It was first officially caught by an angler named Wumbus in Year 1, Summer 15. The angler reported the fish 'looked at me with intelligence before biting.'\n\nWARNING: The Crimsonfish's spines contain a mild neurotoxin that causes temporary memory loss. Two anglers forgot their own names for a period of 4 hours after handling. Wear gloves.\n\nSTATUS: Confirmed species. Present every Summer at the easternmost pier.",
			domain.TierCouncil, domain.FactionPierDocks, "The Legendary Fish",
			[]string{"fish", "crimsonfish", "summer"}, nil, "willy",
		},

		// =========================================================================
		// SUPERNATURAL & EASTER EGGS — Various factions
		// =========================================================================

		{
			"Void Chicken Origin Study",
			"[CLASSIFIED STUDY — KROBUS'S TESTIMONY]\n\nKrobus, a shadow person residing in the Sewer, has disclosed the origin of Void Chickens. They are NOT native to this dimension. According to Krobus, Void Chickens were brought to the valley through a 'shadow portal' approximately 200 years ago by a shadow brigade fleeing persecution in the void realm.\n\nThe chickens adapted rapidly to the valley's climate. Their eggs, while black and occasionally alarming to villagers, are perfectly edible and contain trace amounts of void essence. This essence, when distilled (do NOT try this at home), can be used to craft Void Mayonnaise — a substance coveted by the Witch for 'ritual purposes' (her words).\n\nKrobus currently maintains a small brood of Void Chickens in the Sewers. Population: 6. Egg production: 4/week. Void Chickens are NOT hostile despite their red eyes and occasional possession by shadow spirits. If a Void Chicken attacks, it's probably just hungry.\n\nCROSS-REF: See Wizard's Tower scroll 'Void Essence Containment Report'. See X-Files folder 'Wizard's True Identity'.",
			domain.TierGuild, domain.FactionHarveysClinic, "",
			[]string{"supernatural", "void", "chickens", "krobus"}, nil, "harvey",
		},
		{
			"The Statue of Endless Fortune — Calibration Records",
			"[ARTIFACT ANALYSIS — WIZARD'S TOWER]\n\nThe Statue of Endless Fortune, acquired from the Casino in the Calico Desert, generates one random item per day. Calibration records show the following distribution over a 120-day observation period:\n\nGold Bar:       18 days (15.0%)\nIridium Bar:    12 days (10.0%)\nDiamond:        14 days (11.7%)\nOmni Geode:     21 days (17.5%)\nIridium Ore:    16 days (13.3%)\nPrismatic Shard: 3 days ( 2.5%)\nNo item (bug?):  3 days ( 2.5%)\nOther (mixed):  33 days (27.5%)\n\nNOTE: On Winter 17th, the Statue produced a single gold coin dated from the year 2319. This was verified authentic by both the Museum and the Dwarf. The temporal implications are being investigated by the Wizard.\n\nCROSS-REF: See X-Files folder 'Time Anomaly Report — Grandpa's Bed' for a similar temporal incident.",
			domain.TierArcane, domain.FactionWizardsTower, "",
			[]string{"supernatural", "statue", "temporal", "casino"}, nil, "rasmodius",
		},
		{
			"The Mermaid's Song — Night Market Anomaly Report",
			"[WINTER NIGHT MARKET — ANOMALY REPORT]\n\nDuring the Night Market (Winter 15–17), a mermaid appears on a boat near the eastern pier at approximately 7:00 PM each night. She performs a musical number for the gathered crowd. What the audience perceives as entertainment is, in fact, a coded message.\n\nDecoded from her 5-note sequence: The mermaid's song, when interpreted as a C-Major pentatonic scale mapped to the Stardew Valley clock system, reveals precise timing coordinates for secret fishing spots around the valley. The notes correspond to: 1-5-4-2-3.\n\nOn Winter 17th, a second mermaid was photographed by a tourist. The photo is stored under evidence lock in the Museum. The Wizard believes the merpeople are mapping the valley's coastline for an unknown purpose.\n\nRECOMMENDATION: No action required. Merpeople have been peaceful for centuries. But continue monitoring.",
			domain.TierCouncil, domain.FactionCommunityCenter, "",
			[]string{"supernatural", "mermaid", "night-market", "music"}, nil, "junimo",
		},
		{
			"Iridium Seam Mining Rights Dispute",
			"[INTER-FACTION DISPUTE — MAYOR'S OFFICE]\n\nThe discovery of a massive Iridium Seam beneath Level 115 of the Mines has triggered a legal dispute between Joja Corp and the Adventurer's Guild. Joja claims mining rights under the Ferngill Commerce Act of 1872 (which grants mineral rights to 'corporate entities operating within municipal boundaries'). The Guild claims jurisdiction under the 'Monster Slayer Code' (which grants resource rights to 'those who clear the territory of hostile creatures').\n\nSTATUS: Mayor Lewis has temporarily sealed Mine Levels 110–120 pending resolution. Both factions have submitted briefs. The Iridium Seam is estimated to contain 2,400 kg of pure Iridium Ore — enough to fund the town budget for 10 years.\n\nWARNING: Below Level 120, the mines transition into Skull Cavern territory. Any deep mining operations risk breaching Qi's domain. Mr. Qi has NOT commented on the dispute.\n\nCROSS-REF: See Joja Corp scrolls for expansion plans. See Adventurer's Guild scrolls for expedition logs.",
			domain.TierCorporate, domain.FactionMayorsOffice, "",
			[]string{"dispute", "iridium", "legal", "mines"}, nil, "lewis",
		},

		// =========================================================================
		// SPORTS & EVENTS
		// =========================================================================

		{
			"The Valley Grange Display — Contest Results & Judging Notes",
			"[FALL 16TH — GRANGE DISPLAY CONTEST]\n\nOFFICIAL RESULTS:\n1st Place — Willow Lane Farm: 98 points. Grange displayed: 9 Iridium-quality Ancient Fruit, 3 Truffle Oils, 1 Prismatic Shard, 1 Golden Pumpkin, 1 Rabbit's Foot. Mayor Lewis remarked: 'This is the finest grange I have seen in my 20 years as Mayor.'\n\n2nd Place — Pierre's General Store: 72 points. Vegetables and artisan goods. Solid effort but missing 'the wow factor' according to Judge #2.\n\n3rd Place — Marnie's Ranch: 65 points. Impressive display of animal products. Loses points for having hay mixed into the presentation.\n\nLast Place (disqualified) — JojaMart: 0 points. Morris attempted to enter a display of Joja Cola cans stacked in a pyramid. The pyramid collapsed during judging, covering Mayor Lewis in cola. Morris was asked to leave. The crowd cheered.\n\nJUDGING NOTES: Mayor Lewis (Head Judge), Robin (Craftsmanship), Gus (Culinary), Harvey (Health & Safety).\n\nNEXT YEAR'S THEME SUGGESTION: 'Under the Sea.'",
			domain.TierPublic, domain.FactionBulletinBoard, "",
			[]string{"event", "contest", "fall", "grange"}, nil, "gus",
		},
		{
			"Ice Fishing Competition — Winter Year 2 Final Standings",
			"[WINTER 8TH — ICE FISHING COMPETITION]\n\nOFFICIAL RESULTS:\n1st — Willy (Pier & Docks): 18 fish. Won with a last-minute Lingcod. Prize: 2,000g and a tackle box engraved with 'King of the Ice.'\n\n2nd — Willow Lane Farmer: 15 fish. Came prepared with Dressed Spinners and Cork Bobbers. Lost only because their line froze in the final 10 minutes.\n\n3rd — Pam (Bus Driver): 7 fish. Remarkable performance considering she fished with a bamboo pole while simultaneously drinking a Pale Ale.\n\n4th — Linus (Wild Man): 5 fish. Caught all 5 with his bare hands. Refused the participation prize. 'The mountain needs no trophies.'\n\nLast Place — Lewis (Mayor): 1 fish. The fish was 3 cm long and was later identified as a driftwood splinter. The Mayor blamed 'bad ice,' 'wind direction,' and 'sabotage by Pierre.'\n\nNOTE: The ice measured 22 cm thick. Safety standards were met.",
			domain.TierPublic, domain.FactionPierDocks, "",
			[]string{"event", "fishing", "winter", "competition"}, nil, "willy",
		},
	}

	for _, s := range scrolls {
		authorID := mayor.ID
		if s.author != "" && s.author != "lewis" {
			if a, _ := userRepo.FindByUsername(s.author); a != nil {
				authorID = a.ID
			}
		}

		doc := &domain.Document{
			ID:             "scr_" + uuid.New().String()[:12],
			Title:          s.title,
			Content:        s.content,
			Classification: s.tier,
			Status:         domain.StatusActive,
			Faction:        s.faction,
			Folder:         s.folder,
			Tags:           s.tags,
			ReferenceIDs:   s.refs,
			CreatedBy:      authorID,
		}
		if err := docRepo.Create(doc); err != nil {
			log.Printf("create scroll %s: %v", s.title, err)
		} else {
			folderStr := ""
			if s.folder != "" {
				folderStr = "] ▸ " + s.folder
			}
			fmt.Printf("scribed scroll: %-58s [%-18s | %-15s%s\n", s.title, s.faction, s.tier, folderStr)
		}
	}

	totalScrolls := len(scrolls)
	log.Printf("Seed complete: %d villagers, %d scrolls deposited in the Pelican Town Archives\n", len(villagers), totalScrolls)
}
