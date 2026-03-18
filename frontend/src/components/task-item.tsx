"use client";

import type { Task } from "@/lib/types";
import { Circle, CircleDot, CheckCircle2 } from "lucide-react";

const statusConfig = {
  todo: { icon: Circle, color: "text-text-muted" },
  in_progress: { icon: CircleDot, color: "text-yellow-400" },
  done: { icon: CheckCircle2, color: "text-green-400" },
};

interface TaskItemProps {
  task: Task;
  onSelect: (task: Task) => void;
  onStatusChange: (taskId: string, status: Task["status"]) => void;
}

export function TaskItem({ task, onSelect, onStatusChange }: TaskItemProps) {
  const { icon: StatusIcon, color } = statusConfig[task.status];

  const cycleStatus = (e: React.MouseEvent) => {
    e.stopPropagation();
    const next: Record<string, Task["status"]> = {
      todo: "in_progress",
      in_progress: "done",
      done: "todo",
    };
    onStatusChange(task.id, next[task.status]);
  };

  return (
    <div
      onClick={() => onSelect(task)}
      className="flex items-center gap-3 rounded-lg px-3 py-2.5 cursor-pointer transition-smooth hover:bg-white/[0.03] group"
    >
      <button onClick={cycleStatus} className={`${color} transition-smooth`}>
        <StatusIcon className="h-4 w-4" />
      </button>
      <span className={`flex-1 text-sm ${task.status === "done" ? "text-text-muted line-through" : "text-text-primary"}`}>
        {task.name}
      </span>
      {task.due_date && (
        <span className="text-xs text-text-muted">
          {new Date(task.due_date).toLocaleDateString("en-US", { month: "short", day: "numeric" })}
        </span>
      )}
      {task.subtasks && task.subtasks.length > 0 && (
        <span className="text-xs text-text-muted">
          {task.subtasks.filter((s) => s.is_complete).length}/{task.subtasks.length}
        </span>
      )}
    </div>
  );
}
