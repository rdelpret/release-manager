"use client";

import { useState } from "react";
import type { Campaign } from "@/lib/types";
import { useRouter } from "next/navigation";
import { Copy, Archive, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useDuplicateCampaign, useArchiveCampaign, useDeleteCampaign } from "@/hooks/use-campaign";
import { toast } from "sonner";

function getUrgency(campaign: Campaign, now: number): "green" | "yellow" | "red" | "neutral" {
  if (!campaign.release_date) return "neutral";
  if (campaign.overdue_tasks > 0) return "red";
  const daysUntil = Math.ceil(
    (new Date(campaign.release_date).getTime() - now) / (1000 * 60 * 60 * 24)
  );
  // Behind if <50% done with <30% time remaining
  const pctDone = campaign.total_tasks > 0 ? campaign.done_tasks / campaign.total_tasks : 0;
  const totalDays = campaign.schedule_weeks * 7;
  const pctTimeLeft = totalDays > 0 ? daysUntil / totalDays : 1;
  if (pctDone < 0.5 && pctTimeLeft < 0.3) return "yellow";
  return "green";
}

const urgencyBorder = {
  green: "border-green-500/30",
  yellow: "border-yellow-500/30",
  red: "border-red-500/30",
  neutral: "border-transparent",
};

const urgencyBar = {
  green: "bg-green-500",
  yellow: "bg-yellow-500",
  red: "bg-red-500",
  neutral: "bg-accent",
};

export function CampaignCard({ campaign }: { campaign: Campaign }) {
  const router = useRouter();
  const duplicate = useDuplicateCampaign();
  const archive = useArchiveCampaign();
  const del = useDeleteCampaign();

  const [now] = useState(() => Date.now());
  const urgency = getUrgency(campaign, now);

  const pctDone = campaign.total_tasks > 0
    ? Math.round((campaign.done_tasks / campaign.total_tasks) * 100)
    : 0;

  const daysUntilRelease = campaign.release_date
    ? Math.ceil((new Date(campaign.release_date).getTime() - now) / (1000 * 60 * 60 * 24))
    : null;

  const handleDuplicate = (e: React.MouseEvent) => {
    e.stopPropagation();
    duplicate.mutate(campaign.id, {
      onSuccess: () => toast.success("Campaign duplicated"),
      onError: (err) => toast.error(err.message),
    });
  };

  const handleArchive = (e: React.MouseEvent) => {
    e.stopPropagation();
    archive.mutate(
      { id: campaign.id, archived: !campaign.archived },
      {
        onSuccess: () => toast.success(campaign.archived ? "Campaign unarchived" : "Campaign archived"),
        onError: (err) => toast.error(err.message),
      }
    );
  };

  const handleDelete = (e: React.MouseEvent) => {
    e.stopPropagation();
    if (!confirm("Permanently delete this campaign?")) return;
    del.mutate(campaign.id, {
      onSuccess: () => toast.success("Campaign deleted"),
      onError: (err) => toast.error(err.message),
    });
  };

  return (
    <div
      onClick={() => router.push(`/campaign/${campaign.id}`)}
      className={`group cursor-pointer rounded-xl bg-bg-surface p-5 transition-smooth glow-hover border ${urgencyBorder[urgency]} hover:border-accent/20`}
    >
      <div className="flex items-start justify-between">
        <h3 className="font-heading text-lg font-semibold text-text-primary">
          {campaign.name}
        </h3>
        <div className="flex gap-1 opacity-0 group-hover:opacity-100 transition-smooth">
          <Button variant="ghost" size="icon" className="h-7 w-7" onClick={handleDuplicate}>
            <Copy className="h-3.5 w-3.5" />
          </Button>
          <Button variant="ghost" size="icon" className="h-7 w-7" onClick={handleArchive}>
            <Archive className="h-3.5 w-3.5" />
          </Button>
          <Button variant="ghost" size="icon" className="h-7 w-7 text-destructive" onClick={handleDelete}>
            <Trash2 className="h-3.5 w-3.5" />
          </Button>
        </div>
      </div>

      {/* Progress bar */}
      {campaign.total_tasks > 0 && (
        <div className="mt-3">
          <div className="flex items-center justify-between text-xs text-text-muted mb-1">
            <span>{campaign.done_tasks}/{campaign.total_tasks} done</span>
            <span>{pctDone}%</span>
          </div>
          <div className="h-1.5 rounded-full bg-white/[0.06] overflow-hidden">
            <div
              className={`h-full rounded-full transition-all ${urgencyBar[urgency]}`}
              style={{ width: `${pctDone}%` }}
            />
          </div>
        </div>
      )}

      <div className="mt-2 flex items-center gap-2 flex-wrap">
        <span className="inline-block rounded-full bg-white/[0.06] px-2 py-0.5 text-[10px] font-medium text-text-muted uppercase tracking-wider">
          {campaign.template_type === "soundcloud_flip" ? "SC Flip" : campaign.template_type === "lp_ep" ? "LP/EP" : "Single"}
        </span>
        {campaign.release_date && (
          <span className="text-xs text-text-muted">
            {new Date(campaign.release_date).toLocaleDateString("en-US", { month: "short", day: "numeric" })}
            {daysUntilRelease !== null && (
              <> &middot; {daysUntilRelease > 0 ? `${daysUntilRelease}d` : daysUntilRelease === 0 ? "today" : `${Math.abs(daysUntilRelease)}d ago`}</>
            )}
          </span>
        )}
        {campaign.overdue_tasks > 0 && (
          <span className="text-[10px] font-medium text-red-400">
            {campaign.overdue_tasks} overdue
          </span>
        )}
      </div>
    </div>
  );
}
