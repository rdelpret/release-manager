//go:build integration

package store

import (
	"context"
	"testing"
)

func TestReorderTask(t *testing.T) {
	s := setupTestStore(t)
	user := createTestUser(t, s)
	campaign, _ := s.CreateCampaign(context.Background(), user.ID, "Reorder Test")
	defer s.DeleteCampaign(context.Background(), campaign.ID)

	full, _ := s.GetFullCampaign(context.Background(), campaign.ID)
	groupID := full.TaskLists[0].TaskGroups[0].ID
	tasks := full.TaskLists[0].TaskGroups[0].Tasks

	if len(tasks) < 2 {
		t.Fatal("need at least 2 tasks for reorder test")
	}

	err := s.ReorderTask(context.Background(), tasks[1].ID, groupID, tasks[0].Position-50)
	if err != nil {
		t.Fatalf("failed to reorder task: %v", err)
	}

	refreshed, _ := s.GetFullCampaign(context.Background(), campaign.ID)
	firstTask := refreshed.TaskLists[0].TaskGroups[0].Tasks[0]
	if firstTask.ID != tasks[1].ID {
		t.Errorf("expected task %s to be first, got %s", tasks[1].ID, firstTask.ID)
	}
}
