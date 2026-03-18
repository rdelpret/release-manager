"use client";

import { useState } from "react";
import { useRouter, useParams } from "next/navigation";
import { useCampaign } from "@/hooks/use-campaign";
import { TaskListTabs } from "@/components/task-list-tabs";
import { TaskGroup } from "@/components/task-group";
import { HideDoneToggle } from "@/components/hide-done-toggle";
import { ArrowLeft, Calendar } from "lucide-react";
import { Button } from "@/components/ui/button";
import { updateTask } from "@/lib/api";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import type { Task } from "@/lib/types";
import { TaskDetail } from "@/components/task-detail";

export function CampaignBoard() {
  const router = useRouter();
  const params = useParams<{ id: string }>();
  const id = params?.id ?? "";
  const { data: campaign, isLoading } = useCampaign(id);
  const queryClient = useQueryClient();
  const [activeListId, setActiveListId] = useState<string | null>(null);
  const [hideDone, setHideDone] = useState(false);
  const [selectedTask, setSelectedTask] = useState<Task | null>(null);

  if (isLoading || !campaign) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-text-muted">Loading campaign...</div>
      </div>
    );
  }

  const lists = campaign.task_lists ?? [];
  const activeList = lists.find((l) => l.id === activeListId) ?? lists[0];

  if (activeList && !activeListId) {
    setActiveListId(activeList.id);
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
          {(activeList.task_groups ?? []).map((group) => (
            <TaskGroup
              key={group.id}
              group={group}
              campaignId={id}
              hideDone={hideDone}
              onSelectTask={setSelectedTask}
              onStatusChange={handleStatusChange}
            />
          ))}
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
