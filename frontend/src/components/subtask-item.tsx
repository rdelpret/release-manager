"use client";

import { useState } from "react";
import type { Subtask } from "@/lib/types";
import { Trash2 } from "lucide-react";
import { updateSubtask, deleteSubtask } from "@/lib/api";
import { toast } from "sonner";

interface SubtaskItemProps {
  subtask: Subtask;
  onUpdate: () => void;
}

export function SubtaskItem({ subtask, onUpdate }: SubtaskItemProps) {
  const [optimisticComplete, setOptimisticComplete] = useState(subtask.is_complete);
  const [deleting, setDeleting] = useState(false);

  // Sync optimistic state when prop changes (after refetch)
  if (optimisticComplete !== subtask.is_complete && !deleting) {
    setOptimisticComplete(subtask.is_complete);
  }

  const handleToggle = async () => {
    const newVal = !optimisticComplete;
    setOptimisticComplete(newVal); // optimistic
    try {
      await updateSubtask(subtask.id, { is_complete: newVal });
      onUpdate();
    } catch (err: any) {
      setOptimisticComplete(!newVal); // rollback
      toast.error(err.message);
    }
  };

  const handleDelete = async () => {
    setDeleting(true); // hide immediately
    try {
      await deleteSubtask(subtask.id);
      onUpdate();
    } catch (err: any) {
      setDeleting(false); // rollback
      toast.error(err.message);
    }
  };

  if (deleting) return null;

  return (
    <div className="flex items-center gap-3 group py-1">
      <input
        type="checkbox"
        checked={optimisticComplete}
        onChange={handleToggle}
        className="h-4 w-4 rounded border-border accent-accent"
      />
      <span className={`flex-1 text-sm ${optimisticComplete ? "text-text-muted line-through" : "text-text-primary"}`}>
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
