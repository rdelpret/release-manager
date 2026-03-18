package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rdelpret/music-release-planner/backend/internal/model"
	"github.com/rdelpret/music-release-planner/backend/internal/template"
)

func (s *Store) CreateCampaign(ctx context.Context, userID, name string) (*model.Campaign, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var campaign model.Campaign
	err = tx.QueryRow(ctx, `
		INSERT INTO campaigns (created_by, name)
		VALUES ($1, $2)
		RETURNING id, created_by, name, archived, created_at, updated_at
	`, userID, name).Scan(&campaign.ID, &campaign.CreatedBy, &campaign.Name,
		&campaign.Archived, &campaign.CreatedAt, &campaign.UpdatedAt)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO campaign_members (campaign_id, user_id, role)
		VALUES ($1, $2, 'owner')
	`, campaign.ID, userID)
	if err != nil {
		return nil, err
	}

	// Populate default template (stub — replaced with real implementation in Task 7)
	if err := s.populateTemplate(ctx, tx, campaign.ID); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &campaign, nil
}

func (s *Store) ListCampaigns(ctx context.Context, userID string) ([]model.Campaign, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT c.id, c.created_by, c.name, c.archived, c.created_at, c.updated_at
		FROM campaigns c
		JOIN campaign_members cm ON cm.campaign_id = c.id
		WHERE cm.user_id = $1
		ORDER BY c.updated_at DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var campaigns []model.Campaign
	for rows.Next() {
		var c model.Campaign
		if err := rows.Scan(&c.ID, &c.CreatedBy, &c.Name, &c.Archived, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		campaigns = append(campaigns, c)
	}
	return campaigns, rows.Err()
}

func (s *Store) GetFullCampaign(ctx context.Context, campaignID string) (*model.Campaign, error) {
	var campaign model.Campaign
	err := s.pool.QueryRow(ctx, `
		SELECT id, created_by, name, archived, created_at, updated_at
		FROM campaigns WHERE id = $1
	`, campaignID).Scan(&campaign.ID, &campaign.CreatedBy, &campaign.Name,
		&campaign.Archived, &campaign.CreatedAt, &campaign.UpdatedAt)
	if err != nil {
		return nil, err
	}

	listRows, err := s.pool.Query(ctx, `
		SELECT id, campaign_id, name, color, position
		FROM task_lists WHERE campaign_id = $1
		ORDER BY position
	`, campaignID)
	if err != nil {
		return nil, err
	}
	defer listRows.Close()

	for listRows.Next() {
		var tl model.TaskList
		if err := listRows.Scan(&tl.ID, &tl.CampaignID, &tl.Name, &tl.Color, &tl.Position); err != nil {
			return nil, err
		}
		campaign.TaskLists = append(campaign.TaskLists, tl)
	}
	if err := listRows.Err(); err != nil {
		return nil, err
	}

	for i := range campaign.TaskLists {
		groups, err := s.getGroupsForList(ctx, campaign.TaskLists[i].ID)
		if err != nil {
			return nil, err
		}
		campaign.TaskLists[i].TaskGroups = groups
	}

	return &campaign, nil
}

func (s *Store) getGroupsForList(ctx context.Context, listID string) ([]model.TaskGroup, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, task_list_id, name, position, collapsed
		FROM task_groups WHERE task_list_id = $1
		ORDER BY position
	`, listID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []model.TaskGroup
	for rows.Next() {
		var g model.TaskGroup
		if err := rows.Scan(&g.ID, &g.TaskListID, &g.Name, &g.Position, &g.Collapsed); err != nil {
			return nil, err
		}
		groups = append(groups, g)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range groups {
		tasks, err := s.getTasksForGroup(ctx, groups[i].ID)
		if err != nil {
			return nil, err
		}
		groups[i].Tasks = tasks
	}
	return groups, nil
}

func (s *Store) getTasksForGroup(ctx context.Context, groupID string) ([]model.Task, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, task_group_id, name, description, status, due_date, position, created_at, updated_at
		FROM tasks WHERE task_group_id = $1
		ORDER BY position
	`, groupID)
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
	if err := rows.Err(); err != nil {
		return nil, err
	}

	for i := range tasks {
		subtasks, err := s.getSubtasksForTask(ctx, tasks[i].ID)
		if err != nil {
			return nil, err
		}
		tasks[i].Subtasks = subtasks
	}
	return tasks, nil
}

func (s *Store) getSubtasksForTask(ctx context.Context, taskID string) ([]model.Subtask, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT id, task_id, name, is_complete, position
		FROM subtasks WHERE task_id = $1
		ORDER BY position
	`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var subtasks []model.Subtask
	for rows.Next() {
		var st model.Subtask
		if err := rows.Scan(&st.ID, &st.TaskID, &st.Name, &st.IsComplete, &st.Position); err != nil {
			return nil, err
		}
		subtasks = append(subtasks, st)
	}
	return subtasks, rows.Err()
}

func (s *Store) ArchiveCampaign(ctx context.Context, campaignID string, archived bool) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE campaigns SET archived = $2, updated_at = $3
		WHERE id = $1
	`, campaignID, archived, time.Now())
	return err
}

func (s *Store) DeleteCampaign(ctx context.Context, campaignID string) error {
	_, err := s.pool.Exec(ctx, `DELETE FROM campaigns WHERE id = $1`, campaignID)
	return err
}

func (s *Store) DuplicateCampaign(ctx context.Context, sourceCampaignID, userID string) (*model.Campaign, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var sourceName string
	err = tx.QueryRow(ctx, `SELECT name FROM campaigns WHERE id = $1`, sourceCampaignID).Scan(&sourceName)
	if err != nil {
		return nil, err
	}

	var campaign model.Campaign
	err = tx.QueryRow(ctx, `
		INSERT INTO campaigns (created_by, name)
		VALUES ($1, $2)
		RETURNING id, created_by, name, archived, created_at, updated_at
	`, userID, sourceName+" (Copy)").Scan(&campaign.ID, &campaign.CreatedBy, &campaign.Name,
		&campaign.Archived, &campaign.CreatedAt, &campaign.UpdatedAt)
	if err != nil {
		return nil, err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO campaign_members (campaign_id, user_id, role)
		VALUES ($1, $2, 'owner')
	`, campaign.ID, userID)
	if err != nil {
		return nil, err
	}

	listRows, err := tx.Query(ctx, `
		SELECT id, name, color, position FROM task_lists
		WHERE campaign_id = $1 ORDER BY position
	`, sourceCampaignID)
	if err != nil {
		return nil, err
	}
	defer listRows.Close()

	type listMapping struct {
		oldID string
		newID string
	}
	var listMappings []listMapping

	for listRows.Next() {
		var oldID, name, color string
		var position int
		if err := listRows.Scan(&oldID, &name, &color, &position); err != nil {
			return nil, err
		}
		var newID string
		err = tx.QueryRow(ctx, `
			INSERT INTO task_lists (campaign_id, name, color, position)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`, campaign.ID, name, color, position).Scan(&newID)
		if err != nil {
			return nil, err
		}
		listMappings = append(listMappings, listMapping{oldID, newID})
	}
	if err := listRows.Err(); err != nil {
		return nil, err
	}

	for _, lm := range listMappings {
		if err := s.duplicateGroupsForList(ctx, tx, lm.oldID, lm.newID); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &campaign, nil
}

func (s *Store) duplicateGroupsForList(ctx context.Context, tx pgx.Tx, oldListID, newListID string) error {
	groupRows, err := tx.Query(ctx, `
		SELECT id, name, position, collapsed FROM task_groups
		WHERE task_list_id = $1 ORDER BY position
	`, oldListID)
	if err != nil {
		return err
	}
	defer groupRows.Close()

	type groupMapping struct {
		oldID string
		newID string
	}
	var groupMappings []groupMapping

	for groupRows.Next() {
		var oldID, name string
		var position int
		var collapsed bool
		if err := groupRows.Scan(&oldID, &name, &position, &collapsed); err != nil {
			return err
		}
		var newID string
		err = tx.QueryRow(ctx, `
			INSERT INTO task_groups (task_list_id, name, position, collapsed)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`, newListID, name, position, collapsed).Scan(&newID)
		if err != nil {
			return err
		}
		groupMappings = append(groupMappings, groupMapping{oldID, newID})
	}
	if err := groupRows.Err(); err != nil {
		return err
	}

	for _, gm := range groupMappings {
		if err := s.duplicateTasksForGroup(ctx, tx, gm.oldID, gm.newID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) duplicateTasksForGroup(ctx context.Context, tx pgx.Tx, oldGroupID, newGroupID string) error {
	taskRows, err := tx.Query(ctx, `
		SELECT id, name, description, status, due_date, position
		FROM tasks WHERE task_group_id = $1 ORDER BY position
	`, oldGroupID)
	if err != nil {
		return err
	}
	defer taskRows.Close()

	type taskMapping struct {
		oldID string
		newID string
	}
	var taskMappings []taskMapping

	for taskRows.Next() {
		var t model.Task
		if err := taskRows.Scan(&t.ID, &t.Name, &t.Description, &t.Status, &t.DueDate, &t.Position); err != nil {
			return err
		}
		var newID string
		err = tx.QueryRow(ctx, `
			INSERT INTO tasks (task_group_id, name, description, status, due_date, position)
			VALUES ($1, $2, $3, 'todo', NULL, $4)
			RETURNING id
		`, newGroupID, t.Name, t.Description, t.Position).Scan(&newID)
		if err != nil {
			return err
		}
		taskMappings = append(taskMappings, taskMapping{t.ID, newID})
	}
	if err := taskRows.Err(); err != nil {
		return err
	}

	for _, tm := range taskMappings {
		_, err := tx.Exec(ctx, `
			INSERT INTO subtasks (task_id, name, is_complete, position)
			SELECT $1, name, false, position
			FROM subtasks WHERE task_id = $2
		`, tm.newID, tm.oldID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) populateTemplate(ctx context.Context, tx pgx.Tx, campaignID string) error {
	tmpl := template.DefaultTemplate()

	for listPos, list := range tmpl {
		var listID string
		err := tx.QueryRow(ctx, `
			INSERT INTO task_lists (campaign_id, name, color, position)
			VALUES ($1, $2, $3, $4)
			RETURNING id
		`, campaignID, list.Name, list.Color, (listPos+1)*100).Scan(&listID)
		if err != nil {
			return fmt.Errorf("inserting task list %s: %w", list.Name, err)
		}

		for groupPos, group := range list.Groups {
			var groupID string
			err := tx.QueryRow(ctx, `
				INSERT INTO task_groups (task_list_id, name, position)
				VALUES ($1, $2, $3)
				RETURNING id
			`, listID, group.Name, (groupPos+1)*100).Scan(&groupID)
			if err != nil {
				return fmt.Errorf("inserting task group %s: %w", group.Name, err)
			}

			for taskPos, task := range group.Tasks {
				var taskID string
				err := tx.QueryRow(ctx, `
					INSERT INTO tasks (task_group_id, name, position)
					VALUES ($1, $2, $3)
					RETURNING id
				`, groupID, task.Name, (taskPos+1)*100).Scan(&taskID)
				if err != nil {
					return fmt.Errorf("inserting task %s: %w", task.Name, err)
				}

				for subPos, subtask := range task.Subtasks {
					_, err := tx.Exec(ctx, `
						INSERT INTO subtasks (task_id, name, position)
						VALUES ($1, $2, $3)
					`, taskID, subtask, (subPos+1)*100)
					if err != nil {
						return fmt.Errorf("inserting subtask %s: %w", subtask, err)
					}
				}
			}
		}
	}
	return nil
}
