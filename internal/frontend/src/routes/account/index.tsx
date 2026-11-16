import { createFileRoute, redirect } from "@tanstack/react-router";

// the account's own path, which its breadcrumb links to, opens its password tab
export const Route = createFileRoute("/account/")({
  beforeLoad: () => {
    throw redirect({ to: "/account/reset" });
  },
});
