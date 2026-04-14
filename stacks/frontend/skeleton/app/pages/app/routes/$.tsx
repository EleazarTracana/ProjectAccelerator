import { Link } from "react-router"

import { Button } from "~/components/ui"
import { APP_NAME } from "~/lib/constants"

export function meta() {
  return [{ title: `Not Found | ${APP_NAME}` }]
}

export default function NotFoundPage() {
  return (
    <div className="flex min-h-screen items-center justify-center">
      <div className="text-center">
        <h1 className="text-6xl font-bold text-slate-900 dark:text-slate-50">
          404
        </h1>
        <p className="mt-4 text-lg text-slate-600 dark:text-slate-400">
          Page not found
        </p>
        <p className="mt-2 text-sm text-slate-500 dark:text-slate-500">
          The page you are looking for does not exist or has been moved.
        </p>
        <div className="mt-8">
          <Button asChild>
            <Link to="/">Back to Home</Link>
          </Button>
        </div>
      </div>
    </div>
  )
}
