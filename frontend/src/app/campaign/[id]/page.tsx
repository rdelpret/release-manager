import { CampaignBoard } from "./campaign-board";

export function generateStaticParams() {
  return [{ id: "_" }];
}

export default function CampaignPage() {
  return <CampaignBoard />;
}
