package store

import (
	"context"
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

func (s *Store) CreateCampaign(ctx context.Context, userID, name string, releaseDate *string, templateType string) (*model.Campaign, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	scheduleWeeks := template.DefaultScheduleWeeks(templateType)

	var campaign model.Campaign
	err = tx.QueryRow(ctx, `
		INSERT INTO campaigns (created_by, name, template_type, release_date, schedule_weeks)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_by, name, archived, template_type, release_date, schedule_weeks, created_at, updated_at
	`, userID, name, templateType, releaseDate, scheduleWeeks).Scan(&campaign.ID, &campaign.CreatedBy, &campaign.Name,
		&campaign.Archived, &campaign.TemplateType, &campaign.ReleaseDate, &campaign.ScheduleWeeks, &campaign.CreatedAt, &campaign.UpdatedAt)
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

	if err := s.populateTemplate(ctx, tx, campaign.ID, templateType, releaseDate); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &campaign, nil
}

func (s *Store) ListCampaigns(ctx context.Context, userID string) ([]model.Campaign, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT c.id, c.created_by, c.name, c.archived, c.template_type, c.release_date, c.schedule_weeks, c.created_at, c.updated_at,
			COALESCE(cs.total, 0),
			COALESCE(cs.done, 0),
			COALESCE(cs.overdue, 0)
		FROM campaigns c
		JOIN campaign_members cm ON cm.campaign_id = c.id
		LEFT JOIN LATERAL (
			SELECT
				COUNT(t.id)::int AS total,
				COUNT(t.id) FILTER (WHERE t.status = 'done')::int AS done,
				COUNT(t.id) FILTER (WHERE t.status != 'done' AND t.due_date < CURRENT_DATE)::int AS overdue
			FROM task_lists tl
			JOIN task_groups tg ON tg.task_list_id = tl.id
			JOIN tasks t ON t.task_group_id = tg.id
			WHERE tl.campaign_id = c.id
		) cs ON true
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
		if err := rows.Scan(&c.ID, &c.CreatedBy, &c.Name, &c.Archived, &c.TemplateType,
			&c.ReleaseDate, &c.ScheduleWeeks, &c.CreatedAt, &c.UpdatedAt,
			&c.TotalTasks, &c.DoneTasks, &c.OverdueTasks); err != nil {
			return nil, err
		}
		campaigns = append(campaigns, c)
	}
	return campaigns, rows.Err()
}

func (s *Store) GetFullCampaign(ctx context.Context, campaignID string) (*model.Campaign, error) {
	// Fetch campaign
	var campaign model.Campaign
	err := s.pool.QueryRow(ctx, `
		SELECT id, created_by, name, archived, template_type, release_date, schedule_weeks, created_at, updated_at
		FROM campaigns WHERE id = $1
	`, campaignID).Scan(
		&campaign.ID, &campaign.CreatedBy, &campaign.Name, &campaign.Archived,
		&campaign.TemplateType, &campaign.ReleaseDate, &campaign.ScheduleWeeks,
		&campaign.CreatedAt, &campaign.UpdatedAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, fmt.Errorf("campaign not found")
		}
		return nil, err
	}

	// Fetch task lists
	listRows, err := s.pool.Query(ctx, `
		SELECT id, campaign_id, name, color, position
		FROM task_lists WHERE campaign_id = $1 ORDER BY position
	`, campaignID)
	if err != nil {
		return nil, err
	}
	defer listRows.Close()

	listMap := map[string]int{} // list ID -> index
	for listRows.Next() {
		var tl model.TaskList
		if err := listRows.Scan(&tl.ID, &tl.CampaignID, &tl.Name, &tl.Color, &tl.Position); err != nil {
			return nil, err
		}
		listMap[tl.ID] = len(campaign.TaskLists)
		campaign.TaskLists = append(campaign.TaskLists, tl)
	}
	if err := listRows.Err(); err != nil {
		return nil, err
	}
	listRows.Close()

	if len(campaign.TaskLists) == 0 {
		return &campaign, nil
	}

	// Fetch task groups
	groupRows, err := s.pool.Query(ctx, `
		SELECT tg.id, tg.task_list_id, tg.name, tg.position, tg.collapsed
		FROM task_groups tg
		JOIN task_lists tl ON tl.id = tg.task_list_id
		WHERE tl.campaign_id = $1
		ORDER BY tg.position
	`, campaignID)
	if err != nil {
		return nil, err
	}
	defer groupRows.Close()

	groupMap := map[string]int{}   // group ID -> index in parent list
	groupList := map[string]string{} // group ID -> list ID
	for groupRows.Next() {
		var tg model.TaskGroup
		if err := groupRows.Scan(&tg.ID, &tg.TaskListID, &tg.Name, &tg.Position, &tg.Collapsed); err != nil {
			return nil, err
		}
		li := listMap[tg.TaskListID]
		groupMap[tg.ID] = len(campaign.TaskLists[li].TaskGroups)
		groupList[tg.ID] = tg.TaskListID
		campaign.TaskLists[li].TaskGroups = append(campaign.TaskLists[li].TaskGroups, tg)
	}
	if err := groupRows.Err(); err != nil {
		return nil, err
	}
	groupRows.Close()

	// Fetch tasks
	taskRows, err := s.pool.Query(ctx, `
		SELECT t.id, t.task_group_id, t.name, t.description, t.status, t.due_date, t.assigned_to, t.position, t.created_at, t.updated_at
		FROM tasks t
		JOIN task_groups tg ON tg.id = t.task_group_id
		JOIN task_lists tl ON tl.id = tg.task_list_id
		WHERE tl.campaign_id = $1
		ORDER BY t.position
	`, campaignID)
	if err != nil {
		return nil, err
	}
	defer taskRows.Close()

	taskMap := map[string]int{}    // task ID -> index in parent group
	taskGroup := map[string]string{} // task ID -> group ID
	for taskRows.Next() {
		var t model.Task
		if err := taskRows.Scan(&t.ID, &t.TaskGroupID, &t.Name, &t.Description, &t.Status, &t.DueDate, &t.AssignedTo, &t.Position, &t.CreatedAt, &t.UpdatedAt); err != nil {
			return nil, err
		}
		listID := groupList[t.TaskGroupID]
		li := listMap[listID]
		gi := groupMap[t.TaskGroupID]
		taskMap[t.ID] = len(campaign.TaskLists[li].TaskGroups[gi].Tasks)
		taskGroup[t.ID] = t.TaskGroupID
		campaign.TaskLists[li].TaskGroups[gi].Tasks = append(campaign.TaskLists[li].TaskGroups[gi].Tasks, t)
	}
	if err := taskRows.Err(); err != nil {
		return nil, err
	}
	taskRows.Close()

	// Fetch subtasks
	subtaskRows, err := s.pool.Query(ctx, `
		SELECT st.id, st.task_id, st.name, st.is_complete, st.position
		FROM subtasks st
		JOIN tasks t ON t.id = st.task_id
		JOIN task_groups tg ON tg.id = t.task_group_id
		JOIN task_lists tl ON tl.id = tg.task_list_id
		WHERE tl.campaign_id = $1
		ORDER BY st.position
	`, campaignID)
	if err != nil {
		return nil, err
	}
	defer subtaskRows.Close()

	for subtaskRows.Next() {
		var st model.Subtask
		if err := subtaskRows.Scan(&st.ID, &st.TaskID, &st.Name, &st.IsComplete, &st.Position); err != nil {
			return nil, err
		}
		gID := taskGroup[st.TaskID]
		listID := groupList[gID]
		li := listMap[listID]
		gi := groupMap[gID]
		ti := taskMap[st.TaskID]
		campaign.TaskLists[li].TaskGroups[gi].Tasks[ti].Subtasks = append(
			campaign.TaskLists[li].TaskGroups[gi].Tasks[ti].Subtasks, st,
		)
	}
	if err := subtaskRows.Err(); err != nil {
		return nil, err
	}

	return &campaign, nil
}

func (s *Store) ArchiveCampaign(ctx context.Context, campaignID string, archived bool) error {
	_, err := s.pool.Exec(ctx, `
		UPDATE campaigns SET archived = $2, updated_at = $3
		WHERE id = $1
	`, campaignID, archived, time.Now())
	return err
}

// SetReleaseDate updates the release date and schedule, then recalculates all task due dates.
func (s *Store) SetReleaseDate(ctx context.Context, campaignID, releaseDate string, scheduleWeeks int) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	// Get the campaign's template type
	var templateType string
	err = tx.QueryRow(ctx, `SELECT template_type FROM campaigns WHERE id = $1`, campaignID).Scan(&templateType)
	if err != nil {
		return fmt.Errorf("fetching campaign template type: %w", err)
	}

	// Update campaign
	_, err = tx.Exec(ctx, `
		UPDATE campaigns SET release_date = $2, schedule_weeks = $3, updated_at = now()
		WHERE id = $1
	`, campaignID, releaseDate, scheduleWeeks)
	if err != nil {
		return err
	}

	// Parse release date
	relDate, err := time.Parse("2006-01-02", releaseDate)
	if err != nil {
		return fmt.Errorf("invalid release date: %w", err)
	}

	// Scale factor: 4-week schedule compresses pre-release by 50%
	scale := 1.0
	if scheduleWeeks == 4 {
		scale = 0.5
	}

	// Recalculate due dates for all tasks using the correct template offsets.
	// Collect all name→date pairs and batch into a single UPDATE.
	tmpl := template.GetTemplate(templateType)
	var taskNames []string
	var dueDates []string
	for _, list := range tmpl {
		for _, group := range list.Groups {
			for _, task := range group.Tasks {
				if task.DaysOffset == nil {
					continue
				}
				offset := *task.DaysOffset
				// Scale pre-release days (negative offsets) for shorter schedules
				if offset < 0 {
					offset = int(float64(offset) * scale)
				}
				taskNames = append(taskNames, task.Name)
				dueDates = append(dueDates, relDate.AddDate(0, 0, offset).Format("2006-01-02"))
			}
		}
	}

	if len(taskNames) > 0 {
		_, err = tx.Exec(ctx, `
			UPDATE tasks t SET due_date = v.due_date::date
			FROM unnest($1::text[], $2::text[]) AS v(task_name, due_date)
			WHERE t.name = v.task_name
			AND t.task_group_id IN (
				SELECT tg.id FROM task_groups tg
				JOIN task_lists tl ON tl.id = tg.task_list_id
				WHERE tl.campaign_id = $3
			)
		`, taskNames, dueDates, campaignID)
		if err != nil {
			return fmt.Errorf("batch updating due dates: %w", err)
		}
	}

	return tx.Commit(ctx)
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

	var sourceName, sourceTemplateType string
	err = tx.QueryRow(ctx, `SELECT name, template_type FROM campaigns WHERE id = $1`, sourceCampaignID).Scan(&sourceName, &sourceTemplateType)
	if err != nil {
		return nil, err
	}

	var campaign model.Campaign
	err = tx.QueryRow(ctx, `
		INSERT INTO campaigns (created_by, name, template_type)
		VALUES ($1, $2, $3)
		RETURNING id, created_by, name, archived, template_type, release_date, schedule_weeks, created_at, updated_at
	`, userID, sourceName+" (Copy)", sourceTemplateType).Scan(&campaign.ID, &campaign.CreatedBy, &campaign.Name,
		&campaign.Archived, &campaign.TemplateType, &campaign.ReleaseDate, &campaign.ScheduleWeeks, &campaign.CreatedAt, &campaign.UpdatedAt)
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

	// Duplicate entire hierarchy in 4 batched queries instead of N per entity.
	// Each level uses INSERT...SELECT with a mapping CTE to link old→new IDs.

	// 1. Duplicate task lists
	listRows, err := tx.Query(ctx, `
		WITH inserted AS (
			INSERT INTO task_lists (campaign_id, name, color, position)
			SELECT $1, name, color, position
			FROM task_lists WHERE campaign_id = $2
			ORDER BY position
			RETURNING id, name, color, position
		)
		SELECT ol.id AS old_id, ins.id AS new_id
		FROM (SELECT id, name, position, ROW_NUMBER() OVER (ORDER BY position) rn FROM task_lists WHERE campaign_id = $2) ol
		JOIN (SELECT id, name, position, ROW_NUMBER() OVER (ORDER BY position) rn FROM inserted) ins ON ol.rn = ins.rn
	`, campaign.ID, sourceCampaignID)
	if err != nil {
		return nil, err
	}
	var oldListIDs, newListIDs []string
	for listRows.Next() {
		var oldID, newID string
		if err := listRows.Scan(&oldID, &newID); err != nil {
			listRows.Close()
			return nil, err
		}
		oldListIDs = append(oldListIDs, oldID)
		newListIDs = append(newListIDs, newID)
	}
	listRows.Close()
	if err := listRows.Err(); err != nil {
		return nil, err
	}

	if len(oldListIDs) == 0 {
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		return &campaign, nil
	}

	// 2. Duplicate task groups (batch across all lists)
	groupRows, err := tx.Query(ctx, `
		WITH list_map AS (
			SELECT unnest($1::uuid[]) AS old_id, unnest($2::uuid[]) AS new_id
		),
		inserted AS (
			INSERT INTO task_groups (task_list_id, name, position, collapsed)
			SELECT lm.new_id, tg.name, tg.position, tg.collapsed
			FROM task_groups tg
			JOIN list_map lm ON lm.old_id = tg.task_list_id
			ORDER BY tg.task_list_id, tg.position
			RETURNING id, task_list_id, name, position
		)
		SELECT og.id AS old_id, ins.id AS new_id
		FROM (
			SELECT tg.id, tg.task_list_id, tg.name, tg.position,
				ROW_NUMBER() OVER (ORDER BY tg.task_list_id, tg.position) rn
			FROM task_groups tg
			JOIN list_map lm ON lm.old_id = tg.task_list_id
		) og
		JOIN (
			SELECT ins.id, ins.task_list_id, ins.name, ins.position,
				ROW_NUMBER() OVER (ORDER BY ins.task_list_id, ins.position) rn
			FROM inserted ins
		) ins ON og.rn = ins.rn
	`, oldListIDs, newListIDs)
	if err != nil {
		return nil, err
	}
	var oldGroupIDs, newGroupIDs []string
	for groupRows.Next() {
		var oldID, newID string
		if err := groupRows.Scan(&oldID, &newID); err != nil {
			groupRows.Close()
			return nil, err
		}
		oldGroupIDs = append(oldGroupIDs, oldID)
		newGroupIDs = append(newGroupIDs, newID)
	}
	groupRows.Close()
	if err := groupRows.Err(); err != nil {
		return nil, err
	}

	if len(oldGroupIDs) == 0 {
		if err := tx.Commit(ctx); err != nil {
			return nil, err
		}
		return &campaign, nil
	}

	// 3. Duplicate tasks (batch across all groups)
	taskRows, err := tx.Query(ctx, `
		WITH group_map AS (
			SELECT unnest($1::uuid[]) AS old_id, unnest($2::uuid[]) AS new_id
		),
		inserted AS (
			INSERT INTO tasks (task_group_id, name, description, status, due_date, position)
			SELECT gm.new_id, t.name, t.description, 'todo', NULL, t.position
			FROM tasks t
			JOIN group_map gm ON gm.old_id = t.task_group_id
			ORDER BY t.task_group_id, t.position
			RETURNING id, task_group_id, name, position
		)
		SELECT ot.id AS old_id, ins.id AS new_id
		FROM (
			SELECT t.id, t.task_group_id, t.name, t.position,
				ROW_NUMBER() OVER (ORDER BY t.task_group_id, t.position) rn
			FROM tasks t
			JOIN group_map gm ON gm.old_id = t.task_group_id
		) ot
		JOIN (
			SELECT ins.id, ins.task_group_id, ins.name, ins.position,
				ROW_NUMBER() OVER (ORDER BY ins.task_group_id, ins.position) rn
			FROM inserted ins
		) ins ON ot.rn = ins.rn
	`, oldGroupIDs, newGroupIDs)
	if err != nil {
		return nil, err
	}
	var oldTaskIDs, newTaskIDs []string
	for taskRows.Next() {
		var oldID, newID string
		if err := taskRows.Scan(&oldID, &newID); err != nil {
			taskRows.Close()
			return nil, err
		}
		oldTaskIDs = append(oldTaskIDs, oldID)
		newTaskIDs = append(newTaskIDs, newID)
	}
	taskRows.Close()
	if err := taskRows.Err(); err != nil {
		return nil, err
	}

	// 4. Duplicate subtasks (single batch for all tasks)
	if len(oldTaskIDs) > 0 {
		_, err = tx.Exec(ctx, `
			WITH task_map AS (
				SELECT unnest($1::uuid[]) AS old_id, unnest($2::uuid[]) AS new_id
			)
			INSERT INTO subtasks (task_id, name, is_complete, position)
			SELECT tm.new_id, st.name, false, st.position
			FROM subtasks st
			JOIN task_map tm ON tm.old_id = st.task_id
		`, oldTaskIDs, newTaskIDs)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}
	return &campaign, nil
}

func (s *Store) populateTemplate(ctx context.Context, tx pgx.Tx, campaignID, templateType string, releaseDate *string) error {
	tmpl := template.GetTemplate(templateType)

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
