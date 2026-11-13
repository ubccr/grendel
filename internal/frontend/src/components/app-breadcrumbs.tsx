import {
  Breadcrumb,
  BreadcrumbItem,
  BreadcrumbLink,
  BreadcrumbList,
  BreadcrumbPage,
  BreadcrumbSeparator,
} from "@/components/ui/breadcrumb";
import { cn } from "@/lib/utils";
import { Link, useMatch, useMatches } from "@tanstack/react-router";
import { Fragment } from "react";

// a crumb part is a fixed label, the value of one of the route's path params (a name, like a node, set in mono), or a link to a list the route isn't nested under, like Nodes for /add/node
export type CrumbPart = string | { param: string } | { label: string; to: string };

// routes name their place in the breadcrumbs through their static data; a route with several parts links only its last one to itself
declare module "@tanstack/react-router" {
  interface StaticDataRouteOption {
    crumb?: CrumbPart | CrumbPart[];
  }
}

export type Crumb = {
  label: string;
  name?: boolean;
  link?: { to: string; params?: Record<string, string> };
};

type CrumbMatch = {
  routeId: string;
  fullPath: string;
  params: Record<string, string>;
  staticData: { crumb?: CrumbPart | CrumbPart[] };
};

// crumbsFrom turns the matched routes into the trail shown in the header. A node opened from its rack is shown under that rack in place of the node list
export function crumbsFrom(matches: CrumbMatch[], rack?: string): Crumb[] {
  return matches.flatMap((match): Crumb[] => {
    if (rack && match.routeId === "/nodes") {
      return [
        { label: "Racks", link: { to: "/racks" } },
        { label: rack, name: true, link: { to: "/racks/$rack", params: { rack } } },
      ];
    }
    const crumb = match.staticData.crumb;
    if (crumb === undefined) return [];
    const parts = Array.isArray(crumb) ? crumb : [crumb];
    return parts.map((part, i) => {
      const self =
        i === parts.length - 1 ? { to: match.fullPath, params: match.params } : undefined;
      if (typeof part === "string") return { label: part, link: self };
      if ("param" in part) return { label: match.params[part.param], name: true, link: self };
      return { label: part.label, link: { to: part.to } };
    });
  });
}

// useCrumbs is the trail for the current page
export function useCrumbs(): Crumb[] {
  const matches = useMatches();
  const node = useMatch({ from: "/nodes/$node", shouldThrow: false });
  // only trust the rack a link says a node came from when the node is tagged with it
  const from = node?.search.from;
  const rack =
    from && node.loaderData?.data?.[0]?.tags?.includes(`rack=${from}`) ? from : undefined;
  return crumbsFrom(matches, rack);
}

// AppBreadcrumbs draws a trail on one line; on narrow screens only the page itself is shown
export function AppBreadcrumbs({ crumbs, className }: { crumbs: Crumb[]; className?: string }) {
  return (
    <Breadcrumb className={cn("min-w-0", className)}>
      <BreadcrumbList className="flex-nowrap">
        {crumbs.map((crumb, i) => {
          const last = i === crumbs.length - 1;
          const label = cn("truncate", crumb.name && "font-mono");
          return (
            <Fragment key={i}>
              <BreadcrumbItem className={cn("min-w-0", !last && "hidden md:inline-flex")}>
                {last ? (
                  <BreadcrumbPage className={label}>{crumb.label}</BreadcrumbPage>
                ) : crumb.link ? (
                  <BreadcrumbLink className={label} render={<Link {...crumb.link} />}>
                    {crumb.label}
                  </BreadcrumbLink>
                ) : (
                  <span className={label}>{crumb.label}</span>
                )}
              </BreadcrumbItem>
              {!last && <BreadcrumbSeparator className="hidden md:block" />}
            </Fragment>
          );
        })}
      </BreadcrumbList>
    </Breadcrumb>
  );
}
