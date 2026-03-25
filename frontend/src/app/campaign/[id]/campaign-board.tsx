"use client";

import { useState, useMemo } from "react";
import { useRouter, useParams } from "next/navigation";
import { useCampaign, useUsers } from "@/hooks/use-campaign";
import { TaskListTabs } from "@/components/task-list-tabs";
import { TaskGroup } from "@/components/task-group";
import { HideDoneToggle } from "@/components/hide-done-toggle";
import { ArrowLeft, Calendar } from "lucide-react";
import { Button } from "@/components/ui/button";
import { updateTask, setReleaseDate } from "@/lib/api";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import type { Task } from "@/lib/types";
import { TaskDetail } from "@/components/task-detail";
import { DndContext, closestCenter } from "@dnd-kit/core";
import { useTaskDragDrop } from "@/hooks/use-drag-drop";

export function CampaignBoard() {
  const router = useRouter();
  const params = useParams<{ id: string }>();
  // In production static export, useParams may return "_" from the pre-rendered shell.
  // Fall back to reading the actual UUID from the URL.
  const id = useMemo(() => {
    const paramsId = params?.id;
    if (paramsId && paramsId !== "_") return paramsId;
    if (typeof window !== "undefined") {
      const match = window.location.pathname.match(/\/campaign\/([^/]+)/);
      return match?.[1] ?? "";
    }
    return "";
  }, [params?.id]);
  const { data: campaign, isLoading } = useCampaign(id);
  const { data: users } = useUsers();
  const queryClient = useQueryClient();
  const [activeListId, setActiveListId] = useState<string | null>(null);
  const [hideDone, setHideDone] = useState(false);
  const [selectedTask, setSelectedTask] = useState<Task | null>(null);
  const { handleDragEnd } = useTaskDragDrop(id);

  const releaseDate = campaign?.release_date ?? null;
  const daysUntilRelease = releaseDate
    ? Math.ceil((new Date(releaseDate).getTime() - new Date().getTime()) / (1000 * 60 * 60 * 24))
    : null;

  const lists = campaign?.task_lists ?? [];
  const activeList = lists.find((l) => l.id === activeListId) ?? lists[0];

  // Set initial active tab once data loads
  if (activeList && !activeListId && lists.length > 0) {
    setActiveListId(activeList.id);
  }

  if (isLoading || !campaign) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-text-muted">Loading campaign...</div>
      </div>
    );
  }

  const handleStatusChange = async (taskId: string, status: Task["status"]) => {
    try {
      await updateTask(taskId, { status });
      queryClient.invalidateQueries({ queryKey: ["campaign", id] });
    } catch (err: any) {
      toast.error(err.message);
    }
  };

  return (
    <div className="min-h-screen p-6 max-w-5xl mx-auto">
      {/* Top bar */}
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-3">
          <Button variant="ghost" size="icon" onClick={() => router.push("/dashboard")}>
            <ArrowLeft className="h-4 w-4 text-accent" />
          </Button>
          <h1 className="text-2xl font-heading font-bold text-text-primary">{campaign.name}</h1>
        </div>
        <div className="flex items-center gap-3">
          <HideDoneToggle hidden={hideDone} onToggle={() => setHideDone(!hideDone)} />
          <Button
            variant="ghost"
            size="sm"
            onClick={() => router.push(`/campaign/${id}/calendar`)}
          >
            <Calendar className="h-4 w-4 mr-2" />
            Calendar
          </Button>
        </div>
      </div>

      {/* Release date + schedule */}
      <div className="flex items-center gap-4 mb-4 bg-bg-surface rounded-lg px-4 py-3">
        <label className="text-xs text-text-muted">Release Date</label>
        <input
          type="date"
          value={campaign.release_date ?? ""}
          onChange={async (e) => {
            const date = e.target.value;
            if (!date) return;
            try {
              await setReleaseDate(id, date, campaign.schedule_weeks || 8);
              queryClient.invalidateQueries({ queryKey: ["campaign", id] });
              toast.success("Release date set — task dates updated");
            } catch (err: any) {
              toast.error(err.message);
            }
          }}
          className="bg-transparent border border-border rounded-lg px-3 py-1.5 text-sm text-text-primary focus:outline-none focus:ring-1 focus:ring-accent"
        />
        <label className="text-xs text-text-muted">Schedule</label>
        <div className="flex gap-1">
          {[4, 8].map((weeks) => (
            <button
              key={weeks}
              onClick={async () => {
                if (!campaign.release_date) {
                  toast.error("Set a release date first");
                  return;
                }
                try {
                  await setReleaseDate(id, campaign.release_date, weeks);
                  queryClient.invalidateQueries({ queryKey: ["campaign", id] });
                  toast.success(`Switched to ${weeks}-week schedule`);
                } catch (err: any) {
                  toast.error(err.message);
                }
              }}
              className={`px-3 py-1.5 rounded-md text-xs font-medium transition-smooth ${
                (campaign.schedule_weeks || 8) === weeks
                  ? "bg-accent/20 text-accent"
                  : "text-text-muted hover:text-text-primary"
              }`}
            >
              {weeks}W
            </button>
          ))}
        </div>
        {daysUntilRelease !== null && (
          <span className="text-xs text-text-muted ml-auto">
            {daysUntilRelease} days until release
          </span>
        )}
      </div>

      {/* Tabs */}
      {lists.length > 0 && (
        <TaskListTabs
          lists={lists}
          activeId={activeList?.id ?? ""}
          onSelect={setActiveListId}
        />
      )}

      {/* Active list content */}
      {activeList && (
        <div className="mt-4 bg-bg-surface rounded-xl p-5">
          <DndContext collisionDetection={closestCenter} onDragEnd={handleDragEnd}>
            {(activeList.task_groups ?? []).map((group) => (
              <TaskGroup
                key={group.id}
                group={group}
                campaignId={id}
                hideDone={hideDone}
                users={users}
                onSelectTask={setSelectedTask}
                onStatusChange={handleStatusChange}
              />
            ))}
          </DndContext>
        </div>
      )}

      {selectedTask && (
        <TaskDetail
          task={selectedTask}
          onClose={() => setSelectedTask(null)}
          onUpdate={() => {
            queryClient.invalidateQueries({ queryKey: ["campaign", id] });
            // Refresh selected task
            if (selectedTask) {
              const updated = campaign?.task_lists
                ?.flatMap((l) => l.task_groups ?? [])
                ?.flatMap((g) => g.tasks ?? [])
                ?.find((t) => t.id === selectedTask.id);
              if (updated) setSelectedTask(updated);
            }
          }}
        />
      )}
    </div>
  );
}
