/* eslint-disable */

// @ts-nocheck

// noinspection JSUnusedGlobalSymbols

// This file was automatically generated by TanStack Router.
// You should NOT make any changes in this file as it will be overwritten.
// Additionally, you should also exclude this file from your linter and/or formatter to prevent it from being checked or modified.

// Import Routes

import { Route as rootRoute } from './routes/__root'
import { Route as FloorplanImport } from './routes/floorplan'
import { Route as EventsImport } from './routes/events'
import { Route as SplatImport } from './routes/$'
import { Route as IndexImport } from './routes/index'
import { Route as TemplatesIndexImport } from './routes/templates/index'
import { Route as NodesIndexImport } from './routes/nodes/index'
import { Route as ImagesIndexImport } from './routes/images/index'
import { Route as TemplatesTemplateImport } from './routes/templates/$template'
import { Route as SearchInventoryImport } from './routes/search/inventory'
import { Route as RackRackImport } from './routes/rack/$rack'
import { Route as NodesNodeImport } from './routes/nodes/$node'
import { Route as ImagesImageImport } from './routes/images/$image'
import { Route as AddTemplateImport } from './routes/add/template'
import { Route as AddRoleImport } from './routes/add/role'
import { Route as AddNodeImport } from './routes/add/node'
import { Route as AddImageImport } from './routes/add/image'
import { Route as AccountTokenImport } from './routes/account/token'
import { Route as AccountSignupImport } from './routes/account/signup'
import { Route as AccountSigninImport } from './routes/account/signin'
import { Route as AccountResetImport } from './routes/account/reset'
import { Route as GroupsNodesIndexImport } from './routes/groups/nodes/index'
import { Route as AccountUsersIndexImport } from './routes/account/users/index'
import { Route as AccountRolesIndexImport } from './routes/account/roles/index'
import { Route as GroupsNodesGroupImport } from './routes/groups/nodes/$group'
import { Route as AccountRolesRoleImport } from './routes/account/roles/$role'

// Create/Update Routes

const FloorplanRoute = FloorplanImport.update({
  id: '/floorplan',
  path: '/floorplan',
  getParentRoute: () => rootRoute,
} as any)

const EventsRoute = EventsImport.update({
  id: '/events',
  path: '/events',
  getParentRoute: () => rootRoute,
} as any)

const SplatRoute = SplatImport.update({
  id: '/$',
  path: '/$',
  getParentRoute: () => rootRoute,
} as any)

const IndexRoute = IndexImport.update({
  id: '/',
  path: '/',
  getParentRoute: () => rootRoute,
} as any)

const TemplatesIndexRoute = TemplatesIndexImport.update({
  id: '/templates/',
  path: '/templates/',
  getParentRoute: () => rootRoute,
} as any)

const NodesIndexRoute = NodesIndexImport.update({
  id: '/nodes/',
  path: '/nodes/',
  getParentRoute: () => rootRoute,
} as any)

const ImagesIndexRoute = ImagesIndexImport.update({
  id: '/images/',
  path: '/images/',
  getParentRoute: () => rootRoute,
} as any)

const TemplatesTemplateRoute = TemplatesTemplateImport.update({
  id: '/templates/$template',
  path: '/templates/$template',
  getParentRoute: () => rootRoute,
} as any)

const SearchInventoryRoute = SearchInventoryImport.update({
  id: '/search/inventory',
  path: '/search/inventory',
  getParentRoute: () => rootRoute,
} as any)

const RackRackRoute = RackRackImport.update({
  id: '/rack/$rack',
  path: '/rack/$rack',
  getParentRoute: () => rootRoute,
} as any)

const NodesNodeRoute = NodesNodeImport.update({
  id: '/nodes/$node',
  path: '/nodes/$node',
  getParentRoute: () => rootRoute,
} as any)

const ImagesImageRoute = ImagesImageImport.update({
  id: '/images/$image',
  path: '/images/$image',
  getParentRoute: () => rootRoute,
} as any)

const AddTemplateRoute = AddTemplateImport.update({
  id: '/add/template',
  path: '/add/template',
  getParentRoute: () => rootRoute,
} as any)

const AddRoleRoute = AddRoleImport.update({
  id: '/add/role',
  path: '/add/role',
  getParentRoute: () => rootRoute,
} as any)

const AddNodeRoute = AddNodeImport.update({
  id: '/add/node',
  path: '/add/node',
  getParentRoute: () => rootRoute,
} as any)

const AddImageRoute = AddImageImport.update({
  id: '/add/image',
  path: '/add/image',
  getParentRoute: () => rootRoute,
} as any)

const AccountTokenRoute = AccountTokenImport.update({
  id: '/account/token',
  path: '/account/token',
  getParentRoute: () => rootRoute,
} as any)

const AccountSignupRoute = AccountSignupImport.update({
  id: '/account/signup',
  path: '/account/signup',
  getParentRoute: () => rootRoute,
} as any)

const AccountSigninRoute = AccountSigninImport.update({
  id: '/account/signin',
  path: '/account/signin',
  getParentRoute: () => rootRoute,
} as any)

const AccountResetRoute = AccountResetImport.update({
  id: '/account/reset',
  path: '/account/reset',
  getParentRoute: () => rootRoute,
} as any)

const GroupsNodesIndexRoute = GroupsNodesIndexImport.update({
  id: '/groups/nodes/',
  path: '/groups/nodes/',
  getParentRoute: () => rootRoute,
} as any)

const AccountUsersIndexRoute = AccountUsersIndexImport.update({
  id: '/account/users/',
  path: '/account/users/',
  getParentRoute: () => rootRoute,
} as any)

const AccountRolesIndexRoute = AccountRolesIndexImport.update({
  id: '/account/roles/',
  path: '/account/roles/',
  getParentRoute: () => rootRoute,
} as any)

const GroupsNodesGroupRoute = GroupsNodesGroupImport.update({
  id: '/groups/nodes/$group',
  path: '/groups/nodes/$group',
  getParentRoute: () => rootRoute,
} as any)

const AccountRolesRoleRoute = AccountRolesRoleImport.update({
  id: '/account/roles/$role',
  path: '/account/roles/$role',
  getParentRoute: () => rootRoute,
} as any)

// Populate the FileRoutesByPath interface

declare module '@tanstack/react-router' {
  interface FileRoutesByPath {
    '/': {
      id: '/'
      path: '/'
      fullPath: '/'
      preLoaderRoute: typeof IndexImport
      parentRoute: typeof rootRoute
    }
    '/$': {
      id: '/$'
      path: '/$'
      fullPath: '/$'
      preLoaderRoute: typeof SplatImport
      parentRoute: typeof rootRoute
    }
    '/events': {
      id: '/events'
      path: '/events'
      fullPath: '/events'
      preLoaderRoute: typeof EventsImport
      parentRoute: typeof rootRoute
    }
    '/floorplan': {
      id: '/floorplan'
      path: '/floorplan'
      fullPath: '/floorplan'
      preLoaderRoute: typeof FloorplanImport
      parentRoute: typeof rootRoute
    }
    '/account/reset': {
      id: '/account/reset'
      path: '/account/reset'
      fullPath: '/account/reset'
      preLoaderRoute: typeof AccountResetImport
      parentRoute: typeof rootRoute
    }
    '/account/signin': {
      id: '/account/signin'
      path: '/account/signin'
      fullPath: '/account/signin'
      preLoaderRoute: typeof AccountSigninImport
      parentRoute: typeof rootRoute
    }
    '/account/signup': {
      id: '/account/signup'
      path: '/account/signup'
      fullPath: '/account/signup'
      preLoaderRoute: typeof AccountSignupImport
      parentRoute: typeof rootRoute
    }
    '/account/token': {
      id: '/account/token'
      path: '/account/token'
      fullPath: '/account/token'
      preLoaderRoute: typeof AccountTokenImport
      parentRoute: typeof rootRoute
    }
    '/add/image': {
      id: '/add/image'
      path: '/add/image'
      fullPath: '/add/image'
      preLoaderRoute: typeof AddImageImport
      parentRoute: typeof rootRoute
    }
    '/add/node': {
      id: '/add/node'
      path: '/add/node'
      fullPath: '/add/node'
      preLoaderRoute: typeof AddNodeImport
      parentRoute: typeof rootRoute
    }
    '/add/role': {
      id: '/add/role'
      path: '/add/role'
      fullPath: '/add/role'
      preLoaderRoute: typeof AddRoleImport
      parentRoute: typeof rootRoute
    }
    '/add/template': {
      id: '/add/template'
      path: '/add/template'
      fullPath: '/add/template'
      preLoaderRoute: typeof AddTemplateImport
      parentRoute: typeof rootRoute
    }
    '/images/$image': {
      id: '/images/$image'
      path: '/images/$image'
      fullPath: '/images/$image'
      preLoaderRoute: typeof ImagesImageImport
      parentRoute: typeof rootRoute
    }
    '/nodes/$node': {
      id: '/nodes/$node'
      path: '/nodes/$node'
      fullPath: '/nodes/$node'
      preLoaderRoute: typeof NodesNodeImport
      parentRoute: typeof rootRoute
    }
    '/rack/$rack': {
      id: '/rack/$rack'
      path: '/rack/$rack'
      fullPath: '/rack/$rack'
      preLoaderRoute: typeof RackRackImport
      parentRoute: typeof rootRoute
    }
    '/search/inventory': {
      id: '/search/inventory'
      path: '/search/inventory'
      fullPath: '/search/inventory'
      preLoaderRoute: typeof SearchInventoryImport
      parentRoute: typeof rootRoute
    }
    '/templates/$template': {
      id: '/templates/$template'
      path: '/templates/$template'
      fullPath: '/templates/$template'
      preLoaderRoute: typeof TemplatesTemplateImport
      parentRoute: typeof rootRoute
    }
    '/images/': {
      id: '/images/'
      path: '/images'
      fullPath: '/images'
      preLoaderRoute: typeof ImagesIndexImport
      parentRoute: typeof rootRoute
    }
    '/nodes/': {
      id: '/nodes/'
      path: '/nodes'
      fullPath: '/nodes'
      preLoaderRoute: typeof NodesIndexImport
      parentRoute: typeof rootRoute
    }
    '/templates/': {
      id: '/templates/'
      path: '/templates'
      fullPath: '/templates'
      preLoaderRoute: typeof TemplatesIndexImport
      parentRoute: typeof rootRoute
    }
    '/account/roles/$role': {
      id: '/account/roles/$role'
      path: '/account/roles/$role'
      fullPath: '/account/roles/$role'
      preLoaderRoute: typeof AccountRolesRoleImport
      parentRoute: typeof rootRoute
    }
    '/groups/nodes/$group': {
      id: '/groups/nodes/$group'
      path: '/groups/nodes/$group'
      fullPath: '/groups/nodes/$group'
      preLoaderRoute: typeof GroupsNodesGroupImport
      parentRoute: typeof rootRoute
    }
    '/account/roles/': {
      id: '/account/roles/'
      path: '/account/roles'
      fullPath: '/account/roles'
      preLoaderRoute: typeof AccountRolesIndexImport
      parentRoute: typeof rootRoute
    }
    '/account/users/': {
      id: '/account/users/'
      path: '/account/users'
      fullPath: '/account/users'
      preLoaderRoute: typeof AccountUsersIndexImport
      parentRoute: typeof rootRoute
    }
    '/groups/nodes/': {
      id: '/groups/nodes/'
      path: '/groups/nodes'
      fullPath: '/groups/nodes'
      preLoaderRoute: typeof GroupsNodesIndexImport
      parentRoute: typeof rootRoute
    }
  }
}

// Create and export the route tree

export interface FileRoutesByFullPath {
  '/': typeof IndexRoute
  '/$': typeof SplatRoute
  '/events': typeof EventsRoute
  '/floorplan': typeof FloorplanRoute
  '/account/reset': typeof AccountResetRoute
  '/account/signin': typeof AccountSigninRoute
  '/account/signup': typeof AccountSignupRoute
  '/account/token': typeof AccountTokenRoute
  '/add/image': typeof AddImageRoute
  '/add/node': typeof AddNodeRoute
  '/add/role': typeof AddRoleRoute
  '/add/template': typeof AddTemplateRoute
  '/images/$image': typeof ImagesImageRoute
  '/nodes/$node': typeof NodesNodeRoute
  '/rack/$rack': typeof RackRackRoute
  '/search/inventory': typeof SearchInventoryRoute
  '/templates/$template': typeof TemplatesTemplateRoute
  '/images': typeof ImagesIndexRoute
  '/nodes': typeof NodesIndexRoute
  '/templates': typeof TemplatesIndexRoute
  '/account/roles/$role': typeof AccountRolesRoleRoute
  '/groups/nodes/$group': typeof GroupsNodesGroupRoute
  '/account/roles': typeof AccountRolesIndexRoute
  '/account/users': typeof AccountUsersIndexRoute
  '/groups/nodes': typeof GroupsNodesIndexRoute
}

export interface FileRoutesByTo {
  '/': typeof IndexRoute
  '/$': typeof SplatRoute
  '/events': typeof EventsRoute
  '/floorplan': typeof FloorplanRoute
  '/account/reset': typeof AccountResetRoute
  '/account/signin': typeof AccountSigninRoute
  '/account/signup': typeof AccountSignupRoute
  '/account/token': typeof AccountTokenRoute
  '/add/image': typeof AddImageRoute
  '/add/node': typeof AddNodeRoute
  '/add/role': typeof AddRoleRoute
  '/add/template': typeof AddTemplateRoute
  '/images/$image': typeof ImagesImageRoute
  '/nodes/$node': typeof NodesNodeRoute
  '/rack/$rack': typeof RackRackRoute
  '/search/inventory': typeof SearchInventoryRoute
  '/templates/$template': typeof TemplatesTemplateRoute
  '/images': typeof ImagesIndexRoute
  '/nodes': typeof NodesIndexRoute
  '/templates': typeof TemplatesIndexRoute
  '/account/roles/$role': typeof AccountRolesRoleRoute
  '/groups/nodes/$group': typeof GroupsNodesGroupRoute
  '/account/roles': typeof AccountRolesIndexRoute
  '/account/users': typeof AccountUsersIndexRoute
  '/groups/nodes': typeof GroupsNodesIndexRoute
}

export interface FileRoutesById {
  __root__: typeof rootRoute
  '/': typeof IndexRoute
  '/$': typeof SplatRoute
  '/events': typeof EventsRoute
  '/floorplan': typeof FloorplanRoute
  '/account/reset': typeof AccountResetRoute
  '/account/signin': typeof AccountSigninRoute
  '/account/signup': typeof AccountSignupRoute
  '/account/token': typeof AccountTokenRoute
  '/add/image': typeof AddImageRoute
  '/add/node': typeof AddNodeRoute
  '/add/role': typeof AddRoleRoute
  '/add/template': typeof AddTemplateRoute
  '/images/$image': typeof ImagesImageRoute
  '/nodes/$node': typeof NodesNodeRoute
  '/rack/$rack': typeof RackRackRoute
  '/search/inventory': typeof SearchInventoryRoute
  '/templates/$template': typeof TemplatesTemplateRoute
  '/images/': typeof ImagesIndexRoute
  '/nodes/': typeof NodesIndexRoute
  '/templates/': typeof TemplatesIndexRoute
  '/account/roles/$role': typeof AccountRolesRoleRoute
  '/groups/nodes/$group': typeof GroupsNodesGroupRoute
  '/account/roles/': typeof AccountRolesIndexRoute
  '/account/users/': typeof AccountUsersIndexRoute
  '/groups/nodes/': typeof GroupsNodesIndexRoute
}

export interface FileRouteTypes {
  fileRoutesByFullPath: FileRoutesByFullPath
  fullPaths:
    | '/'
    | '/$'
    | '/events'
    | '/floorplan'
    | '/account/reset'
    | '/account/signin'
    | '/account/signup'
    | '/account/token'
    | '/add/image'
    | '/add/node'
    | '/add/role'
    | '/add/template'
    | '/images/$image'
    | '/nodes/$node'
    | '/rack/$rack'
    | '/search/inventory'
    | '/templates/$template'
    | '/images'
    | '/nodes'
    | '/templates'
    | '/account/roles/$role'
    | '/groups/nodes/$group'
    | '/account/roles'
    | '/account/users'
    | '/groups/nodes'
  fileRoutesByTo: FileRoutesByTo
  to:
    | '/'
    | '/$'
    | '/events'
    | '/floorplan'
    | '/account/reset'
    | '/account/signin'
    | '/account/signup'
    | '/account/token'
    | '/add/image'
    | '/add/node'
    | '/add/role'
    | '/add/template'
    | '/images/$image'
    | '/nodes/$node'
    | '/rack/$rack'
    | '/search/inventory'
    | '/templates/$template'
    | '/images'
    | '/nodes'
    | '/templates'
    | '/account/roles/$role'
    | '/groups/nodes/$group'
    | '/account/roles'
    | '/account/users'
    | '/groups/nodes'
  id:
    | '__root__'
    | '/'
    | '/$'
    | '/events'
    | '/floorplan'
    | '/account/reset'
    | '/account/signin'
    | '/account/signup'
    | '/account/token'
    | '/add/image'
    | '/add/node'
    | '/add/role'
    | '/add/template'
    | '/images/$image'
    | '/nodes/$node'
    | '/rack/$rack'
    | '/search/inventory'
    | '/templates/$template'
    | '/images/'
    | '/nodes/'
    | '/templates/'
    | '/account/roles/$role'
    | '/groups/nodes/$group'
    | '/account/roles/'
    | '/account/users/'
    | '/groups/nodes/'
  fileRoutesById: FileRoutesById
}

export interface RootRouteChildren {
  IndexRoute: typeof IndexRoute
  SplatRoute: typeof SplatRoute
  EventsRoute: typeof EventsRoute
  FloorplanRoute: typeof FloorplanRoute
  AccountResetRoute: typeof AccountResetRoute
  AccountSigninRoute: typeof AccountSigninRoute
  AccountSignupRoute: typeof AccountSignupRoute
  AccountTokenRoute: typeof AccountTokenRoute
  AddImageRoute: typeof AddImageRoute
  AddNodeRoute: typeof AddNodeRoute
  AddRoleRoute: typeof AddRoleRoute
  AddTemplateRoute: typeof AddTemplateRoute
  ImagesImageRoute: typeof ImagesImageRoute
  NodesNodeRoute: typeof NodesNodeRoute
  RackRackRoute: typeof RackRackRoute
  SearchInventoryRoute: typeof SearchInventoryRoute
  TemplatesTemplateRoute: typeof TemplatesTemplateRoute
  ImagesIndexRoute: typeof ImagesIndexRoute
  NodesIndexRoute: typeof NodesIndexRoute
  TemplatesIndexRoute: typeof TemplatesIndexRoute
  AccountRolesRoleRoute: typeof AccountRolesRoleRoute
  GroupsNodesGroupRoute: typeof GroupsNodesGroupRoute
  AccountRolesIndexRoute: typeof AccountRolesIndexRoute
  AccountUsersIndexRoute: typeof AccountUsersIndexRoute
  GroupsNodesIndexRoute: typeof GroupsNodesIndexRoute
}

const rootRouteChildren: RootRouteChildren = {
  IndexRoute: IndexRoute,
  SplatRoute: SplatRoute,
  EventsRoute: EventsRoute,
  FloorplanRoute: FloorplanRoute,
  AccountResetRoute: AccountResetRoute,
  AccountSigninRoute: AccountSigninRoute,
  AccountSignupRoute: AccountSignupRoute,
  AccountTokenRoute: AccountTokenRoute,
  AddImageRoute: AddImageRoute,
  AddNodeRoute: AddNodeRoute,
  AddRoleRoute: AddRoleRoute,
  AddTemplateRoute: AddTemplateRoute,
  ImagesImageRoute: ImagesImageRoute,
  NodesNodeRoute: NodesNodeRoute,
  RackRackRoute: RackRackRoute,
  SearchInventoryRoute: SearchInventoryRoute,
  TemplatesTemplateRoute: TemplatesTemplateRoute,
  ImagesIndexRoute: ImagesIndexRoute,
  NodesIndexRoute: NodesIndexRoute,
  TemplatesIndexRoute: TemplatesIndexRoute,
  AccountRolesRoleRoute: AccountRolesRoleRoute,
  GroupsNodesGroupRoute: GroupsNodesGroupRoute,
  AccountRolesIndexRoute: AccountRolesIndexRoute,
  AccountUsersIndexRoute: AccountUsersIndexRoute,
  GroupsNodesIndexRoute: GroupsNodesIndexRoute,
}

export const routeTree = rootRoute
  ._addFileChildren(rootRouteChildren)
  ._addFileTypes<FileRouteTypes>()

/* ROUTE_MANIFEST_START
{
  "routes": {
    "__root__": {
      "filePath": "__root.tsx",
      "children": [
        "/",
        "/$",
        "/events",
        "/floorplan",
        "/account/reset",
        "/account/signin",
        "/account/signup",
        "/account/token",
        "/add/image",
        "/add/node",
        "/add/role",
        "/add/template",
        "/images/$image",
        "/nodes/$node",
        "/rack/$rack",
        "/search/inventory",
        "/templates/$template",
        "/images/",
        "/nodes/",
        "/templates/",
        "/account/roles/$role",
        "/groups/nodes/$group",
        "/account/roles/",
        "/account/users/",
        "/groups/nodes/"
      ]
    },
    "/": {
      "filePath": "index.tsx"
    },
    "/$": {
      "filePath": "$.tsx"
    },
    "/events": {
      "filePath": "events.tsx"
    },
    "/floorplan": {
      "filePath": "floorplan.tsx"
    },
    "/account/reset": {
      "filePath": "account/reset.tsx"
    },
    "/account/signin": {
      "filePath": "account/signin.tsx"
    },
    "/account/signup": {
      "filePath": "account/signup.tsx"
    },
    "/account/token": {
      "filePath": "account/token.tsx"
    },
    "/add/image": {
      "filePath": "add/image.tsx"
    },
    "/add/node": {
      "filePath": "add/node.tsx"
    },
    "/add/role": {
      "filePath": "add/role.tsx"
    },
    "/add/template": {
      "filePath": "add/template.tsx"
    },
    "/images/$image": {
      "filePath": "images/$image.tsx"
    },
    "/nodes/$node": {
      "filePath": "nodes/$node.tsx"
    },
    "/rack/$rack": {
      "filePath": "rack/$rack.tsx"
    },
    "/search/inventory": {
      "filePath": "search/inventory.tsx"
    },
    "/templates/$template": {
      "filePath": "templates/$template.tsx"
    },
    "/images/": {
      "filePath": "images/index.tsx"
    },
    "/nodes/": {
      "filePath": "nodes/index.tsx"
    },
    "/templates/": {
      "filePath": "templates/index.tsx"
    },
    "/account/roles/$role": {
      "filePath": "account/roles/$role.tsx"
    },
    "/groups/nodes/$group": {
      "filePath": "groups/nodes/$group.tsx"
    },
    "/account/roles/": {
      "filePath": "account/roles/index.tsx"
    },
    "/account/users/": {
      "filePath": "account/users/index.tsx"
    },
    "/groups/nodes/": {
      "filePath": "groups/nodes/index.tsx"
    }
  }
}
ROUTE_MANIFEST_END */
