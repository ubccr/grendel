import { createFileRoute, redirect } from "@tanstack/react-router";

// a node's own path, which its breadcrumb links to, opens its node tab
export const Route = createFileRoute("/nodes/$node/")({
  beforeLoad: ({ params, search }) => {
    throw redirect({ to: "/nodes/$node/node", params, search });
  },
});
