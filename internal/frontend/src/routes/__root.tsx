import Root from "@/components/root";
import { User } from "@/hooks/user-provider";
import { createRootRouteWithContext } from "@tanstack/react-router";

interface RouterContext {
  auth: User;
  // queryClient: QueryClient;
}

export const Route = createRootRouteWithContext<RouterContext>()({
  component: () => <Root />,
});
