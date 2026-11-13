import { createFileRoute, Outlet } from "@tanstack/react-router";

// the pages under /nodes share this parent so its crumb heads their breadcrumbs
export const Route = createFileRoute("/nodes")({
  staticData: { crumb: "Nodes" },
  component: Outlet,
});
