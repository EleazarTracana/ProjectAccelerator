import { FeatureLayout, type TabItem } from "~/components/layouts/feature-layout"

const TABS: TabItem[] = [
  { label: "Overview", to: "/reports" },
  { label: "Activity", to: "/reports/activity" },
]

export default function ReportsLayout() {
  return (
    <FeatureLayout
      title="Reports"
      description="Analytics and summaries for your project activity."
      tabs={TABS}
    />
  )
}
