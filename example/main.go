package main

import (
	"context"
	"log"
	"os"

	"github.com/ladonsqlmanager"
	"github.com/ory/ladon"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	dbstring := os.Getenv("DB_STRING")
	if dbstring == "" {
		log.Fatalf("DB_STRING not set")
	}

	db, err := gorm.Open(postgres.Open(dbstring), &gorm.Config{})
	if err != nil {
		log.Fatalf("error connecting to db: %s\n", err)
	}

	//var policy = &ladon.DefaultPolicy{
	//	ID:          "2",
	//	Description: "description",
	//	Subjects:    []string{"user"},
	//	Effect:      ladon.AllowAccess,
	//	Resources:   []string{"article:1"},
	//	Actions:     []string{"create", "update"},
	//}

	manager := ladonsqlmanager.New(db, "postgres")
	if err := manager.Init(); err != nil {
		log.Fatalf("error initalizing ladonsqlmanager: %s\n", err)
	}

	warden := &ladon.Ladon{
		Manager: manager,
	}

	ctx := context.Background()
	//if err := warden.Manager.Create(ctx, policy); err != nil {
	//	log.Fatalf("failed to create policy: %s\n", err)
	//}
	r := &ladon.Request{
		Subject:  "user",
		Resource: "article:1",
		Action:   "test",
	}

	pols, err := warden.Manager.FindRequestCandidates(ctx, r)
	if err != nil {
		log.Fatalf("error getting policies: %s\n", err)
	}

	for _, pol := range pols {
		log.Printf("%#v\n", pol)
	}

	err = warden.IsAllowed(ctx, r)

	if err != nil {
		log.Fatalf("error should be allowed: %s\n", err)
	}
}
