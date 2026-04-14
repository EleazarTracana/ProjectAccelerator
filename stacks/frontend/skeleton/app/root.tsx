import {
  Links,
  Meta,
  Outlet,
  Scripts,
  ScrollRestoration,
} from "react-router"

import { THEME_BOOTSTRAP_SCRIPT } from "~/lib/theme"
import appStylesHref from "~/styles/app.css?url"

import type { Route } from "./+types/root"

export const links: Route.LinksFunction = () => [
  { rel: "preconnect", href: "https://fonts.googleapis.com" },
  {
    rel: "preconnect",
    href: "https://fonts.gstatic.com",
    crossOrigin: "anonymous",
  },
  {
    rel: "stylesheet",
    href: "https://fonts.googleapis.com/css2?family=Inter:wght@100..900&display=swap",
  },
  { rel: "stylesheet", href: appStylesHref },
]

export function loader() {
  return {
    ENV: {
      API_URL: process.env.API_URL ?? "http://localhost:8000",
    },
  }
}

export default function App() {
  return (
    <html lang="en" suppressHydrationWarning>
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <Meta />
        <Links />
        <script dangerouslySetInnerHTML={{ __html: THEME_BOOTSTRAP_SCRIPT }} />
      </head>
      <body className="min-h-screen bg-slate-50 font-sans text-slate-900 antialiased dark:bg-slate-950 dark:text-slate-50">
        <Outlet />
        <ScrollRestoration />
        <Scripts />
      </body>
    </html>
  )
}

export function ErrorBoundary({ error }: Route.ErrorBoundaryProps) {
  return (
    <html lang="en">
      <head>
        <meta charSet="utf-8" />
        <meta name="viewport" content="width=device-width, initial-scale=1" />
        <Meta />
        <Links />
      </head>
      <body className="flex min-h-screen items-center justify-center bg-slate-50 font-sans text-slate-900 dark:bg-slate-950 dark:text-slate-50">
        <div className="text-center">
          <h1 className="text-4xl font-bold">Something went wrong</h1>
          <p className="mt-4 text-slate-600 dark:text-slate-400">
            {error instanceof Error
              ? error.message
              : "An unexpected error occurred."}
          </p>
        </div>
        <Scripts />
      </body>
    </html>
  )
}
