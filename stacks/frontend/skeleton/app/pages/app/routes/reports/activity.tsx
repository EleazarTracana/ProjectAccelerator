import { APP_NAME } from "~/lib/constants"

export function meta() {
  return [{ title: `Reports — Activity | ${APP_NAME}` }]
}

export default function ReportsActivityPage() {
  return (
    <div className="space-y-4">
      <p className="text-sm text-slate-600 dark:text-slate-400">
        This is a placeholder for the activity report. Replace this with your
        own content — event logs, timelines, or activity feeds.
      </p>
    </div>
  )
}
