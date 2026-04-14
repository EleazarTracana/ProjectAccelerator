import { APP_NAME } from "~/lib/constants"

export function meta() {
  return [{ title: `Reports — Overview | ${APP_NAME}` }]
}

export default function ReportsOverviewPage() {
  return (
    <div className="space-y-4">
      <p className="text-sm text-slate-600 dark:text-slate-400">
        This is a placeholder for the reports overview. Replace this with your
        own content — charts, tables, or summary cards.
      </p>
    </div>
  )
}
