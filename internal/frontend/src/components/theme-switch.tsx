import { Moon, Sun } from "lucide-react";

import { DropdownMenuRadioGroup, DropdownMenuRadioItem } from "@/components/ui/dropdown-menu";
import { Theme, useTheme } from "@/hooks/theme-provider";

// ThemeIcon is a sun in light mode and a moon in dark, turning between them as the theme changes
export function ThemeIcon() {
  return (
    <span className="relative inline-flex size-4">
      <Sun className="size-4 scale-100 rotate-0 transition-all dark:scale-0 dark:-rotate-90" />
      <Moon className="absolute size-4 scale-0 rotate-90 transition-all dark:scale-100 dark:rotate-0" />
    </span>
  );
}

// ThemeRadioGroup picks the theme inside a dropdown menu
export function ThemeRadioGroup() {
  const { theme, setTheme } = useTheme();
  return (
    <DropdownMenuRadioGroup value={theme} onValueChange={(e) => setTheme(e as Theme)}>
      <DropdownMenuRadioItem value="light">Light</DropdownMenuRadioItem>
      <DropdownMenuRadioItem value="dark">Dark</DropdownMenuRadioItem>
      <DropdownMenuRadioItem value="system">System</DropdownMenuRadioItem>
    </DropdownMenuRadioGroup>
  );
}
