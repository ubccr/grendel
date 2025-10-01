import { Outlet } from "@tanstack/react-router";
import { AppSidebar } from "./app-sidebar";
import { SidebarInset, SidebarProvider } from "./ui/sidebar";
import { Toaster } from "sonner";
import { useTheme } from "@/hooks/theme-provider";
import Header from "./header";
// import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";

export default function Root() {
  const theme = useTheme();
  return (
    <>
      <SidebarProvider className="h-dvh max-h-dvh">
        <AppSidebar />
        {/* @container/content has-[[data-layout=fixed]]:h-svh peer-data-[variant=inset]:has-[[data-layout=fixed]]:h-[calc(100svh-(var(--spacing)*4))] */}
        <SidebarInset className="overflow-scroll pr-2">
          <Header />
          <Outlet />
          <Toaster richColors theme={theme.theme} />
          {/* <TanStackRouterDevtools /> */}
        </SidebarInset>
      </SidebarProvider>
    </>
  );
}
