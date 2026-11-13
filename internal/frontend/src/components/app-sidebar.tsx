import {
  ClipboardList,
  Images,
  type LucideIcon,
  Plus,
  Server,
  ShieldUser,
  SquareMenu,
  UserPen,
} from "lucide-react";

import favicon from "@/assets/favicon.ico";
import { NavUser } from "@/components/nav-user";
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuAction,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarRail,
  useSidebar,
} from "@/components/ui/sidebar";
import { useUser } from "@/hooks/user-provider";
import "@fontsource/unifrakturcook/latin-700.css";
import { Link, type LinkProps, useMatchRoute, useRouter } from "@tanstack/react-router";
import { useEffect } from "react";

type NavItem = {
  title: string;
  url: LinkProps["to"];
  icon: LucideIcon;
  add?: { title: string; url: LinkProps["to"] };
};
type NavGroup = { label: string; items: NavItem[] };

const fleet: NavGroup = {
  label: "Fleet",
  items: [
    { title: "Racks", url: "/racks", icon: SquareMenu },
    { title: "Nodes", url: "/nodes", icon: Server, add: { title: "Add node", url: "/add/node" } },
    {
      title: "Images",
      url: "/images",
      icon: Images,
      add: { title: "Add image", url: "/add/image" },
    },
    { title: "Events", url: "/events", icon: ClipboardList },
  ],
};

const administration: NavGroup = {
  label: "Administration",
  items: [
    { title: "Roles", url: "/account/roles", icon: ShieldUser },
    { title: "Users", url: "/account/users", icon: UserPen },
  ],
};

export function AppSidebar() {
  const { user } = useUser();
  const { setOpenMobile } = useSidebar();
  const router = useRouter();
  const matchRoute = useMatchRoute();
  const groups = user?.role === "admin" ? [fleet, administration] : [fleet];
  // on mobile the sidebar is a sheet over the page, so any navigation closes it, whether from the nav, the account menu or signing out
  useEffect(
    () => router.subscribe("onResolved", () => setOpenMobile(false)),
    [router, setOpenMobile],
  );

  return (
    <Sidebar collapsible="icon" variant="floating" side="left">
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton size="lg" tooltip="Home" render={<Link to="/" />}>
              <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-sidebar-accent">
                <img src={favicon} alt="" className="size-5" />
              </div>
              <div className="grid flex-1 leading-tight">
                <span className="truncate font-[UnifrakturCook] text-2xl leading-none">
                  Grendel
                </span>
                <span className="truncate text-xs text-muted-foreground">
                  Bare-metal provisioning
                </span>
              </div>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        {groups.map((group) => (
          <SidebarGroup key={group.label}>
            <SidebarGroupLabel>{group.label}</SidebarGroupLabel>
            <SidebarGroupContent>
              <SidebarMenu>
                {group.items.map((item) => (
                  <SidebarMenuItem key={item.url}>
                    <SidebarMenuButton
                      tooltip={item.title}
                      isActive={!!matchRoute({ to: item.url, fuzzy: true })}
                      render={<Link to={item.url} />}
                    >
                      <item.icon />
                      <span>{item.title}</span>
                    </SidebarMenuButton>
                    {item.add && (
                      <SidebarMenuAction title={item.add.title} render={<Link to={item.add.url} />}>
                        <Plus />
                        <span className="sr-only">{item.add.title}</span>
                      </SidebarMenuAction>
                    )}
                  </SidebarMenuItem>
                ))}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        ))}
      </SidebarContent>
      <SidebarFooter>
        <NavUser />
      </SidebarFooter>
      <SidebarRail />
    </Sidebar>
  );
}
