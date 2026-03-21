"use client";

import { Eye, EyeOff } from "lucide-react";

export function HideDoneToggle({
  hidden,
  onToggle,
}: {
  hidden: boolean;
  onToggle: () => void;
}) {
  return (
    <button
      onClick={onToggle}
      className="flex items-center gap-2 text-xs text-text-muted hover:text-text-primary transition-smooth"
    >
      {hidden ? <EyeOff className="h-3.5 w-3.5" /> : <Eye className="h-3.5 w-3.5" />}
      {hidden ? "Show completed" : "Hide completed"}
    </button>
  );
}
