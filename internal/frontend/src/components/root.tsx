import { useTheme } from "@/hooks/theme-provider";
import { Outlet } from "@tanstack/react-router";
import { Toaster } from "sonner";
import { AppSidebar } from "./app-sidebar";
import Header from "./header";
import { SidebarInset, SidebarProvider } from "./ui/sidebar";
import { TooltipProvider } from "./ui/tooltip";

export default function Root() {
  const theme = useTheme();
  return (
    <TooltipProvider>
      <SidebarProvider className="h-dvh max-h-dvh">
        <AppSidebar />
        {/* Nova cards clip their overflow, which lets a flex item shrink below its content, so the children are kept at full height to make the inset scroll; the 1px side padding keeps a card's outer ring inside the scroll area, which clips it */}
        <SidebarInset className="overflow-y-scroll px-px pr-2 pb-2 *:shrink-0">
          <Header />
          <Outlet />
          <Toaster richColors theme={theme.theme} />
        </SidebarInset>
      </SidebarProvider>
    </TooltipProvider>
  );
}
