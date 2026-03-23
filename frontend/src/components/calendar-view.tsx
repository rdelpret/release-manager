"use client";

import { useMemo, useState } from "react";
import { ChevronLeft, ChevronRight } from "lucide-react";
import { Button } from "@/components/ui/button";
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
  const [year, setYear] = useState(today.getFullYear());
  const [month, setMonth] = useState(today.getMonth());

  const daysInMonth = new Date(year, month + 1, 0).getDate();
  const firstDayOfWeek = new Date(year, month, 1).getDay();

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
    const d = new Date(year, month, day);
    return d.toISOString().split("T")[0];
  };

  const goToPrevMonth = () => {
    if (month === 0) {
      setMonth(11);
      setYear(year - 1);
    } else {
      setMonth(month - 1);
    }
  };

  const goToNextMonth = () => {
    if (month === 11) {
      setMonth(0);
      setYear(year + 1);
    } else {
      setMonth(month + 1);
    }
  };

  const goToToday = () => {
    setYear(today.getFullYear());
    setMonth(today.getMonth());
  };

  const isCurrentMonth =
    year === today.getFullYear() && month === today.getMonth();

  return (
    <div>
      <div className="flex items-center justify-between mb-4">
        <h3 className="text-lg font-heading font-semibold text-text-primary">
          {new Date(year, month).toLocaleDateString("en-US", {
            month: "long",
            year: "numeric",
          })}
        </h3>
        <div className="flex items-center gap-2">
          {!isCurrentMonth && (
            <Button
              variant="ghost"
              size="sm"
              onClick={goToToday}
              className="text-xs text-text-muted hover:text-accent"
            >
              Today
            </Button>
          )}
          <Button variant="ghost" size="icon" className="h-8 w-8" onClick={goToPrevMonth}>
            <ChevronLeft className="h-4 w-4" />
          </Button>
          <Button variant="ghost" size="icon" className="h-8 w-8" onClick={goToNextMonth}>
            <ChevronRight className="h-4 w-4" />
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-7 gap-1">
        {["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"].map((d) => (
          <div
            key={d}
            className="text-xs font-semibold text-text-muted text-center py-2"
          >
            {d}
          </div>
        ))}
        {blanks.map((i) => (
          <div key={`blank-${i}`} />
        ))}
        {days.map((day) => {
          const dateStr = formatDate(day);
          const dayTasks = tasksByDate[dateStr] ?? [];
          const isToday =
            isCurrentMonth && day === today.getDate();

          return (
            <div
              key={day}
              className={`min-h-[80px] rounded-lg p-2 text-xs ${
                isToday
                  ? "border border-accent/30 bg-accent/5"
                  : "bg-bg-surface"
              }`}
            >
              <div
                className={`font-medium mb-1 ${
                  isToday ? "text-accent" : "text-text-muted"
                }`}
              >
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
