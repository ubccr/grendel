# Grendel frontend

## Technologies

- React v19
- Shadcn/ui
- Tailwind
- Tanstack Form
- Tanstack Query
- Tanstack Router
- Tanstack Table
- Monaco
- Zod

## Dev setup:

- `cd internal/frontend`
- `yarn install`
- `yarn run dev`

By default, the grendel API is hosted at `0.0.0.0:8080`. This will serve the dist located in `internal/api/dist`, and can be rebuilt with the `yarn run build` command.

The vite dev server uses vite's proxy, defined in `internal/frontend/vite.config.ts` to proxy API requests from frontend dev server to the Grendel API.

## OpenAPI codegen

We use Tanstack Query with [open-api-react-query-codegen](https://github.com/7nohe/openapi-react-query-codegen) to automatically compile a typesafe client for our frontend. This can be built by running `yarn run codegen`.
