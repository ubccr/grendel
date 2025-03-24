import { Outlet } from "@tanstack/react-router";
import { AppSidebar } from "./app-sidebar";
import { SidebarProvider, SidebarTrigger } from "./ui/sidebar";
import { Toaster } from "sonner";
import { useTheme } from "@/hooks/theme-provider";

export default function Root() {
  const theme = useTheme();
  return (
    <>
      <SidebarProvider>
        <AppSidebar />
        <main className="w-full">
          <SidebarTrigger />
          <Outlet />
          <Toaster richColors theme={theme.theme} />
        </main>
      </SidebarProvider>
      {/* <TanStackRouterDevtools /> */}
      {/* <ReactQueryDevtools buttonPosition="bottom-left" /> */}
    </>
  );
}
