import { useTheme } from "@/hooks/theme-provider";
import { Outlet } from "@tanstack/react-router";
import { Toaster } from "sonner";
import { AppSidebar } from "./app-sidebar";
import Header from "./header";
import { SidebarInset, SidebarProvider } from "./ui/sidebar";

export default function Root() {
  const theme = useTheme();
  return (
    <>
      <SidebarProvider className="h-dvh max-h-dvh">
        <AppSidebar />
        <SidebarInset className="overflow-scroll pr-2">
          <Header />
          <Outlet />
          <Toaster richColors theme={theme.theme} />
        </SidebarInset>
      </SidebarProvider>
    </>
  );
}
