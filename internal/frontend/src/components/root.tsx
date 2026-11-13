import { useTheme } from "@/hooks/theme-provider";
import { Outlet, useMatches } from "@tanstack/react-router";
import { Toaster } from "sonner";
import { AppSidebar } from "./app-sidebar";
import Header from "./header";
import { NodeField } from "./node-field";
import { SidebarInset, SidebarProvider } from "./ui/sidebar";
import { TooltipProvider } from "./ui/tooltip";

export default function Root() {
  const theme = useTheme();
  const fullField = useMatches({
    select: (matches) => matches.some((m) => m.staticData?.nodeField === "full"),
  });
  return (
    <TooltipProvider>
      {/* the node field sits behind every page, rendered once here so its groups carry on across navigation; it is full behind the pages that ask for it and quiet behind the rest */}
      <NodeField full={fullField} />
      <SidebarProvider className="h-dvh max-h-dvh">
        <AppSidebar />
        {/* Nova cards clip their overflow, which lets a flex item shrink below its content, so the children are kept at full height to make the inset scroll; the 1px side padding keeps a card's outer ring inside the scroll area, which clips it */}
        {/* the inset is transparent so the node field behind it shows through */}
        <SidebarInset className="overflow-y-scroll bg-transparent px-px pr-2 pb-2 *:shrink-0">
          <Header />
          <Outlet />
          <Toaster richColors theme={theme.theme} />
        </SidebarInset>
      </SidebarProvider>
    </TooltipProvider>
  );
}
