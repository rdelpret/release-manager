import { CalendarPageContent } from "./calendar-page";

export function generateStaticParams() {
  return [{ id: "_" }];
}

export default function CalendarPage() {
  return <CalendarPageContent />;
}
