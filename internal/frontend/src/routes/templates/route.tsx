import { createFileRoute, Outlet } from "@tanstack/react-router";

// the pages under /templates share this parent so its crumb heads their breadcrumbs
export const Route = createFileRoute("/templates")({
  staticData: { crumb: "Templates" },
  component: Outlet,
});
