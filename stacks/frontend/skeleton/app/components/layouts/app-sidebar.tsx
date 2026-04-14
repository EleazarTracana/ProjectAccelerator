import { Link, NavLink, useLocation } from "react-router"
import { ChevronLeft } from "lucide-react"

import { ThemeToggle } from "~/components/layouts/theme-toggle"
import { cn } from "~/lib/utils"

export interface NavItem {
  label: string
  to: string
  icon: React.ReactNode
}

export interface NavSection {
  label?: string
  items: NavItem[]
}

interface SidebarNavItemProps {
  item: NavItem
}

function SidebarNavItem({ item }: SidebarNavItemProps) {
  const location = useLocation()
  const isActive =
    item.to === "/"
      ? location.pathname === "/"
      : location.pathname.startsWith(item.to)

  return (
    <NavLink
      to={item.to}
      end={item.to === "/"}
      className={cn(
        "flex items-center gap-3 rounded-md px-3 py-2 text-sm font-medium transition-colors",
        isActive
          ? "bg-slate-100 text-slate-900 dark:bg-slate-800 dark:text-slate-50"
          : "text-slate-600 hover:bg-slate-50 hover:text-slate-900 dark:text-slate-300 dark:hover:bg-slate-800 dark:hover:text-slate-50",
      )}
    >
      {item.icon}
      {item.label}
    </NavLink>
  )
}

interface AppSidebarProps {
  brandLabel: string
  brandTo: string
  navSections: NavSection[]
  collapsed: boolean
  onToggle: () => void
}

export function AppSidebar({
  brandLabel,
  brandTo,
  navSections,
  collapsed,
  onToggle,
}: AppSidebarProps) {
  return (
    <aside
      className={cn(
        "fixed inset-y-0 left-0 z-40 flex w-60 flex-col border-r border-slate-200 bg-white dark:border-slate-800 dark:bg-slate-950",
        "transition-transform duration-200 ease-in-out motion-reduce:transition-none",
        collapsed && "-translate-x-full",
      )}
      aria-hidden={collapsed}
      inert={collapsed ? true : undefined}
    >
      {/* Header */}
      <div className="flex h-14 items-center justify-between border-b border-slate-200 px-4 dark:border-slate-800">
        <Link
          to={brandTo}
          className="text-base font-bold text-slate-900 dark:text-slate-50"
        >
          {brandLabel}
        </Link>
        <button
          onClick={onToggle}
          aria-label="Collapse sidebar"
          className="rounded-md p-1.5 text-slate-400 transition-colors hover:bg-slate-100 hover:text-slate-600 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-slate-500 dark:text-slate-500 dark:hover:bg-slate-800 dark:hover:text-slate-300 motion-reduce:transition-none"
        >
          <ChevronLeft className="h-4 w-4" />
        </button>
      </div>

      {/* Body */}
      <div className="flex-1 overflow-y-auto px-3 py-4">
        {navSections.map((section, i) => (
          <div key={section.label ?? i} className={i > 0 ? "mt-6" : ""}>
            {section.label && (
              <p className="mb-2 px-3 text-xs font-semibold uppercase tracking-wider text-slate-400 dark:text-slate-500">
                {section.label}
              </p>
            )}
            <nav className="space-y-1">
              {section.items.map((item) => (
                <SidebarNavItem key={item.to} item={item} />
              ))}
            </nav>
          </div>
        ))}
      </div>

      {/* Footer */}
      <div className="border-t border-slate-200 p-4 dark:border-slate-800">
        <div className="mb-3">
          <ThemeToggle />
        </div>
        <Link
          to="/about"
          className="flex items-center gap-2 rounded-md px-3 py-2 text-sm font-medium text-slate-600 transition-colors hover:bg-slate-50 hover:text-slate-900 dark:text-slate-300 dark:hover:bg-slate-800 dark:hover:text-slate-50"
        >
          About
        </Link>
      </div>
    </aside>
  )
}
