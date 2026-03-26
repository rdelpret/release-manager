"use client";

import { useState } from "react";
import type { Task, TaskGroup as TaskGroupType, User } from "@/lib/types";
import { TaskItem } from "./task-item";
import { ChevronDown, ChevronRight, Plus } from "lucide-react";
import { SortableContext, verticalListSortingStrategy } from "@dnd-kit/sortable";
import { createTask } from "@/lib/api";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";

interface TaskGroupProps {
  group: TaskGroupType;
  campaignId: string;
  hideDone: boolean;
  users?: User[];
  onSelectTask: (task: Task) => void;
  onStatusChange: (taskId: string, status: Task["status"]) => void;
}

export function TaskGroup({ group, campaignId, hideDone, users, onSelectTask, onStatusChange }: TaskGroupProps) {
  const [collapsed, setCollapsed] = useState(group.collapsed);
  const [adding, setAdding] = useState(false);
  const [newTaskName, setNewTaskName] = useState("");
  const queryClient = useQueryClient();

  const visibleTasks = hideDone
    ? (group.tasks ?? []).filter((t) => t.status !== "done")
    : (group.tasks ?? []);

  const handleAddTask = async () => {
    if (!newTaskName.trim()) return;
    try {
      await createTask(group.id, newTaskName.trim());
      setNewTaskName("");
      setAdding(false);
      queryClient.invalidateQueries({ queryKey: ["campaign", campaignId] });
    } catch (err: any) {
      toast.error(err.message);
    }
  };

  return (
    <div className="mb-4">
      <button
        onClick={() => setCollapsed(!collapsed)}
        className="flex items-center gap-2 text-xs font-semibold text-text-muted uppercase tracking-wider mb-2 hover:text-text-primary transition-smooth"
      >
        {collapsed ? <ChevronRight className="h-3.5 w-3.5" /> : <ChevronDown className="h-3.5 w-3.5" />}
        {group.name}
        <span className="text-text-muted font-normal">({visibleTasks.length})</span>
      </button>

      {!collapsed && (
        <div className="space-y-0.5">
          <SortableContext
            items={visibleTasks.map((t) => t.id)}
            strategy={verticalListSortingStrategy}
          >
            {visibleTasks.map((task) => (
              <TaskItem
                key={task.id}
                task={task}
                users={users}
                onSelect={onSelectTask}
                onStatusChange={onStatusChange}
              />
            ))}
          </SortableContext>

          {adding ? (
            <div className="flex gap-2 px-3 py-2">
              <input
                autoFocus
                autoComplete="off"
                value={newTaskName}
                onChange={(e) => setNewTaskName(e.target.value)}
                onKeyDown={(e) => {
                  if (e.key === "Enter") handleAddTask();
                  if (e.key === "Escape") setAdding(false);
                }}
                placeholder="Task name..."
                className="flex-1 bg-transparent border-b border-border text-sm text-text-primary placeholder:text-text-muted focus:outline-none focus:border-accent"
              />
            </div>
          ) : (
            <button
              onClick={() => setAdding(true)}
              className="flex items-center gap-2 px-3 py-1.5 text-xs text-text-muted hover:text-accent transition-smooth"
            >
              <Plus className="h-3 w-3" />
              Add task
            </button>
          )}
        </div>
      )}
    </div>
  );
}
