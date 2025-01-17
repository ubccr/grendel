import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/groups/nodes/$group')({
  component: RouteComponent,
})

function RouteComponent() {
  return <div>Hello "/groups/node/$node-group"!</div>
}
