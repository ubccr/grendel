import { AppSidebar } from "@/components/app-sidebar";
import { SidebarProvider, SidebarTrigger } from "@/components/ui/sidebar";
import { Toaster } from "@/components/ui/sonner";
import { createRootRoute, Outlet } from "@tanstack/react-router";
// import { TanStackRouterDevtools } from "@tanstack/router-devtools";
// import { ReactQueryDevtools } from "@tanstack/react-query-devtools";

export const Route = createRootRoute({
    component: () => (
        <>
            <SidebarProvider>
                <AppSidebar />
                <main className="w-full">
                    <SidebarTrigger />
                    <Outlet />
                    <Toaster richColors />
                </main>
            </SidebarProvider>
            {/* <TanStackRouterDevtools /> */}
            {/* <ReactQueryDevtools buttonPosition="bottom-left" /> */}
        </>
    ),
});
