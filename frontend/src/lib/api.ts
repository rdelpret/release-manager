import type { Campaign, Task, Subtask, TemplateType } from "./types";

// In dev, Next.js rewrites proxy /api/* to Go backend.
// In production, Cloudflare Worker proxies /api/* to the container.
// Either way, use relative URLs.

async function fetchJSON<T>(url: string, options?: RequestInit): Promise<T> {
  const res = await fetch(url, {
    credentials: "include",
    headers: { "Content-Type": "application/json" },
    ...options,
  });
  if (!res.ok) {
    const body = await res.json().catch(() => ({ error: res.statusText }));
    throw new Error(body.error || `Request failed: ${res.status}`);
  }
  return res.json();
}

// Auth
export const getMe = () => fetchJSON<{ email: string; user_id: string }>("/api/me");
export const logout = () => fetchJSON<void>("/auth/logout", { method: "POST" });

// Campaigns
export const listCampaigns = () => fetchJSON<Campaign[]>("/api/campaigns");
export const createCampaign = (name: string, releaseDate?: string, templateType?: TemplateType) =>
  fetchJSON<Campaign>("/api/campaigns", {
    method: "POST",
    body: JSON.stringify({
      name,
      release_date: releaseDate || undefined,
      template_type: templateType || "single",
    }),
  });
export const getCampaign = (id: string) => fetchJSON<Campaign>(`/api/campaigns/${id}`);
export const duplicateCampaign = (id: string) =>
  fetchJSON<Campaign>(`/api/campaigns/${id}/duplicate`, { method: "POST" });
export const archiveCampaign = (id: string, archived: boolean) =>
  fetchJSON<void>(`/api/campaigns/${id}/archive`, { method: "PATCH", body: JSON.stringify({ archived }) });
export const setReleaseDate = (id: string, releaseDate: string, scheduleWeeks: number) =>
  fetchJSON<void>(`/api/campaigns/${id}/release-date`, {
    method: "PATCH",
    body: JSON.stringify({ release_date: releaseDate, schedule_weeks: scheduleWeeks }),
  });
export const deleteCampaign = (id: string) =>
  fetchJSON<void>(`/api/campaigns/${id}`, { method: "DELETE" });

// Tasks
export const createTask = (groupId: string, name: string) =>
  fetchJSON<Task>(`/api/task-groups/${groupId}/tasks`, { method: "POST", body: JSON.stringify({ name }) });
export const updateTask = (id: string, updates: Partial<Task>) =>
  fetchJSON<Task>(`/api/tasks/${id}`, { method: "PATCH", body: JSON.stringify(updates) });
export const deleteTask = (id: string) =>
  fetchJSON<void>(`/api/tasks/${id}`, { method: "DELETE" });
export const reorderTask = (id: string, targetGroupId: string, position: number) =>
  fetchJSON<void>(`/api/tasks/${id}/reorder`, {
    method: "PATCH",
    body: JSON.stringify({ target_group_id: targetGroupId, position }),
  });

// Reorder
export const reorderTaskList = (id: string, position: number) =>
  fetchJSON<void>(`/api/task-lists/${id}/reorder`, { method: "PATCH", body: JSON.stringify({ position }) });
export const reorderTaskGroup = (id: string, position: number) =>
  fetchJSON<void>(`/api/task-groups/${id}/reorder`, { method: "PATCH", body: JSON.stringify({ position }) });

// Subtasks
export const createSubtask = (taskId: string, name: string) =>
  fetchJSON<Subtask>(`/api/tasks/${taskId}/subtasks`, { method: "POST", body: JSON.stringify({ name }) });
export const updateSubtask = (id: string, updates: { name?: string; is_complete?: boolean }) =>
  fetchJSON<Subtask>(`/api/subtasks/${id}`, { method: "PATCH", body: JSON.stringify(updates) });
export const deleteSubtask = (id: string) =>
  fetchJSON<void>(`/api/subtasks/${id}`, { method: "DELETE" });
