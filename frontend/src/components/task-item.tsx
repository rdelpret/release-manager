"use client";

import type { Task } from "@/lib/types";
import { Circle, CircleDot, CheckCircle2, GripVertical } from "lucide-react";
import { useSortable } from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";

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
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging,
  } = useSortable({
    id: task.id,
    data: { groupId: task.task_group_id, position: task.position },
  });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
    opacity: isDragging ? 0.5 : 1,
  };

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
      ref={setNodeRef}
      style={style}
      onClick={() => onSelect(task)}
      className="flex items-center gap-3 rounded-lg px-3 py-2.5 cursor-pointer transition-smooth hover:bg-white/[0.03] group"
    >
      <button
        {...attributes}
        {...listeners}
        className="opacity-0 group-hover:opacity-100 cursor-grab text-text-muted"
        onClick={(e) => e.stopPropagation()}
      >
        <GripVertical className="h-3.5 w-3.5" />
      </button>
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
