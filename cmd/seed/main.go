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

	hash, _ := auth.HashPassword("admin123")

	users := []struct {
		username string
		email    string
		role     domain.Role
		password string
	}{
		{"admin", "admin@vault.local", domain.RoleAdmin, "admin123"},
		{"ana_analyst", "ana@vault.local", domain.RoleAnalyst, "analyst123"},
		{"vitor_viewer", "vitor@vault.local", domain.RoleViewer, "viewer123"},
		{"igor_intern", "igor@vault.local", domain.RoleIntern, "intern123"},
	}

	for _, u := range users {
		existing, _ := userRepo.FindByUsername(u.username)
		if existing != nil {
			fmt.Printf("skip existing user: %s\n", u.username)
			continue
		}
		user := &domain.User{
			ID:           "usr_" + uuid.New().String()[:12],
			Username:     u.username,
			Email:        u.email,
			Role:         u.role,
			Clearance:    domain.MaxClearanceForRole(u.role),
			Active:       true,
			PasswordHash: hash,
		}
		_ = hash
		password := u.password
		if password == "" {
			password = uuid.New().String()[:12]
		}
		h, _ := auth.HashPassword(password)
		user.PasswordHash = h
		if err := userRepo.Create(user); err != nil {
			log.Printf("create user %s: %v", u.username, err)
		} else {
			fmt.Printf("created user: %s (%s) password=%s\n", u.username, u.role, password)
		}
	}

	admin, _ := userRepo.FindByUsername("admin")
	if admin == nil {
		log.Fatal("admin not found")
	}

	docs := []struct {
		title   string
		content string
		cle     domain.ClearanceLevel
		tags    []string
	}{
		{"Welcome Guide", "This document contains the onboarding guide for new employees. Please read carefully before accessing any classified materials.", domain.ClearancePublic, []string{"onboarding", "hr"}},
		{"Office Security Protocol", "All personnel must badge in at the main entrance. Visitors must be escorted at all times. Report suspicious activity to security immediately.", domain.ClearanceRestricted, []string{"security", "physical"}},
		{"Q1 Vulnerability Assessment", "Critical: We identified 3 high-severity vulnerabilities in the public-facing API. Patches deployed on March 1st. Full details follow.", domain.ClearanceConfidential, []string{"security", "api", "audit"}},
		{"Project Nightfall — Phase 2", "Continuation of the Nightfall initiative. New encryption protocols deployed. Satellite relay confirmed. Full mission parameters enclosed.", domain.ClearanceSecret, []string{"project", "encryption"}},
		{"Operation Silent Watch — Field Report", "Agent 47 confirmed target neutralized. Evidence package delivered to embassy. Awaiting extraction at coordinates 34.0522 N, 118.2437 W.", domain.ClearanceTopSecret, []string{"field-report", "covert"}},
		{"Holiday Party Planning", "The annual holiday party will be held on December 15th. Theme: Tropical Paradise. Sign up sheet in the break room. Bring a dish to share!", domain.ClearancePublic, []string{"social", "events"}},
		{"Server Maintenance Log", "Weekly maintenance: applied OS patches, updated firewall rules, rotated backup tapes. All systems nominal. Next maintenance: Friday.", domain.ClearanceRestricted, []string{"it", "maintenance"}},
	}

	for _, d := range docs {
		doc := &domain.Document{
			ID:             "doc_" + uuid.New().String()[:12],
			Title:          d.title,
			Content:        d.content,
			Classification: d.cle,
			Status:         domain.StatusActive,
			Tags:           d.tags,
			CreatedBy:      admin.ID,
		}
		if err := docRepo.Create(doc); err != nil {
			log.Printf("create doc %s: %v", d.title, err)
		} else {
			fmt.Printf("created document: %s [%s]\n", d.title, d.cle)
		}
	}

	log.Printf("Seed complete: %d users, %d documents\n", len(users), len(docs))
}
