import { useNavigation } from "react-router"

/** Thin top-bar progress indicator shown during route navigations (loader in flight). */
export function NavigationPendingIndicator() {
  const navigation = useNavigation()
  const isNavigating = navigation.state !== "idle"

  if (!isNavigating) return null

  return (
    <div
      role="progressbar"
      aria-busy
      aria-label="Loading page"
      className="fixed inset-x-0 top-0 z-[100] h-0.5 overflow-hidden bg-slate-200 dark:bg-slate-800"
    >
      <div className="h-full w-1/3 animate-[slide_1s_ease-in-out_infinite] bg-teal-600 dark:bg-teal-400" />
    </div>
  )
}
