package store

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/rdelpret/music-release-planner/backend/internal/model"
)

type TaskUpdate struct {
	Name        *string          `json:"name,omitempty"`
	Description *json.RawMessage `json:"description,omitempty"`
	Status      *string          `json:"status,omitempty"`
	DueDate     *string          `json:"due_date,omitempty"`
}

func (s *Store) CreateTask(ctx context.Context, groupID, name string) (*model.Task, error) {
	var maxPos *int
	s.pool.QueryRow(ctx, `
		SELECT MAX(position) FROM tasks WHERE task_group_id = $1
	`, groupID).Scan(&maxPos)

	nextPos := 100
	if maxPos != nil {
		nextPos = *maxPos + 100
	}

	var task model.Task
	err := s.pool.QueryRow(ctx, `
		INSERT INTO tasks (task_group_id, name, position)
		VALUES ($1, $2, $3)
		RETURNING id, task_group_id, name, description, status, due_date, position, created_at, updated_at
	`, groupID, name, nextPos).Scan(&task.ID, &task.TaskGroupID, &task.Name, &task.Description,
		&task.Status, &task.DueDate, &task.Position, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *Store) UpdateTask(ctx context.Context, taskID string, updates TaskUpdate) (*model.Task, error) {
	setClauses := []string{"updated_at = now()"}
	args := []interface{}{}
	argN := 1

	if updates.Name != nil {
		setClauses = append(setClauses, "name = $"+strconv.Itoa(argN))
		args = append(args, *updates.Name)
		argN++
	}
	if updates.Description != nil {
		setClauses = append(setClauses, "description = $"+strconv.Itoa(argN))
		args = append(args, updates.Description)
		argN++
	}
	if updates.Status != nil {
		setClauses = append(setClauses, "status = $"+strconv.Itoa(argN))
		args = append(args, *updates.Status)
		argN++
	}
	if updates.DueDate != nil {
		if *updates.DueDate == "" {
			setClauses = append(setClauses, "due_date = NULL")
		} else {
			setClauses = append(setClauses, "due_date = $"+strconv.Itoa(argN))
			args = append(args, *updates.DueDate)
			argN++
		}
	}

	args = append(args, taskID)

	query := "UPDATE tasks SET "
	for i, clause := range setClauses {
		if i > 0 {
			query += ", "
		}
		query += clause
	}
	query += " WHERE id = $" + strconv.Itoa(argN)
	query += " RETURNING id, task_group_id, name, description, status, due_date, position, created_at, updated_at"

	var task model.Task
	err := s.pool.QueryRow(ctx, query, args...).Scan(&task.ID, &task.TaskGroupID, &task.Name,
		&task.Description, &task.Status, &task.DueDate, &task.Position, &task.CreatedAt, &task.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &task, nil
}

func (s *Store) DeleteTask(ctx context.Context, taskID string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, taskID)
	return err
}

func (s *Store) CreateSubtask(ctx context.Context, taskID, name string) (*model.Subtask, error) {
	var maxPos *int
	s.pool.QueryRow(ctx, `
		SELECT MAX(position) FROM subtasks WHERE task_id = $1
	`, taskID).Scan(&maxPos)

	nextPos := 100
	if maxPos != nil {
		nextPos = *maxPos + 100
	}

	var subtask model.Subtask
	err := s.pool.QueryRow(ctx, `
		INSERT INTO subtasks (task_id, name, position)
		VALUES ($1, $2, $3)
		RETURNING id, task_id, name, is_complete, position
	`, taskID, name, nextPos).Scan(&subtask.ID, &subtask.TaskID, &subtask.Name, &subtask.IsComplete, &subtask.Position)
	if err != nil {
		return nil, err
	}
	return &subtask, nil
}

func (s *Store) UpdateSubtask(ctx context.Context, subtaskID string, name *string, isComplete *bool) (*model.Subtask, error) {
	setClauses := []string{}
	args := []interface{}{}
	argN := 1

	if name != nil {
		setClauses = append(setClauses, "name = $"+strconv.Itoa(argN))
		args = append(args, *name)
		argN++
	}
	if isComplete != nil {
		setClauses = append(setClauses, "is_complete = $"+strconv.Itoa(argN))
		args = append(args, *isComplete)
		argN++
	}

	if len(setClauses) == 0 {
		return s.getSubtask(ctx, subtaskID)
	}

	args = append(args, subtaskID)

	query := "UPDATE subtasks SET "
	for i, clause := range setClauses {
		if i > 0 {
			query += ", "
		}
		query += clause
	}
	query += " WHERE id = $" + strconv.Itoa(argN)
	query += " RETURNING id, task_id, name, is_complete, position"

	var subtask model.Subtask
	err := s.pool.QueryRow(ctx, query, args...).Scan(&subtask.ID, &subtask.TaskID, &subtask.Name, &subtask.IsComplete, &subtask.Position)
	if err != nil {
		return nil, err
	}
	return &subtask, nil
}

func (s *Store) DeleteSubtask(ctx context.Context, subtaskID string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM subtasks WHERE id = $1`, subtaskID)
	return err
}

func (s *Store) getSubtask(ctx context.Context, subtaskID string) (*model.Subtask, error) {
	var subtask model.Subtask
	err := s.pool.QueryRow(ctx, `
		SELECT id, task_id, name, is_complete, position
		FROM subtasks WHERE id = $1
	`, subtaskID).Scan(&subtask.ID, &subtask.TaskID, &subtask.Name, &subtask.IsComplete, &subtask.Position)
	if err != nil {
		return nil, err
	}
	return &subtask, nil
}

func (s *Store) GetTasksByDueDate(ctx context.Context, campaignID string) ([]model.Task, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT t.id, t.task_group_id, t.name, t.description, t.status, t.due_date, t.position, t.created_at, t.updated_at
		FROM tasks t
		JOIN task_groups tg ON tg.id = t.task_group_id
		JOIN task_lists tl ON tl.id = tg.task_list_id
		WHERE tl.campaign_id = $1 AND t.due_date IS NOT NULL
		ORDER BY t.due_date
	`, campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		var t model.Task
		if err := rows.Scan(&t.ID, &t.TaskGroupID, &t.Name, &t.Description, &t.Status,
			&t.DueDate, &t.Position, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}
