import Root from "@/components/root";
import { User } from "@/hooks/user-provider";
import { createRootRouteWithContext } from "@tanstack/react-router";
// import { TanStackRouterDevtools } from "@tanstack/router-devtools";
// import { ReactQueryDevtools } from "@tanstack/react-query-devtools";

interface RouterContext {
  // The ReturnType of your useAuth hook or the value of your AuthContext
  auth: User;
}

export const Route = createRootRouteWithContext<RouterContext>()({
  component: () => <Root />,
});
