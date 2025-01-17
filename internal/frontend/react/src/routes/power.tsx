import { createFileRoute } from '@tanstack/react-router'

export const Route = createFileRoute('/power')({
  component: RouteComponent,
})

function RouteComponent() {
  return <div>Hello "/power"!</div>
}
