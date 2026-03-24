package store

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/rdelpret/music-release-planner/backend/internal/model"
	"github.com/rdelpret/music-release-planner/backend/internal/template"
)

// IsCampaignMember checks if a user is a member of a campaign.
func (s *Store) IsCampaignMember(ctx context.Context, campaignID, userID string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM campaign_members WHERE campaign_id = $1 AND user_id = $2)
	`, campaignID, userID).Scan(&exists)
	return exists, err
}

// IsCampaignMemberViaTask checks if a user owns the campaign that a task belongs to.
func (s *Store) IsCampaignMemberViaTask(ctx context.Context, taskID, userID string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM tasks t
			JOIN task_groups tg ON tg.id = t.task_group_id
			JOIN task_lists tl ON tl.id = tg.task_list_id
			JOIN campaign_members cm ON cm.campaign_id = tl.campaign_id
			WHERE t.id = $1 AND cm.user_id = $2
		)
	`, taskID, userID).Scan(&exists)
	return exists, err
}

// IsCampaignMemberViaSubtask checks if a user owns the campaign that a subtask belongs to.
func (s *Store) IsCampaignMemberViaSubtask(ctx context.Context, subtaskID, userID string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM subtasks st
			JOIN tasks t ON t.id = st.task_id
			JOIN task_groups tg ON tg.id = t.task_group_id
			JOIN task_lists tl ON tl.id = tg.task_list_id
			JOIN campaign_members cm ON cm.campaign_id = tl.campaign_id
			WHERE st.id = $1 AND cm.user_id = $2
		)
	`, subtaskID, userID).Scan(&exists)
	return exists, err
}

// IsCampaignMemberViaGroup checks if a user owns the campaign that a task group belongs to.
func (s *Store) IsCampaignMemberViaGroup(ctx context.Context, groupID, userID string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM task_groups tg
			JOIN task_lists tl ON tl.id = tg.task_list_id
			JOIN campaign_members cm ON cm.campaign_id = tl.campaign_id
			WHERE tg.id = $1 AND cm.user_id = $2
		)
	`, groupID, userID).Scan(&exists)
	return exists, err
}

// IsCampaignMemberViaList checks if a user owns the campaign that a task list belongs to.
func (s *Store) IsCampaignMemberViaList(ctx context.Context, listID, userID string) (bool, error) {
	var exists bool
	err := s.pool.QueryRow(ctx, `
		SELECT EXISTS(
			SELECT 1 FROM task_lists tl
			JOIN campaign_members cm ON cm.campaign_id = tl.campaign_id
			WHERE tl.id = $1 AND cm.user_id = $2
		)
	`, listID, userID).Scan(&exists)
	return exists, err
}

func (s *Store) CreateCampaign(ctx context.Context, userID, name string, releaseDate *string) (*model.Campaign, error) {
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

	if err := s.populateTemplate(ctx, tx, campaign.ID, releaseDate); err != nil {
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
	// Single query with JOINs to avoid N+1 round trips to the database
	rows, err := s.pool.Query(ctx, `
		SELECT
			c.id, c.created_by, c.name, c.archived, c.created_at, c.updated_at,
			tl.id, tl.campaign_id, tl.name, tl.color, tl.position,
			tg.id, tg.task_list_id, tg.name, tg.position, tg.collapsed,
			t.id, t.task_group_id, t.name, t.description, t.status, t.due_date, t.position, t.created_at, t.updated_at,
			st.id, st.task_id, st.name, st.is_complete, st.position
		FROM campaigns c
		LEFT JOIN task_lists tl ON tl.campaign_id = c.id
		LEFT JOIN task_groups tg ON tg.task_list_id = tl.id
		LEFT JOIN tasks t ON t.task_group_id = tg.id
		LEFT JOIN subtasks st ON st.task_id = t.id
		WHERE c.id = $1
		ORDER BY tl.position, tg.position, t.position, st.position
	`, campaignID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var campaign *model.Campaign
	listMap := map[string]int{}    // list ID -> index in campaign.TaskLists
	groupMap := map[string]int{}   // group ID -> index in its parent list's TaskGroups
	taskMap := map[string]int{}    // task ID -> index in its parent group's Tasks
	subtaskSeen := map[string]bool{}

	for rows.Next() {
		var (
			cID, cCreatedBy, cName                     string
			cArchived                                  bool
			cCreatedAt, cUpdatedAt                     time.Time
			tlID, tlCampaignID, tlName, tlColor        *string
			tlPosition                                 *int
			tgID, tgTaskListID, tgName                 *string
			tgPosition                                 *int
			tgCollapsed                                *bool
			tID, tTaskGroupID, tName                   *string
			tDescription                               *json.RawMessage
			tStatus, tDueDate                          *string
			tPosition                                  *int
			tCreatedAt, tUpdatedAt                     *time.Time
			stID, stTaskID, stName                     *string
			stIsComplete                               *bool
			stPosition                                 *int
		)

		err := rows.Scan(
			&cID, &cCreatedBy, &cName, &cArchived, &cCreatedAt, &cUpdatedAt,
			&tlID, &tlCampaignID, &tlName, &tlColor, &tlPosition,
			&tgID, &tgTaskListID, &tgName, &tgPosition, &tgCollapsed,
			&tID, &tTaskGroupID, &tName, &tDescription, &tStatus, &tDueDate, &tPosition, &tCreatedAt, &tUpdatedAt,
			&stID, &stTaskID, &stName, &stIsComplete, &stPosition,
		)
		if err != nil {
			return nil, err
		}

		if campaign == nil {
			campaign = &model.Campaign{
				ID: cID, CreatedBy: cCreatedBy, Name: cName, Archived: cArchived,
				CreatedAt: cCreatedAt, UpdatedAt: cUpdatedAt,
			}
		}

		// Task list
		if tlID != nil {
			if _, ok := listMap[*tlID]; !ok {
				listMap[*tlID] = len(campaign.TaskLists)
				campaign.TaskLists = append(campaign.TaskLists, model.TaskList{
					ID: *tlID, CampaignID: *tlCampaignID, Name: *tlName, Color: *tlColor, Position: *tlPosition,
				})
			}
		}

		// Task group
		if tgID != nil && tlID != nil {
			if _, ok := groupMap[*tgID]; !ok {
				li := listMap[*tlID]
				groupMap[*tgID] = len(campaign.TaskLists[li].TaskGroups)
				collapsed := false
				if tgCollapsed != nil {
					collapsed = *tgCollapsed
				}
				campaign.TaskLists[li].TaskGroups = append(campaign.TaskLists[li].TaskGroups, model.TaskGroup{
					ID: *tgID, TaskListID: *tgTaskListID, Name: *tgName, Position: *tgPosition, Collapsed: collapsed,
				})
			}
		}

		// Task
		if tID != nil && tgID != nil && tlID != nil {
			if _, ok := taskMap[*tID]; !ok {
				li := listMap[*tlID]
				gi := groupMap[*tgID]
				task := model.Task{
					ID: *tID, TaskGroupID: *tTaskGroupID, Name: *tName, Description: tDescription,
					Status: *tStatus, Position: *tPosition,
				}
				if tDueDate != nil {
					task.DueDate = tDueDate
				}
				if tCreatedAt != nil {
					task.CreatedAt = *tCreatedAt
				}
				if tUpdatedAt != nil {
					task.UpdatedAt = *tUpdatedAt
				}
				taskMap[*tID] = len(campaign.TaskLists[li].TaskGroups[gi].Tasks)
				campaign.TaskLists[li].TaskGroups[gi].Tasks = append(campaign.TaskLists[li].TaskGroups[gi].Tasks, task)
			}
		}

		// Subtask
		if stID != nil && tID != nil && tgID != nil && tlID != nil {
			if !subtaskSeen[*stID] {
				subtaskSeen[*stID] = true
				li := listMap[*tlID]
				gi := groupMap[*tgID]
				ti := taskMap[*tID]
				isComplete := false
				if stIsComplete != nil {
					isComplete = *stIsComplete
				}
				campaign.TaskLists[li].TaskGroups[gi].Tasks[ti].Subtasks = append(
					campaign.TaskLists[li].TaskGroups[gi].Tasks[ti].Subtasks,
					model.Subtask{ID: *stID, TaskID: *stTaskID, Name: *stName, IsComplete: isComplete, Position: *stPosition},
				)
			}
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	if campaign == nil {
		return nil, fmt.Errorf("campaign not found")
	}
	return campaign, nil
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

func (s *Store) populateTemplate(ctx context.Context, tx pgx.Tx, campaignID string, releaseDate *string) error {
	tmpl := template.DefaultTemplate()

	// Parse release date if provided
	var relDate *time.Time
	if releaseDate != nil && *releaseDate != "" {
		t, err := time.Parse("2006-01-02", *releaseDate)
		if err == nil {
			relDate = &t
		}
	}

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
				// Calculate due date from release date + offset
				var dueDate *string
				if relDate != nil && task.DaysOffset != nil {
					d := relDate.AddDate(0, 0, *task.DaysOffset).Format("2006-01-02")
					dueDate = &d
				}

				var taskID string
				err := tx.QueryRow(ctx, `
					INSERT INTO tasks (task_group_id, name, due_date, position)
					VALUES ($1, $2, $3, $4)
					RETURNING id
				`, groupID, task.Name, dueDate, (taskPos+1)*100).Scan(&taskID)
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
