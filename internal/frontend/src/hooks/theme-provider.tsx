import { useStoredValue } from "@/hooks/use-stored-value";
import { createContext, useContext, useEffect } from "react";

export type Theme = "dark" | "light" | "system";

type ThemeProviderProps = {
  children: React.ReactNode;
  defaultTheme?: Theme;
  storageKey?: string;
};

type ThemeProviderState = {
  theme: Theme;
  setTheme: (theme: Theme) => void;
};

const initialState: ThemeProviderState = {
  theme: "system",
  setTheme: () => null,
};

const ThemeProviderContext = createContext<ThemeProviderState>(initialState);

export function ThemeProvider({
  children,
  defaultTheme = "system",
  storageKey = "grendel-ui-theme",
  ...props
}: ThemeProviderProps) {
  // the theme is stored, so it follows a change made in another tab
  const [stored, setStored] = useStoredValue(storageKey);
  const theme = (stored as Theme | undefined) || defaultTheme;

  useEffect(() => {
    const root = window.document.documentElement;

    root.classList.remove("light", "dark");

    if (theme === "system") {
      const systemTheme = window.matchMedia("(prefers-color-scheme: dark)").matches
        ? "dark"
        : "light";

      root.classList.add(systemTheme);
      return;
    }

    root.classList.add(theme);
  }, [theme]);

  const value = {
    theme,
    setTheme: (theme: Theme) => setStored(theme),
  };

  return (
    <ThemeProviderContext.Provider {...props} value={value}>
      {children}
    </ThemeProviderContext.Provider>
  );
}

export const useTheme = () => {
  const context = useContext(ThemeProviderContext);

  if (context === undefined) throw new Error("useTheme must be used within a ThemeProvider");

  return context;
};

export function themeToMonaco(theme: Theme) {
  switch (theme) {
    case "dark":
      return "vs-dark";
    case "light":
      return "light";
    case "system":
      return window.matchMedia("(prefers-color-scheme: dark)").matches ? "vs-dark" : "light";
    default:
      console.error("failed to match theme to monaco theme.", theme);
      break;
  }
}
