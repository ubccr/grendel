import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import { client } from "./client/client.gen.ts";
import { ThemeProvider } from "./hooks/theme-provider.tsx";
import "./main.css";

import { RouterProvider, createRouter } from "@tanstack/react-router";
import { DefaultNotFound } from "./components/default-not-found.tsx";
import { DefaultErrorComponent } from "./components/error.tsx";
import { Loading } from "./components/loading.tsx";
import { UserProvider, useUser } from "./hooks/user-provider.tsx";
import { routeTree } from "./routeTree.gen";

import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";

const queryClient = new QueryClient();
queryClient.setDefaultOptions({
  queries: {
    retry: false,
    refetchOnMount: false,
    refetchOnWindowFocus: false,
    refetchOnReconnect: false,
  },
});

const router = createRouter({
  routeTree,
  basepath: "/ui",
  context: {
    auth: null,
    // queryClient: queryClient
  },
  scrollRestoration: true,
  defaultPreload: "intent",
  defaultPendingMs: 800,
  defaultViewTransition: true,
  defaultErrorComponent: (props) => <DefaultErrorComponent {...props} />,
  defaultNotFoundComponent: (props) => <DefaultNotFound {...props} />,
  defaultPendingComponent: () => <Loading />,
  defaultStaleTime: 10_000,
});

declare module "@tanstack/react-router" {
  interface Register {
    router: typeof router;
  }
}

client.setConfig({
  baseUrl: `${window.location.origin}`,
  throwOnError: true,
});

createRoot(document.getElementById("root")!).render(
  <StrictMode>
    <QueryClientProvider client={queryClient}>
      <UserProvider>
        <ThemeProvider defaultTheme="dark" storageKey="grendel-ui-theme">
          <InnerApp />
        </ThemeProvider>
      </UserProvider>
      <ReactQueryDevtools initialIsOpen={false} />
    </QueryClientProvider>
    <TanStackRouterDevtools router={router} initialIsOpen={false} />
  </StrictMode>,
);

function InnerApp() {
  const { user } = useUser();
  return <RouterProvider router={router} context={{ auth: user }} />;
}
