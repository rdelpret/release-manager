"use client";

import { useState } from "react";
import type { Campaign, User } from "@/lib/types";

interface CampaignOverviewProps {
  campaign: Campaign;
  users?: User[];
}

export function CampaignOverview({ campaign, users }: CampaignOverviewProps) {
  const [now] = useState(() => Date.now());
  const today = new Date(now).toISOString().slice(0, 10);

  const allTasks = (campaign.task_lists ?? []).flatMap((l) =>
    (l.task_groups ?? []).flatMap((g) => (g.tasks ?? []).map((t) => ({ ...t, listName: l.name, listColor: l.color })))
  );

  const totalTasks = allTasks.length;
  const doneTasks = allTasks.filter((t) => t.status === "done").length;
  const overdueTasks = allTasks.filter((t) => t.status !== "done" && t.due_date && t.due_date < today);
  const upcomingTasks = allTasks
    .filter((t) => t.status !== "done" && t.due_date && t.due_date >= today)
    .sort((a, b) => (a.due_date ?? "").localeCompare(b.due_date ?? ""))
    .slice(0, 10);

  // Per-list stats
  const listStats = (campaign.task_lists ?? []).map((list) => {
    const tasks = (list.task_groups ?? []).flatMap((g) => g.tasks ?? []);
    const total = tasks.length;
    const done = tasks.filter((t) => t.status === "done").length;
    const overdue = tasks.filter((t) => t.status !== "done" && t.due_date && t.due_date < today).length;
    return { name: list.name, color: list.color, total, done, overdue };
  });

  // Per-person stats
  const personMap = new Map<string, { name: string; total: number; done: number; overdue: number }>();
  for (const task of allTasks) {
    if (!task.assigned_to) continue;
    const user = users?.find((u) => u.id === task.assigned_to);
    const key = task.assigned_to;
    if (!personMap.has(key)) {
      personMap.set(key, { name: user?.name ?? "Unknown", total: 0, done: 0, overdue: 0 });
    }
    const p = personMap.get(key)!;
    p.total++;
    if (task.status === "done") p.done++;
    else if (task.due_date && task.due_date < today) p.overdue++;
  }
  const personStats = Array.from(personMap.values()).sort((a, b) => b.total - a.total);

  const unassignedCount = allTasks.filter((t) => !t.assigned_to && t.status !== "done").length;

  return (
    <div className="mt-4 space-y-6">
      {/* Overall progress */}
      <div className="bg-bg-surface rounded-xl p-5">
        <h3 className="text-sm font-semibold text-text-muted uppercase tracking-wider mb-3">Overall Progress</h3>
        <div className="flex items-center gap-4">
          <div className="flex-1">
            <div className="h-2.5 rounded-full bg-white/[0.06] overflow-hidden">
              <div
                className="h-full rounded-full bg-accent transition-all"
                style={{ width: `${totalTasks > 0 ? (doneTasks / totalTasks) * 100 : 0}%` }}
              />
            </div>
          </div>
          <span className="text-sm font-medium text-text-primary">
            {doneTasks}/{totalTasks}
          </span>
          <span className="text-xs text-text-muted">
            ({totalTasks > 0 ? Math.round((doneTasks / totalTasks) * 100) : 0}%)
          </span>
        </div>
        {overdueTasks.length > 0 && (
          <p className="mt-2 text-xs text-red-400">{overdueTasks.length} overdue task{overdueTasks.length !== 1 ? "s" : ""}</p>
        )}
      </div>

      {/* Per-list progress */}
      <div className="bg-bg-surface rounded-xl p-5">
        <h3 className="text-sm font-semibold text-text-muted uppercase tracking-wider mb-3">Progress by List</h3>
        <div className="space-y-3">
          {listStats.map((list) => (
            <div key={list.name}>
              <div className="flex items-center justify-between mb-1">
                <div className="flex items-center gap-2">
                  <span className="inline-block h-2 w-2 rounded-full" style={{ backgroundColor: list.color }} />
                  <span className="text-sm text-text-primary">{list.name}</span>
                </div>
                <div className="flex items-center gap-2">
                  <span className="text-xs text-text-muted">{list.done}/{list.total}</span>
                  {list.overdue > 0 && (
                    <span className="text-[10px] font-medium text-red-400">{list.overdue} overdue</span>
                  )}
                </div>
              </div>
              <div className="h-1.5 rounded-full bg-white/[0.06] overflow-hidden">
                <div
                  className="h-full rounded-full transition-all"
                  style={{
                    width: `${list.total > 0 ? (list.done / list.total) * 100 : 0}%`,
                    backgroundColor: list.color,
                  }}
                />
              </div>
            </div>
          ))}
        </div>
      </div>

      <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
        {/* Upcoming due dates */}
        <div className="bg-bg-surface rounded-xl p-5">
          <h3 className="text-sm font-semibold text-text-muted uppercase tracking-wider mb-3">Upcoming Due Dates</h3>
          {upcomingTasks.length === 0 ? (
            <p className="text-xs text-text-muted">No upcoming tasks with due dates</p>
          ) : (
            <div className="space-y-2">
              {upcomingTasks.map((task) => (
                <div key={task.id} className="flex items-center justify-between gap-2">
                  <span className="text-sm text-text-primary truncate flex-1">{task.name}</span>
                  <span className="text-xs text-text-muted whitespace-nowrap">
                    {new Date(task.due_date!).toLocaleDateString("en-US", { month: "short", day: "numeric" })}
                  </span>
                </div>
              ))}
            </div>
          )}
        </div>

        {/* Per-person breakdown */}
        <div className="bg-bg-surface rounded-xl p-5">
          <h3 className="text-sm font-semibold text-text-muted uppercase tracking-wider mb-3">By Person</h3>
          {personStats.length === 0 ? (
            <p className="text-xs text-text-muted">No tasks assigned yet</p>
          ) : (
            <div className="space-y-2">
              {personStats.map((person) => (
                <div key={person.name} className="flex items-center justify-between gap-2">
                  <span className="text-sm text-text-primary">{person.name}</span>
                  <div className="flex items-center gap-2">
                    <span className="text-xs text-text-muted">{person.done}/{person.total}</span>
                    {person.overdue > 0 && (
                      <span className="text-[10px] font-medium text-red-400">{person.overdue} late</span>
                    )}
                  </div>
                </div>
              ))}
              {unassignedCount > 0 && (
                <div className="flex items-center justify-between gap-2 pt-1 border-t border-border">
                  <span className="text-sm text-text-muted">Unassigned</span>
                  <span className="text-xs text-text-muted">{unassignedCount} remaining</span>
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Overdue tasks list */}
      {overdueTasks.length > 0 && (
        <div className="bg-bg-surface rounded-xl p-5">
          <h3 className="text-sm font-semibold text-red-400 uppercase tracking-wider mb-3">Overdue Tasks</h3>
          <div className="space-y-2">
            {overdueTasks.map((task) => (
              <div key={task.id} className="flex items-center justify-between gap-2">
                <div className="flex items-center gap-2 flex-1 min-w-0">
                  <span className="inline-block h-2 w-2 rounded-full flex-shrink-0" style={{ backgroundColor: task.listColor }} />
                  <span className="text-sm text-text-primary truncate">{task.name}</span>
                </div>
                <span className="text-xs text-red-400 whitespace-nowrap">
                  {new Date(task.due_date!).toLocaleDateString("en-US", { month: "short", day: "numeric" })}
                </span>
              </div>
            ))}
          </div>
        </div>
      )}
    </div>
  );
}
