import { useCallback } from "react";
import { useQueryClient } from "@tanstack/react-query";
import { reorderTask } from "@/lib/api";
import type { DragEndEvent } from "@dnd-kit/core";
import { toast } from "sonner";

export function useTaskDragDrop(campaignId: string) {
  const queryClient = useQueryClient();

  const handleDragEnd = useCallback(
    async (event: DragEndEvent) => {
      const { active, over } = event;
      if (!over || active.id === over.id) return;

      const taskId = active.id as string;
      const activeData = active.data.current as { groupId: string; position: number };
      const overData = over.data.current as { groupId: string; position: number };

      const targetGroupId = overData?.groupId ?? activeData.groupId;
      const newPosition = overData?.position ?? 0;

      try {
        await reorderTask(taskId, targetGroupId, newPosition);
        queryClient.invalidateQueries({ queryKey: ["campaign", campaignId] });
      } catch (err: any) {
        toast.error(err.message);
      }
    },
    [campaignId, queryClient]
  );

  return { handleDragEnd };
}
