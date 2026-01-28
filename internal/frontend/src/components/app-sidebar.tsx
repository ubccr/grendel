import {
  ClipboardList,
  Home,
  Images,
  Plus,
  SearchIcon,
  Server,
  ShieldUser,
  SquareMenu,
  UserPen,
} from "lucide-react";

import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuAction,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";
import { Link } from "@tanstack/react-router";

import favicon from "@/assets/favicon.ico";
import { useUser } from "../hooks/user-provider";

export function AppSidebar() {
  const { user } = useUser();

  const items = [
    {
      label: "",
      items: [
        {
          title: "Home",
          url: "/",
          icon: Home,
        },
        {
          title: "Racks",
          url: "/racks",
          icon: SquareMenu,
        },
        {
          title: "Nodes",
          url: "/nodes",
          icon: Server,
          action: {
            title: "Add Node",
            url: "/add/node",
            icon: Plus,
          },
        },
        {
          title: "Images",
          url: "/images",
          icon: Images,
          action: {
            title: "Add Node",
            url: "/add/image",
            icon: Plus,
          },
        },
        {
          title: "Events",
          url: "/events",
          icon: ClipboardList,
        },
        {
          title: "Inventory",
          url: "/search/inventory",
          icon: SearchIcon,
        },
      ],
    },
  ];

  if (user?.role === "admin") {
    items.push({
      label: "Administration",
      items: [
        {
          title: "Roles",
          url: "/account/roles",
          icon: ShieldUser,
        },
        {
          title: "Users",
          url: "/account/users",
          icon: UserPen,
        },
      ],
    });
  }
  return (
    <Sidebar collapsible="icon" variant="floating" side="left">
      <SidebarHeader>
        <SidebarMenuButton asChild>
          <Link to="/">
            <img src={favicon} className="h-6 w-4" />
            <span>Grendel</span>
          </Link>
        </SidebarMenuButton>
      </SidebarHeader>
      <SidebarContent>
        {items.map((group, g) => (
          <SidebarGroup key={g}>
            {group.label && <SidebarGroupLabel>{group.label}</SidebarGroupLabel>}
            <SidebarGroupContent>
              <SidebarMenu>
                {group.items.map((item, x) => (
                  <SidebarMenuItem key={x}>
                    <SidebarMenuButton asChild>
                      <Link to={item.url} activeProps={{ className: "font-bold border" }}>
                        <item.icon />
                        <span>{item.title}</span>
                      </Link>
                    </SidebarMenuButton>
                    {item.action && (
                      <SidebarMenuAction asChild>
                        <Link to={item.action.url}>
                          <item.action.icon />
                          <span className="sr-only">{item.action.title}</span>
                        </Link>
                      </SidebarMenuAction>
                    )}
                  </SidebarMenuItem>
                ))}
              </SidebarMenu>
            </SidebarGroupContent>
          </SidebarGroup>
        ))}
      </SidebarContent>
    </Sidebar>
  );
}
