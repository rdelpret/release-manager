"use client";

import { useState, useMemo } from "react";
import { useRouter, useParams } from "next/navigation";
import { useCampaign } from "@/hooks/use-campaign";
import { CalendarView } from "@/components/calendar-view";
import { TaskDetail } from "@/components/task-detail";
import { ArrowLeft, LayoutGrid } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useQueryClient } from "@tanstack/react-query";
import type { Task } from "@/lib/types";

export function CalendarPageContent() {
  const router = useRouter();
  const params = useParams<{ id: string }>();
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
  const queryClient = useQueryClient();
  const [selectedTask, setSelectedTask] = useState<Task | null>(null);

  if (isLoading || !campaign) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-text-muted">Loading...</div>
      </div>
    );
  }

  // Flatten all tasks from all lists/groups
  const allTasks = (campaign.task_lists ?? [])
    .flatMap((l) => l.task_groups ?? [])
    .flatMap((g) => g.tasks ?? []);

  return (
    <div className="min-h-screen px-4 py-6 md:px-6 max-w-5xl mx-auto">
      <div className="flex items-center justify-between mb-6">
        <div className="flex items-center gap-3 min-w-0">
          <Button variant="ghost" size="icon" className="shrink-0" onClick={() => router.push("/dashboard")}>
            <ArrowLeft className="h-4 w-4 text-accent" />
          </Button>
          <h1 className="text-xl md:text-2xl font-heading font-bold text-text-primary truncate">{campaign.name}</h1>
        </div>
        <Button variant="ghost" size="sm" className="shrink-0" onClick={() => router.push(`/campaign/${id}`)}>
          <LayoutGrid className="h-4 w-4 sm:mr-2" />
          <span className="hidden sm:inline">Board</span>
        </Button>
      </div>

      <CalendarView tasks={allTasks} onSelectTask={setSelectedTask} />

      {selectedTask && (
        <TaskDetail
          task={selectedTask}
          onClose={() => setSelectedTask(null)}
          onUpdate={() => queryClient.invalidateQueries({ queryKey: ["campaign", id] })}
        />
      )}
    </div>
  );
}
