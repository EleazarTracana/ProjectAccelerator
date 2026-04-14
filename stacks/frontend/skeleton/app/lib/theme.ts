export const THEME_STORAGE_KEY = "theme-preference"

export type ThemePreference = "light" | "dark" | "system"

const VALID_THEMES: ThemePreference[] = ["light", "dark", "system"]

export function isThemePreference(value: unknown): value is ThemePreference {
  return typeof value === "string" && VALID_THEMES.includes(value as ThemePreference)
}

export function getStoredThemePreference(): ThemePreference {
  if (typeof window === "undefined") return "system"
  const stored = window.localStorage.getItem(THEME_STORAGE_KEY)
  return isThemePreference(stored) ? stored : "system"
}

export function resolveThemePreference(theme: ThemePreference): "light" | "dark" {
  if (theme === "system") {
    return typeof window.matchMedia === "function" &&
      window.matchMedia("(prefers-color-scheme: dark)").matches
      ? "dark"
      : "light"
  }
  return theme
}

export function applyThemePreference(theme: ThemePreference): void {
  if (typeof document === "undefined" || typeof window === "undefined") return

  const resolved = resolveThemePreference(theme)
  document.documentElement.classList.toggle("dark", resolved === "dark")
  document.documentElement.dataset.theme = theme
  document.documentElement.style.colorScheme = resolved
}

export const THEME_BOOTSTRAP_SCRIPT = `
(() => {
  const storageKey = "${THEME_STORAGE_KEY}";
  const validThemes = ["light", "dark", "system"];
  const storedTheme = window.localStorage.getItem(storageKey);
  const theme = validThemes.includes(storedTheme) ? storedTheme : "system";
  const resolvedTheme =
    theme === "system"
      ? window.matchMedia("(prefers-color-scheme: dark)").matches
        ? "dark"
        : "light"
      : theme;

  document.documentElement.classList.toggle("dark", resolvedTheme === "dark");
  document.documentElement.dataset.theme = theme;
  document.documentElement.style.colorScheme = resolvedTheme;
})();
`.trim()
