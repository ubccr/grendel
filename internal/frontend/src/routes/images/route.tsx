import { createFileRoute, Outlet } from "@tanstack/react-router";

// the pages under /images share this parent so its crumb heads their breadcrumbs
export const Route = createFileRoute("/images")({
  staticData: { crumb: "Images" },
  component: Outlet,
});
