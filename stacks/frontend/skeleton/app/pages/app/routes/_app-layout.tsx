import {
  BarChart3,
  ChevronRight,
  FileText,
  LayoutDashboard,
  Settings,
} from "lucide-react"
import { Outlet } from "react-router"

import { AppSidebar, type NavSection } from "~/components/layouts/app-sidebar"
import { NavigationPendingIndicator } from "~/components/layouts/navigation-pending-indicator"
import { APP_NAME } from "~/lib/constants"
import { useSidebarCollapsed } from "~/lib/sidebar"
import { cn } from "~/lib/utils"

const NAV_SECTIONS: NavSection[] = [
  {
    items: [
      {
        label: "Dashboard",
        to: "/",
        icon: <LayoutDashboard className="h-4 w-4" />,
      },
    ],
  },
  {
    label: "Features",
    items: [
      {
        label: "Reports",
        to: "/reports",
        icon: <BarChart3 className="h-4 w-4" />,
      },
      {
        label: "Documents",
        to: "/documents",
        icon: <FileText className="h-4 w-4" />,
      },
    ],
  },
  {
    label: "System",
    items: [
      {
        label: "Settings",
        to: "/settings",
        icon: <Settings className="h-4 w-4" />,
      },
    ],
  },
]

export default function AppLayout() {
  const [collapsed, toggle] = useSidebarCollapsed()

  return (
    <div className="min-h-screen bg-white dark:bg-slate-950">
      <AppSidebar
        brandLabel={APP_NAME}
        brandTo="/"
        navSections={NAV_SECTIONS}
        collapsed={collapsed}
        onToggle={toggle}
      />
      {collapsed && (
        <button
          onClick={toggle}
          aria-label="Expand sidebar"
          className="fixed left-0 top-1/2 z-50 flex h-8 w-5 -translate-y-1/2 items-center justify-center rounded-r-md border border-l-0 border-slate-200 bg-white text-slate-400 shadow-sm transition-colors hover:bg-slate-50 hover:text-slate-600 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-slate-500 dark:border-slate-700 dark:bg-slate-900 dark:text-slate-500 dark:hover:bg-slate-800 dark:hover:text-slate-300 motion-reduce:transition-none"
        >
          <ChevronRight className="h-3 w-3" />
        </button>
      )}
      <main
        className={cn(
          "transition-[padding-left] duration-200 ease-in-out motion-reduce:transition-none",
          collapsed ? "pl-0" : "pl-60",
        )}
      >
        <NavigationPendingIndicator />
        <div className="mx-auto max-w-7xl p-6">
          <Outlet />
        </div>
      </main>
    </div>
  )
}
