import { Monitor, Moon, Sun } from "lucide-react"
import { useEffect, useMemo, useState } from "react"

import { Label } from "~/components/ui"
import {
  applyThemePreference,
  getStoredThemePreference,
  THEME_STORAGE_KEY,
  type ThemePreference,
} from "~/lib/theme"

const themeIcons = {
  light: Sun,
  dark: Moon,
  system: Monitor,
} satisfies Record<ThemePreference, typeof Sun>

const selectClassName =
  "min-w-[8.5rem] rounded-md border border-slate-200 bg-white px-2 py-1.5 text-sm text-slate-900 shadow-sm dark:border-slate-700 dark:bg-slate-900 dark:text-slate-100"

export function ThemeToggle() {
  const [theme, setTheme] = useState<ThemePreference>("system")

  useEffect(() => {
    const storedTheme = getStoredThemePreference()
    setTheme(storedTheme)
    applyThemePreference(storedTheme)
  }, [])

  useEffect(() => {
    if (typeof window.matchMedia !== "function") return

    const mediaQuery = window.matchMedia("(prefers-color-scheme: dark)")
    const handleChange = () => {
      if (theme === "system") {
        applyThemePreference("system")
      }
    }

    mediaQuery.addEventListener("change", handleChange)
    return () => mediaQuery.removeEventListener("change", handleChange)
  }, [theme])

  const Icon = useMemo(() => themeIcons[theme], [theme])

  function handleThemeChange(nextTheme: ThemePreference) {
    setTheme(nextTheme)
    window.localStorage.setItem(THEME_STORAGE_KEY, nextTheme)
    applyThemePreference(nextTheme)
  }

  return (
    <div className="flex flex-col gap-1.5 sm:flex-row sm:items-center sm:gap-2">
      <Label
        htmlFor="theme-select"
        className="flex items-center gap-2 text-xs font-medium uppercase tracking-wider text-slate-500 whitespace-nowrap dark:text-slate-400"
      >
        <Icon className="h-3.5 w-3.5 shrink-0" aria-hidden />
        Theme
      </Label>
      <select
        id="theme-select"
        aria-label="Theme"
        className={selectClassName}
        value={theme}
        onChange={(event) =>
          handleThemeChange(event.target.value as ThemePreference)
        }
      >
        <option value="system">System</option>
        <option value="light">Light</option>
        <option value="dark">Dark</option>
      </select>
    </div>
  )
}
