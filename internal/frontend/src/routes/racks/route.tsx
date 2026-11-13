import { createFileRoute, Outlet } from "@tanstack/react-router";

// the pages under /racks share this parent so its crumb heads their breadcrumbs
export const Route = createFileRoute("/racks")({
  staticData: { crumb: "Racks" },
  component: Outlet,
});
