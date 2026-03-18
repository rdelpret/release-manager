//go:build integration

package store

import (
	"context"
	"os"
	"testing"

	"github.com/rdelpret/music-release-planner/backend/internal/model"
)

func setupTestStore(t *testing.T) *Store {
	t.Helper()
	if os.Getenv("DATABASE_URL") == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}
	s, err := New()
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func createTestUser(t *testing.T, s *Store) *model.User {
	t.Helper()
	user, err := s.UpsertUser(context.Background(), "test@example.com", "Test User", nil)
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	return user
}

func TestCreateCampaign(t *testing.T) {
	s := setupTestStore(t)
	user := createTestUser(t, s)

	campaign, err := s.CreateCampaign(context.Background(), user.ID, "Test Release")
	if err != nil {
		t.Fatalf("failed to create campaign: %v", err)
	}
	if campaign.Name != "Test Release" {
		t.Errorf("expected name 'Test Release', got '%s'", campaign.Name)
	}
	if campaign.CreatedBy != user.ID {
		t.Errorf("expected created_by '%s', got '%s'", user.ID, campaign.CreatedBy)
	}

	s.DeleteCampaign(context.Background(), campaign.ID)
}

func TestListCampaigns(t *testing.T) {
	s := setupTestStore(t)
	user := createTestUser(t, s)

	c1, _ := s.CreateCampaign(context.Background(), user.ID, "Campaign 1")
	c2, _ := s.CreateCampaign(context.Background(), user.ID, "Campaign 2")
	defer s.DeleteCampaign(context.Background(), c1.ID)
	defer s.DeleteCampaign(context.Background(), c2.ID)

	campaigns, err := s.ListCampaigns(context.Background(), user.ID)
	if err != nil {
		t.Fatalf("failed to list campaigns: %v", err)
	}
	if len(campaigns) < 2 {
		t.Errorf("expected at least 2 campaigns, got %d", len(campaigns))
	}
}

func TestGetFullCampaign(t *testing.T) {
	s := setupTestStore(t)
	user := createTestUser(t, s)

	campaign, _ := s.CreateCampaign(context.Background(), user.ID, "Full Test")
	defer s.DeleteCampaign(context.Background(), campaign.ID)

	full, err := s.GetFullCampaign(context.Background(), campaign.ID)
	if err != nil {
		t.Fatalf("failed to get full campaign: %v", err)
	}
	if full.Name != "Full Test" {
		t.Errorf("expected name 'Full Test', got '%s'", full.Name)
	}
	// Template population is a stub until Task 7 — no task lists expected yet
	// After Task 7, this test should verify len(full.TaskLists) == 5
}
