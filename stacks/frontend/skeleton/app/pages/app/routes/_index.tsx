import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "~/components/ui"
import { APP_NAME } from "~/lib/constants"

export function meta() {
  return [
    { title: `Dashboard | ${APP_NAME}` },
    { name: "description", content: `${APP_NAME} dashboard` },
  ]
}

export default function DashboardPage() {
  return (
    <div className="space-y-6">
      <div>
        <h1 className="text-3xl font-bold tracking-tight">Dashboard</h1>
        <p className="mt-1 text-slate-500 dark:text-slate-400">
          Welcome to {APP_NAME}. This is your starting point.
        </p>
      </div>

      <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3">
        <Card>
          <CardHeader>
            <CardTitle>Getting Started</CardTitle>
            <CardDescription>
              Your skeleton is up and running
            </CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-slate-600 dark:text-slate-300">
              This React Router 7 skeleton includes SSR, Tailwind CSS v4, a
              theme system, and a component library ready for you to build on.
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Add a Domain</CardTitle>
            <CardDescription>
              Organize features by domain
            </CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-slate-600 dark:text-slate-300">
              Create a new folder under <code className="rounded bg-slate-100 px-1 py-0.5 text-xs dark:bg-slate-800">app/pages/</code> for
              each domain, with its own routes, components, and hooks.
            </p>
          </CardContent>
        </Card>

        <Card>
          <CardHeader>
            <CardTitle>Theme Support</CardTitle>
            <CardDescription>
              Light, dark, and system modes
            </CardDescription>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-slate-600 dark:text-slate-300">
              Toggle the theme using the button in the sidebar footer. The
              preference is persisted in localStorage and applied before hydration
              to prevent flash.
            </p>
          </CardContent>
        </Card>
      </div>
    </div>
  )
}
