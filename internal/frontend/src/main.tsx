import { QueryClientProvider } from "@tanstack/react-query";
import { StrictMode, useEffect, useRef } from "react";
import { createRoot } from "react-dom/client";
import { client } from "./client/client.gen.ts";
import { ThemeProvider } from "./hooks/theme-provider.tsx";
// Geist is imported here rather than from main.css, because the Tailwind PostCSS plugin inlines CSS imports without rebasing their font urls
import "@fontsource-variable/geist";
import "./main.css";

import { RouterProvider, createRouter } from "@tanstack/react-router";
import { DefaultNotFound } from "./components/default-not-found.tsx";
import { DefaultErrorComponent } from "./components/error.tsx";
import { Loading } from "./components/loading.tsx";
import { UserProvider, useUser } from "./hooks/user-provider.tsx";
import { createQueryClient } from "./lib/query-client.ts";
import { routeTree } from "./routeTree.gen";

import { ReactQueryDevtools } from "@tanstack/react-query-devtools";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";

const queryClient = createQueryClient();

const router = createRouter({
  routeTree,
  basepath: "/ui",
  context: {
    auth: null,
    queryClient,
  },
  scrollRestoration: true,
  defaultPreload: "intent",
  // loaders read through the query client, which decides what is fresh, so the router keeps no copy of its own to go stale beside it
  defaultPreloadStaleTime: 0,
  defaultPendingMs: 800,
  defaultViewTransition: true,
  defaultErrorComponent: (props) => <DefaultErrorComponent {...props} />,
  defaultNotFoundComponent: (props) => <DefaultNotFound {...props} />,
  defaultPendingComponent: () => <Loading />,
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
  const signedIn = user !== null;
  const wasSignedIn = useRef(signedIn);
  // however a session starts or ends, by the menu, a password change, its token expiring or another tab, it ends up here. The route guards only run when a route loads, so they are run again, which takes a signed out person to the sign in page, and only then is what the session read thrown away, so the page being left does not ask for it again
  useEffect(() => {
    if (wasSignedIn.current === signedIn) return;
    wasSignedIn.current = signedIn;
    void router.invalidate().then(() => {
      if (!signedIn) queryClient.clear();
    });
  }, [signedIn]);
  return <RouterProvider router={router} context={{ auth: user, queryClient }} />;
}
