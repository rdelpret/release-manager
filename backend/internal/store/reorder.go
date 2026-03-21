package store

import (
	"context"
)

func (s *Store) ReorderTask(ctx context.Context, taskID, targetGroupID string, newPosition int) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE tasks SET task_group_id = $2, position = $3, updated_at = now()
		WHERE id = $1
	`, taskID, targetGroupID, newPosition)
	return err
}

func (s *Store) ReorderTaskList(ctx context.Context, listID string, newPosition int) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE task_lists SET position = $2
		WHERE id = $1
	`, listID, newPosition)
	return err
}

func (s *Store) ReorderTaskGroup(ctx context.Context, groupID string, newPosition int) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE task_groups SET position = $2
		WHERE id = $1
	`, groupID, newPosition)
	return err
}
