"use client";

import type { Subtask } from "@/lib/types";
import { Trash2 } from "lucide-react";
import { updateSubtask, deleteSubtask } from "@/lib/api";
import { toast } from "sonner";

interface SubtaskItemProps {
  subtask: Subtask;
  onUpdate: () => void;
}

export function SubtaskItem({ subtask, onUpdate }: SubtaskItemProps) {
  const handleToggle = async () => {
    try {
      await updateSubtask(subtask.id, { is_complete: !subtask.is_complete });
      onUpdate();
    } catch (err: any) {
      toast.error(err.message);
    }
  };

  const handleDelete = async () => {
    try {
      await deleteSubtask(subtask.id);
      onUpdate();
    } catch (err: any) {
      toast.error(err.message);
    }
  };

  return (
    <div className="flex items-center gap-3 group py-1">
      <input
        type="checkbox"
        checked={subtask.is_complete}
        onChange={handleToggle}
        className="h-4 w-4 rounded border-border accent-accent"
      />
      <span className={`flex-1 text-sm ${subtask.is_complete ? "text-text-muted line-through" : "text-text-primary"}`}>
        {subtask.name}
      </span>
      <button
        onClick={handleDelete}
        className="opacity-0 group-hover:opacity-100 text-text-muted hover:text-destructive transition-smooth"
      >
        <Trash2 className="h-3.5 w-3.5" />
      </button>
    </div>
  );
}
