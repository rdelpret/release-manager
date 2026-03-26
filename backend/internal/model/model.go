package model

import (
	"encoding/json"
	"time"
)

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	AvatarURL *string   `json:"avatar_url,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

type Campaign struct {
	ID            string     `json:"id"`
	CreatedBy     string     `json:"created_by"`
	Name          string     `json:"name"`
	Archived      bool       `json:"archived"`
	TemplateType  string     `json:"template_type"`
	ReleaseDate   *string    `json:"release_date,omitempty"`
	ScheduleWeeks int        `json:"schedule_weeks"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	TaskLists     []TaskList `json:"task_lists,omitempty"`
}

type CampaignMember struct {
	CampaignID string `json:"campaign_id"`
	UserID     string `json:"user_id"`
	Role       string `json:"role"`
}

type TaskList struct {
	ID         string      `json:"id"`
	CampaignID string      `json:"campaign_id"`
	Name       string      `json:"name"`
	Color      string      `json:"color"`
	Position   int         `json:"position"`
	TaskGroups []TaskGroup `json:"task_groups,omitempty"`
}

type TaskGroup struct {
	ID         string `json:"id"`
	TaskListID string `json:"task_list_id"`
	Name       string `json:"name"`
	Position   int    `json:"position"`
	Collapsed  bool   `json:"collapsed"`
	Tasks      []Task `json:"tasks,omitempty"`
}

type Task struct {
	ID          string           `json:"id"`
	TaskGroupID string           `json:"task_group_id"`
	Name        string           `json:"name"`
	Description *json.RawMessage `json:"description,omitempty"`
	Status      string           `json:"status"`
	DueDate     *string          `json:"due_date,omitempty"`
	Position    int              `json:"position"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`
	Subtasks    []Subtask        `json:"subtasks,omitempty"`
}

type Subtask struct {
	ID         string `json:"id"`
	TaskID     string `json:"task_id"`
	Name       string `json:"name"`
	IsComplete bool   `json:"is_complete"`
	Position   int    `json:"position"`
}
