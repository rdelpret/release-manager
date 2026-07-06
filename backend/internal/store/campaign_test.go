//go:build integration

package store

import (
	"context"
	"fmt"
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

	campaign, err := s.CreateCampaign(context.Background(), user.ID, "Test Release", nil, "single")
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

	c1, _ := s.CreateCampaign(context.Background(), user.ID, "Campaign 1", nil, "single")
	c2, _ := s.CreateCampaign(context.Background(), user.ID, "Campaign 2", nil, "single")
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

	campaign, _ := s.CreateCampaign(context.Background(), user.ID, "Full Test", nil, "single")
	defer s.DeleteCampaign(context.Background(), campaign.ID)

	full, err := s.GetFullCampaign(context.Background(), campaign.ID)
	if err != nil {
		t.Fatalf("failed to get full campaign: %v", err)
	}
	if full.Name != "Full Test" {
		t.Errorf("expected name 'Full Test', got '%s'", full.Name)
	}

	// The "single" template creates 5 task lists.
	if len(full.TaskLists) != 5 {
		t.Fatalf("expected 5 task lists, got %d", len(full.TaskLists))
	}

	// Lists must be ordered by ascending Position.
	for i := 1; i < len(full.TaskLists); i++ {
		if full.TaskLists[i].Position < full.TaskLists[i-1].Position {
			t.Errorf("task lists not ordered by position: index %d (%d) < index %d (%d)",
				i, full.TaskLists[i].Position, i-1, full.TaskLists[i-1].Position)
		}
	}

	// At least one list must have groups, and at least one group must have tasks.
	foundGroup := false
	foundTask := false
	foundSubtask := false
	for _, tl := range full.TaskLists {
		if len(tl.TaskGroups) > 0 {
			foundGroup = true
		}
		for _, tg := range tl.TaskGroups {
			if len(tg.Tasks) > 0 {
				foundTask = true
			}
			for _, task := range tg.Tasks {
				if len(task.Subtasks) > 0 {
					foundSubtask = true
				}
			}
		}
	}
	if !foundGroup {
		t.Error("expected at least one task list to have groups")
	}
	if !foundTask {
		t.Error("expected at least one group to have tasks")
	}
	if !foundSubtask {
		t.Error("expected at least one task to have subtasks")
	}
}

func BenchmarkCreateCampaign(b *testing.B) {
	if os.Getenv("DATABASE_URL") == "" {
		b.Skip("DATABASE_URL not set, skipping integration benchmark")
	}
	s, err := New()
	if err != nil {
		b.Fatalf("failed to create store: %v", err)
	}
	defer s.Close()

	user, err := s.UpsertUser(context.Background(), "bench@example.com", "Bench User", nil)
	if err != nil {
		b.Fatalf("failed to create user: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		campaign, err := s.CreateCampaign(context.Background(), user.ID, fmt.Sprintf("Bench %d", i), nil, "single")
		if err != nil {
			b.Fatalf("failed to create campaign: %v", err)
		}
		s.DeleteCampaign(context.Background(), campaign.ID)
	}
}
