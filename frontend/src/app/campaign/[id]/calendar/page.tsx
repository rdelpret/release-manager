import { CalendarPageContent } from "./calendar-page";

export function generateStaticParams() {
  // Placeholder to satisfy static export requirement.
  // All routing is handled client-side via useParams().
  return [{ id: "_" }];
}

export default function CalendarPage() {
  return <CalendarPageContent />;
}
