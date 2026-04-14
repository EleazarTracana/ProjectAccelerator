import { useState, useCallback } from "react"

export const SIDEBAR_STORAGE_KEY = "sidebar-collapsed"

function getStoredCollapsed(): boolean {
  if (typeof window === "undefined") return false
  try {
    return window.localStorage.getItem(SIDEBAR_STORAGE_KEY) === "true"
  } catch {
    return false
  }
}

export function useSidebarCollapsed(): [collapsed: boolean, toggle: () => void] {
  const [collapsed, setCollapsed] = useState<boolean>(getStoredCollapsed)

  const toggle = useCallback(() => {
    setCollapsed((prev) => {
      const next = !prev
      try {
        window.localStorage.setItem(SIDEBAR_STORAGE_KEY, String(next))
      } catch {
        // private browsing — ignore
      }
      return next
    })
  }, [])

  return [collapsed, toggle]
}
