import { createFileRoute, Outlet } from "@tanstack/react-router";

// the pages under /account/roles share this parent so its crumb heads their breadcrumbs
export const Route = createFileRoute("/account/roles")({
  staticData: { crumb: "Roles" },
  component: Outlet,
});
