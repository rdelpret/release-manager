//go:build integration

package store

import (
	"context"
	"testing"
)

func createTestGroup(t *testing.T, s *Store, userID string) (campaignID, groupID string) {
	t.Helper()
	ctx := context.Background()
	campaign, err := s.CreateCampaign(ctx, userID, "Test Campaign")
	if err != nil {
		t.Fatalf("failed to create campaign: %v", err)
	}
	var listID string
	err = s.pool.QueryRow(ctx, `
		INSERT INTO task_lists (campaign_id, name, color, position) VALUES ($1, 'Test List', '#3B82F6', 100) RETURNING id
	`, campaign.ID).Scan(&listID)
	if err != nil {
		t.Fatalf("failed to create task list: %v", err)
	}
	var gID string
	err = s.pool.QueryRow(ctx, `
		INSERT INTO task_groups (task_list_id, name, position) VALUES ($1, 'Test Group', 100) RETURNING id
	`, listID).Scan(&gID)
	if err != nil {
		t.Fatalf("failed to create task group: %v", err)
	}
	return campaign.ID, gID
}

func TestCreateTask(t *testing.T) {
	s := setupTestStore(t)
	user := createTestUser(t, s)
	campaignID, groupID := createTestGroup(t, s, user.ID)
	defer s.DeleteCampaign(context.Background(), campaignID)

	task, err := s.CreateTask(context.Background(), groupID, "New Task")
	if err != nil {
		t.Fatalf("failed to create task: %v", err)
	}
	if task.Name != "New Task" {
		t.Errorf("expected name 'New Task', got '%s'", task.Name)
	}
	if task.Status != "todo" {
		t.Errorf("expected status 'todo', got '%s'", task.Status)
	}
}

func TestUpdateTask(t *testing.T) {
	s := setupTestStore(t)
	user := createTestUser(t, s)
	campaignID, groupID := createTestGroup(t, s, user.ID)
	defer s.DeleteCampaign(context.Background(), campaignID)

	task, _ := s.CreateTask(context.Background(), groupID, "To Update")
	updates := TaskUpdate{Status: strPtr("done")}
	updated, err := s.UpdateTask(context.Background(), task.ID, updates)
	if err != nil {
		t.Fatalf("failed to update task: %v", err)
	}
	if updated.Status != "done" {
		t.Errorf("expected status 'done', got '%s'", updated.Status)
	}
}

func TestSubtasks(t *testing.T) {
	s := setupTestStore(t)
	user := createTestUser(t, s)
	campaignID, groupID := createTestGroup(t, s, user.ID)
	defer s.DeleteCampaign(context.Background(), campaignID)

	task, _ := s.CreateTask(context.Background(), groupID, "With Subtasks")
	taskID := task.ID

	subtask, err := s.CreateSubtask(context.Background(), taskID, "Sub 1")
	if err != nil {
		t.Fatalf("failed to create subtask: %v", err)
	}
	if subtask.Name != "Sub 1" {
		t.Errorf("expected 'Sub 1', got '%s'", subtask.Name)
	}

	updated, err := s.UpdateSubtask(context.Background(), subtask.ID, nil, boolPtr(true))
	if err != nil {
		t.Fatalf("failed to toggle subtask: %v", err)
	}
	if !updated.IsComplete {
		t.Error("expected subtask to be complete")
	}
}

func strPtr(s string) *string { return &s }
func boolPtr(b bool) *bool    { return &b }
