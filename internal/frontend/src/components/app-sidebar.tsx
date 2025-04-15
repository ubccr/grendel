import {
  ChevronUp,
  ClipboardList,
  Grid3x3,
  Home,
  Images,
  LoaderCircle,
  Plus,
  Server,
  UserRound,
} from "lucide-react";

import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarGroup,
  SidebarGroupContent,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuAction,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Link } from "@tanstack/react-router";

import favicon from "@/assets/favicon.ico";
import { ModeToggle } from "../hooks/mode-toggle";
import { useUser } from "../hooks/user-provider";
import { toast } from "sonner";
import { useDeleteV1AuthSignout } from "@/openapi/queries";

export function AppSidebar() {
  const { user, setUser } = useUser();
  const logout_mutation = useDeleteV1AuthSignout();

  const items = [
    {
      title: "Home",
      url: "/",
      icon: Home,
    },
    {
      title: "Floorplan",
      url: "/floorplan",
      icon: Grid3x3,
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
  ];
  return (
    <Sidebar collapsible="icon">
      <SidebarHeader>
        <SidebarMenuButton asChild>
          <Link to="/">
            <img src={favicon} className="h-6 w-4" />
            <span>Grendel</span>
          </Link>
        </SidebarMenuButton>
      </SidebarHeader>
      <SidebarContent>
        <SidebarGroup>
          <SidebarGroupContent>
            <SidebarMenu>
              {items.map((item, x) => (
                <SidebarMenuItem key={x}>
                  <SidebarMenuButton asChild>
                    <Link
                      to={item.url}
                      activeProps={{ className: "font-bold border" }}
                    >
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
      </SidebarContent>
      <SidebarFooter>
        <SidebarMenu>
          <SidebarMenuItem>
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <SidebarMenuButton
                  size="lg"
                  className="data-[state=open]:bg-sidebar-accent data-[state=open]:text-sidebar-accent-foreground"
                >
                  <div className="flex aspect-square size-8 items-center justify-center rounded-lg bg-sidebar-primary text-sidebar-primary-foreground">
                    <UserRound className="size-4" />
                  </div>
                  <div className="flex flex-col gap-0.5 leading-none px-2">
                    <span
                      className={`transition-opacity duration-700 ease-in-out font-semibold ${
                        user ? "opacity-100" : "opacity-0"
                      }`}
                    >
                      {user?.username}
                    </span>
                    <span
                      className={`transition-opacity duration-700 ease-in-out text-xs text-muted-foreground ${
                        user ? "opacity-100" : "opacity-0"
                      }`}
                    >
                      Role: {user?.role}
                    </span>
                  </div>
                  <ChevronUp className="ml-auto" />
                </SidebarMenuButton>
              </DropdownMenuTrigger>
              <DropdownMenuContent
                side="top"
                className="w-(--radix-popper-anchor-width)"
              >
                <DropdownMenuItem>
                  <ModeToggle />
                </DropdownMenuItem>
                {user?.role === "admin" && (
                  <DropdownMenuItem asChild>
                    <Link to="/account/users">Users</Link>
                  </DropdownMenuItem>
                )}
                {user?.role === "admin" && (
                  <DropdownMenuItem asChild>
                    <Link to="/account/roles">Roles</Link>
                  </DropdownMenuItem>
                )}
                {user && (
                  <DropdownMenuItem asChild>
                    <Link to="/account/token">API Token</Link>
                  </DropdownMenuItem>
                )}
                {user && (
                  <DropdownMenuItem asChild>
                    <Link to="/account/reset">Change Password</Link>
                  </DropdownMenuItem>
                )}
                {user && (
                  <DropdownMenuItem
                    onClick={() => {
                      logout_mutation.mutate(
                        {},
                        {
                          onSuccess: ({ data }) => {
                            setUser(null);
                            toast.success(data?.title, {
                              description: data?.detail,
                            });
                          },
                          onError: (e) => {
                            toast.error(e.title, {
                              description: e.detail,
                            });
                          },
                        }
                      );
                    }}
                  >
                    {logout_mutation.isPending ? (
                      <LoaderCircle className="animate-spin" />
                    ) : (
                      "Logout"
                    )}
                  </DropdownMenuItem>
                )}
                {!user && (
                  <DropdownMenuItem asChild>
                    <Link to="/account/signup">Signup</Link>
                  </DropdownMenuItem>
                )}
                {!user && (
                  <DropdownMenuItem asChild>
                    <Link to="/account/signin">Login</Link>
                  </DropdownMenuItem>
                )}
              </DropdownMenuContent>
            </DropdownMenu>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarFooter>
    </Sidebar>
  );
}
