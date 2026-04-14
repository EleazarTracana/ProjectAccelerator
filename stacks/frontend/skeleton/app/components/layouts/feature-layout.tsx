import { NavLink, Outlet, useLocation } from "react-router"

export interface TabItem {
  label: string
  to: string
}

interface FeatureLayoutProps {
  title: string
  description: string
  tabs: TabItem[]
}

export function FeatureLayout({ title, description, tabs }: FeatureLayoutProps) {
  const { search } = useLocation()

  return (
    <div className="container mx-auto max-w-5xl px-4 py-8 sm:px-6 lg:px-8">
      <div className="mb-6">
        <h1 className="mb-2 text-2xl font-semibold text-slate-900 dark:text-slate-50">
          {title}
        </h1>
        <p className="text-sm text-slate-600 dark:text-slate-400">
          {description}
        </p>
      </div>

      <nav className="mb-8 flex flex-wrap gap-4 border-b border-slate-200 dark:border-slate-700">
        {tabs.map((tab) => (
          <NavLink
            key={tab.to}
            to={`${tab.to}${search}`}
            end={tab.to === tabs[0]?.to}
            className={({ isActive, isPending }) =>
              `-mb-px border-b-2 pb-2 text-sm font-medium transition-colors ${
                isPending
                  ? "pointer-events-none border-transparent text-slate-400 dark:text-slate-500"
                  : isActive
                    ? "border-teal-700 text-teal-800 dark:border-teal-500 dark:text-teal-300"
                    : "border-transparent text-slate-600 hover:text-slate-900 dark:text-slate-400 dark:hover:text-slate-50"
              }`
            }
          >
            {tab.label}
          </NavLink>
        ))}
      </nav>

      <Outlet />
    </div>
  )
}
