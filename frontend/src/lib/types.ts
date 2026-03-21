export interface User {
  id: string;
  email: string;
  name: string;
  avatar_url?: string;
  created_at: string;
}

export interface Campaign {
  id: string;
  created_by: string;
  name: string;
  archived: boolean;
  created_at: string;
  updated_at: string;
  task_lists?: TaskList[];
}

export interface TaskList {
  id: string;
  campaign_id: string;
  name: string;
  color: string;
  position: number;
  task_groups?: TaskGroup[];
}

export interface TaskGroup {
  id: string;
  task_list_id: string;
  name: string;
  position: number;
  collapsed: boolean;
  tasks?: Task[];
}

export interface Task {
  id: string;
  task_group_id: string;
  name: string;
  description?: Record<string, unknown>;
  status: "todo" | "in_progress" | "done";
  due_date?: string;
  position: number;
  created_at: string;
  updated_at: string;
  subtasks?: Subtask[];
}

export interface Subtask {
  id: string;
  task_id: string;
  name: string;
  is_complete: boolean;
  position: number;
}
