"use client";

import { useState } from "react";
import type { Task } from "@/lib/types";
import { SubtaskItem } from "./subtask-item";
import { X, Plus, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { updateTask, deleteTask, createSubtask } from "@/lib/api";
import { toast } from "sonner";
import { RichTextEditor } from "./rich-text-editor";
import { useUsers } from "@/hooks/use-campaign";

const statusOptions = [
  { value: "todo", label: "To Do", color: "text-text-muted" },
  { value: "in_progress", label: "In Progress", color: "text-yellow-400" },
  { value: "done", label: "Done", color: "text-green-400" },
] as const;

interface TaskDetailProps {
  task: Task;
  onClose: () => void;
  onUpdate: () => void;
}

export function TaskDetail({ task, onClose, onUpdate }: TaskDetailProps) {
  const [name, setName] = useState(task.name);
  const [status, setStatus] = useState(task.status);
  const [dueDate, setDueDate] = useState(task.due_date ?? "");
  const [assignedTo, setAssignedTo] = useState(task.assigned_to ?? "");
  const [newSubtaskName, setNewSubtaskName] = useState("");
  const { data: users } = useUsers();

  // Re-sync local state when a different task is selected
  const [trackedId, setTrackedId] = useState(task.id);
  if (task.id !== trackedId) {
    setTrackedId(task.id);
    setName(task.name);
    setStatus(task.status);
    setDueDate(task.due_date ?? "");
    setAssignedTo(task.assigned_to ?? "");
  }

  const handleNameBlur = async () => {
    if (name !== task.name && name.trim()) {
      try {
        await updateTask(task.id, { name: name.trim() } as any);
        onUpdate();
      } catch (err: any) {
        toast.error(err.message);
      }
    }
  };

  const handleStatusChange = async (newStatus: string) => {
    const prev = status;
    setStatus(newStatus as Task["status"]);
    try {
      await updateTask(task.id, { status: newStatus } as any);
      onUpdate();
    } catch (err: any) {
      setStatus(prev);
      toast.error(err.message);
    }
  };

  const handleDueDateChange = async (date: string) => {
    setDueDate(date);
    try {
      await updateTask(task.id, { due_date: date || undefined } as any);
      onUpdate();
    } catch (err: any) {
      toast.error(err.message);
    }
  };

  const handleAssign = async (userId: string) => {
    const prev = assignedTo;
    setAssignedTo(userId);
    try {
      await updateTask(task.id, { assigned_to: userId || undefined } as any);
      onUpdate();
    } catch (err: any) {
      setAssignedTo(prev);
      toast.error(err.message);
    }
  };

  const handleAddSubtask = async () => {
    if (!newSubtaskName.trim()) return;
    try {
      await createSubtask(task.id, newSubtaskName.trim());
      setNewSubtaskName("");
      onUpdate();
    } catch (err: any) {
      toast.error(err.message);
    }
  };

  const handleDelete = async () => {
    if (!confirm("Delete this task?")) return;
    try {
      await deleteTask(task.id);
      onClose();
      onUpdate();
    } catch (err: any) {
      toast.error(err.message);
    }
  };

  return (
    <div className="fixed inset-y-0 right-0 w-96 bg-bg-surface border-l border-border glass p-6 overflow-y-auto z-50">
      <div className="flex items-center justify-between mb-6">
        <input
          value={name}
          onChange={(e) => setName(e.target.value)}
          onBlur={handleNameBlur}
          autoComplete="off"
          className="text-lg font-heading font-semibold bg-transparent text-text-primary border-none focus:outline-none focus:ring-0 w-full"
        />
        <Button variant="ghost" size="icon" onClick={onClose}>
          <X className="h-4 w-4" />
        </Button>
      </div>

      {/* Status */}
      <div className="mb-5">
        <label className="text-xs font-semibold text-text-muted uppercase tracking-wider mb-2 block">
          Status
        </label>
        <div className="flex gap-2">
          {statusOptions.map((opt) => (
            <button
              key={opt.value}
              onClick={() => handleStatusChange(opt.value)}
              className={`px-3 py-1.5 rounded-md text-xs font-medium transition-smooth ${
                status === opt.value
                  ? `${opt.color} bg-white/10`
                  : "text-text-muted hover:text-text-primary"
              }`}
            >
              {opt.label}
            </button>
          ))}
        </div>
      </div>

      {/* Assignee */}
      <div className="mb-5">
        <label className="text-xs font-semibold text-text-muted uppercase tracking-wider mb-2 block">
          Assigned To
        </label>
        <select
          value={assignedTo}
          onChange={(e) => handleAssign(e.target.value)}
          autoComplete="off"
          className="w-full bg-transparent border border-border rounded-lg px-3 py-2 text-sm text-text-primary focus:outline-none focus:ring-1 focus:ring-accent"
        >
          <option value="">Unassigned</option>
          {(users ?? []).map((u) => (
            <option key={u.id} value={u.id}>
              {u.name}
            </option>
          ))}
        </select>
      </div>

      {/* Due date */}
      <div className="mb-5">
        <label className="text-xs font-semibold text-text-muted uppercase tracking-wider mb-2 block">
          Due Date
        </label>
        <input
          type="date"
          autoComplete="off"
          data-1p-ignore
          data-lpignore="true"
          data-form-type="other"
          value={dueDate}
          onChange={(e) => handleDueDateChange(e.target.value)}
          className="bg-transparent border border-border rounded-lg px-3 py-2 text-sm text-text-primary focus:outline-none focus:ring-1 focus:ring-accent"
        />
      </div>

      {/* Description */}
      <div className="mb-5">
        <label className="text-xs font-semibold text-text-muted uppercase tracking-wider mb-2 block">
          Description
        </label>
        <RichTextEditor
          content={task.description ?? null}
          onUpdate={async (content) => {
            try {
              await updateTask(task.id, { description: content } as any);
              onUpdate();
            } catch (err: any) {
              toast.error(err.message);
            }
          }}
        />
      </div>

      {/* Subtasks */}
      <div className="mb-5">
        <label className="text-xs font-semibold text-text-muted uppercase tracking-wider mb-2 block">
          Subtasks
        </label>
        <div className="space-y-1">
          {(task.subtasks ?? []).map((subtask) => (
            <SubtaskItem key={subtask.id} subtask={subtask} onUpdate={onUpdate} />
          ))}
        </div>
        <div className="flex gap-2 mt-2">
          <input
            value={newSubtaskName}
            onChange={(e) => setNewSubtaskName(e.target.value)}
            onKeyDown={(e) => e.key === "Enter" && handleAddSubtask()}
            placeholder="Add subtask..."
            autoComplete="off"
            className="flex-1 bg-transparent border-b border-border text-sm text-text-primary placeholder:text-text-muted focus:outline-none focus:border-accent"
          />
          <button onClick={handleAddSubtask} className="text-accent">
            <Plus className="h-4 w-4" />
          </button>
        </div>
      </div>

      {/* Delete */}
      <Button
        variant="ghost"
        size="sm"
        className="text-destructive hover:text-destructive mt-4"
        onClick={handleDelete}
      >
        <Trash2 className="h-4 w-4 mr-2" />
        Delete task
      </Button>
    </div>
  );
}
