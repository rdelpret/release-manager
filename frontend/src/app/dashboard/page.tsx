"use client";

import { useAuth } from "@/lib/auth";
import { useRouter } from "next/navigation";
import { useEffect, useState } from "react";
import { useCampaigns, useCreateCampaign } from "@/hooks/use-campaign";
import { CampaignCard } from "@/components/campaign-card";
import { Button } from "@/components/ui/button";
import { Plus, LogOut, Music, CloudUpload, Disc3 } from "lucide-react";
import { logout } from "@/lib/api";
import { toast } from "sonner";
import type { TemplateType } from "@/lib/types";

export default function DashboardPage() {
  const { email, loading, waking } = useAuth();
  const router = useRouter();
  const { data: campaigns, isLoading } = useCampaigns();
  const createCampaign = useCreateCampaign();
  const [newName, setNewName] = useState("");
  const [releaseDate, setReleaseDate] = useState("");
  const [templateType, setTemplateType] = useState<TemplateType>("single");
  const [showCreate, setShowCreate] = useState(false);

  useEffect(() => {
    if (!loading && !email) {
      router.replace("/login");
    }
  }, [loading, email, router]);

  if (loading || isLoading) {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-center">
          <div className="text-text-muted">Loading...</div>
          {waking && (
            <div className="text-xs text-text-muted mt-2">Waking up server — this takes a few seconds on first visit</div>
          )}
        </div>
      </div>
    );
  }

  const activeCampaigns = (campaigns?.filter((c) => !c.archived) ?? []).sort((a, b) => {
    // Sort by release date (soonest first), campaigns without a date go last
    if (a.release_date && b.release_date) return a.release_date.localeCompare(b.release_date);
    if (a.release_date) return -1;
    if (b.release_date) return 1;
    return 0;
  });
  const archivedCampaigns = campaigns?.filter((c) => c.archived) ?? [];

  const handleCreate = () => {
    if (!newName.trim()) return;
    createCampaign.mutate(
      { name: newName.trim(), releaseDate: releaseDate || undefined, templateType },
      {
        onSuccess: () => {
          setNewName("");
          setReleaseDate("");
          setTemplateType("single");
          setShowCreate(false);
          toast.success("Campaign created");
        },
        onError: (err) => toast.error(err.message),
      }
    );
  };

  const handleLogout = async () => {
    await logout();
    router.replace("/login");
  };

  return (
    <div className="min-h-screen p-6 max-w-5xl mx-auto">
      <div className="flex items-center justify-between mb-8">
        <h1 className="text-3xl font-heading font-bold text-accent">Subwave</h1>
        <Button variant="ghost" size="sm" onClick={handleLogout}>
          <LogOut className="h-4 w-4 mr-2" />
          Sign out
        </Button>
      </div>

      <div className="flex items-center justify-between mb-6">
        <h2 className="text-xl font-heading font-semibold">Campaigns</h2>
        <Button
          onClick={() => setShowCreate(true)}
          className="bg-accent text-bg-base hover:bg-accent-dark"
        >
          <Plus className="h-4 w-4 mr-2" />
          New Campaign
        </Button>
      </div>

      {showCreate && (
        <div className="mb-6 bg-bg-surface p-4 rounded-xl space-y-3">
          {/* Template picker */}
          <div className="grid grid-cols-3 gap-2">
            {([
              { type: "single" as TemplateType, label: "Single", desc: "Standard release (8 weeks)", icon: Music },
              { type: "soundcloud_flip" as TemplateType, label: "SoundCloud Flip", desc: "Social-only, no DSPs (4 weeks)", icon: CloudUpload },
              { type: "lp_ep" as TemplateType, label: "LP / EP", desc: "Multi-track + merch (10-12 weeks)", icon: Disc3 },
            ]).map(({ type, label, desc, icon: Icon }) => (
              <button
                key={type}
                onClick={() => setTemplateType(type)}
                className={`flex flex-col items-center gap-1.5 rounded-lg border p-3 text-center transition-smooth ${
                  templateType === type
                    ? "border-accent bg-accent/10 text-accent"
                    : "border-border text-text-muted hover:border-text-muted hover:text-text-primary"
                }`}
              >
                <Icon className="h-5 w-5" />
                <span className="text-sm font-medium">{label}</span>
                <span className="text-[10px] leading-tight opacity-70">{desc}</span>
              </button>
            ))}
          </div>
          <div className="flex gap-3 items-center">
            <input
              autoFocus
              autoComplete="off"
              value={newName}
              onChange={(e) => setNewName(e.target.value)}
              onKeyDown={(e) => e.key === "Enter" && handleCreate()}
              placeholder="Campaign name..."
              className="flex-1 bg-transparent border border-border rounded-lg px-3 py-2 text-sm text-text-primary placeholder:text-text-muted focus:outline-none focus:ring-1 focus:ring-accent"
            />
            <Button onClick={handleCreate} className="bg-accent text-bg-base hover:bg-accent-dark">
              Create
            </Button>
            <Button variant="ghost" onClick={() => setShowCreate(false)}>
              Cancel
            </Button>
          </div>
          <div className="flex items-center gap-3">
            <label className="text-xs text-text-muted">Release Date</label>
            <input
              type="date"
              autoComplete="off"
              value={releaseDate}
              onChange={(e) => setReleaseDate(e.target.value)}
              className="bg-transparent border border-border rounded-lg px-3 py-1.5 text-sm text-text-primary focus:outline-none focus:ring-1 focus:ring-accent"
            />
            {releaseDate && (
              <span className="text-xs text-text-muted">
                Tasks will be auto-dated relative to this date
              </span>
            )}
          </div>
        </div>
      )}

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        {activeCampaigns.map((campaign) => (
          <CampaignCard key={campaign.id} campaign={campaign} />
        ))}
      </div>

      {archivedCampaigns.length > 0 && (
        <>
          <h3 className="text-lg font-heading font-semibold text-text-muted mt-10 mb-4">Archived</h3>
          <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 opacity-60">
            {archivedCampaigns.map((campaign) => (
              <CampaignCard key={campaign.id} campaign={campaign} />
            ))}
          </div>
        </>
      )}
    </div>
  );
}
