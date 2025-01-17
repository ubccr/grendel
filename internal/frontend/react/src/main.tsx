import { StrictMode } from "react";
import { createRoot } from "react-dom/client";
import "./main.css";
import { ThemeProvider } from "./components/theme-provider.tsx";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { client } from "./openapi/requests/services.gen.ts";
import { broadcastQueryClient } from "@tanstack/query-broadcast-client-experimental";

// Import the generated route tree
import { routeTree } from "./routeTree.gen";
import { RouterProvider, createRouter } from "@tanstack/react-router";

// Create a new router instance
const router = createRouter({ routeTree });

// Register the router instance for type safety
declare module "@tanstack/react-router" {
    interface Register {
        router: typeof router;
    }
}

const queryClient = new QueryClient();

client.setConfig({
    baseUrl: `${window.location.origin}/v1`,
    throwOnError: true, // If you want to handle errors on `onError` callback of `useQuery` and `useMutation`, set this to `true`
});

client.interceptors.request.use((config) => {
    // Add your request interceptor logic here
    return config;
});

client.interceptors.response.use((response) => {
    // Add your response interceptor logic here
    return response;
});

// TODO: look into server side websocket syncing
broadcastQueryClient({
    queryClient,
});

createRoot(document.getElementById("root")!).render(
    <StrictMode>
        <QueryClientProvider client={queryClient}>
            <ThemeProvider defaultTheme="dark" storageKey="vite-ui-theme">
                <RouterProvider router={router} />
            </ThemeProvider>
            {/* <ReactQueryDevtools initialIsOpen={false} /> */}
        </QueryClientProvider>
    </StrictMode>,
);
