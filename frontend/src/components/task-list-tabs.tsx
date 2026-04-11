"use client";

import type { TaskList } from "@/lib/types";

interface TaskListTabsProps {
  lists: TaskList[];
  activeId: string;
  onSelect: (id: string) => void;
}

export function TaskListTabs({ lists, activeId, onSelect }: TaskListTabsProps) {
  return (
    <div className="overflow-x-auto -mx-4 px-4 md:mx-0 md:px-0">
      <div className="flex gap-0 bg-bg-surface rounded-lg overflow-hidden w-max md:w-auto">
        {lists.map((list) => (
          <button
            key={list.id}
            onClick={() => onSelect(list.id)}
            className={`px-4 py-3 text-xs font-bold uppercase tracking-wider transition-smooth whitespace-nowrap ${
              activeId === list.id
                ? "border-b-2"
                : "text-text-muted hover:text-text-primary"
            }`}
            style={
              activeId === list.id
                ? { color: list.color, borderBottomColor: list.color }
                : undefined
            }
          >
            {list.name}
          </button>
        ))}
      </div>
    </div>
  );
}
