"use client";

import { useMemo } from "react";
import type { Task } from "@/lib/types";

interface CalendarViewProps {
  tasks: Task[];
  onSelectTask: (task: Task) => void;
}

const statusColors: Record<string, string> = {
  todo: "bg-gray-500/20 text-gray-400",
  in_progress: "bg-yellow-500/20 text-yellow-400",
  done: "bg-green-500/20 text-green-400",
};

export function CalendarView({ tasks, onSelectTask }: CalendarViewProps) {
  const today = new Date();
  const currentMonth = today.getMonth();
  const currentYear = today.getFullYear();

  const daysInMonth = new Date(currentYear, currentMonth + 1, 0).getDate();
  const firstDayOfWeek = new Date(currentYear, currentMonth, 1).getDay();

  const tasksByDate = useMemo(() => {
    const map: Record<string, Task[]> = {};
    for (const task of tasks) {
      if (task.due_date) {
        const key = task.due_date;
        if (!map[key]) map[key] = [];
        map[key].push(task);
      }
    }
    return map;
  }, [tasks]);

  const days = Array.from({ length: daysInMonth }, (_, i) => i + 1);
  const blanks = Array.from({ length: firstDayOfWeek }, (_, i) => i);

  const formatDate = (day: number) => {
    const d = new Date(currentYear, currentMonth, day);
    return d.toISOString().split("T")[0];
  };

  return (
    <div>
      <h3 className="text-lg font-heading font-semibold text-text-primary mb-4">
        {new Date(currentYear, currentMonth).toLocaleDateString("en-US", { month: "long", year: "numeric" })}
      </h3>

      <div className="grid grid-cols-7 gap-1">
        {["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"].map((d) => (
          <div key={d} className="text-xs font-semibold text-text-muted text-center py-2">
            {d}
          </div>
        ))}
        {blanks.map((i) => (
          <div key={`blank-${i}`} />
        ))}
        {days.map((day) => {
          const dateStr = formatDate(day);
          const dayTasks = tasksByDate[dateStr] ?? [];
          const isToday = day === today.getDate();

          return (
            <div
              key={day}
              className={`min-h-[80px] rounded-lg p-2 text-xs ${
                isToday ? "border border-accent/30 bg-accent/5" : "bg-bg-surface"
              }`}
            >
              <div className={`font-medium mb-1 ${isToday ? "text-accent" : "text-text-muted"}`}>
                {day}
              </div>
              {dayTasks.map((task) => (
                <button
                  key={task.id}
                  onClick={() => onSelectTask(task)}
                  className={`w-full text-left px-1.5 py-0.5 rounded text-[10px] truncate mb-0.5 transition-smooth hover:opacity-80 ${statusColors[task.status]}`}
                >
                  {task.name}
                </button>
              ))}
            </div>
          );
        })}
      </div>
    </div>
  );
}
