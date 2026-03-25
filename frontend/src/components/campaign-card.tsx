"use client";

import type { Campaign } from "@/lib/types";
import { useRouter } from "next/navigation";
import { Copy, Archive, Trash2 } from "lucide-react";
import { Button } from "@/components/ui/button";
import { useDuplicateCampaign, useArchiveCampaign, useDeleteCampaign } from "@/hooks/use-campaign";
import { toast } from "sonner";

export function CampaignCard({ campaign }: { campaign: Campaign }) {
  const router = useRouter();
  const duplicate = useDuplicateCampaign();
  const archive = useArchiveCampaign();
  const del = useDeleteCampaign();

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
      className="group cursor-pointer rounded-xl bg-bg-surface p-5 transition-smooth glow-hover border border-transparent hover:border-accent/20"
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
      <div className="mt-2 flex items-center gap-2">
        <span className="inline-block rounded-full bg-white/[0.06] px-2 py-0.5 text-[10px] font-medium text-text-muted uppercase tracking-wider">
          {campaign.template_type === "soundcloud_flip" ? "SC Flip" : campaign.template_type === "lp_ep" ? "LP/EP" : "Single"}
        </span>
        <span className="text-sm text-text-muted">
          Updated {new Date(campaign.updated_at).toLocaleDateString()}
        </span>
      </div>
    </div>
  );
}
