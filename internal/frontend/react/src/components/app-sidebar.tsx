import {
    ChevronUp,
    Grid3x3,
    Group,
    Home,
    Images,
    LayoutTemplate,
    Network,
    Plus,
    Search,
    Server,
    SmartphoneCharging,
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
import { ModeToggle } from "./mode-toggle";

export function AppSidebar() {
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
            title: "Node Groups",
            url: "/groups/nodes",
            icon: Group,
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
            title: "Templates",
            url: "/templates",
            icon: LayoutTemplate,
            action: {
                title: "Add Node",
                url: "/add/template",
                icon: Plus,
            },
        },
        {
            title: "Network",
            url: "/network",
            icon: Network,
        },
        {
            title: "Power",
            url: "/power",
            icon: SmartphoneCharging,
        },
        {
            title: "Search",
            url: "/search",
            icon: Search,
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
            </SidebarContent>
            <SidebarFooter>
                <SidebarMenu>
                    <SidebarMenuItem>
                        <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                                <SidebarMenuButton>
                                    <UserRound />
                                    <span>Username</span>
                                    <ChevronUp className="ml-auto" />
                                </SidebarMenuButton>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent side="top" className="w-[--radix-popper-anchor-width]">
                                <DropdownMenuItem>
                                    <span>Account</span>
                                </DropdownMenuItem>
                                <DropdownMenuItem>
                                    <span>Settings</span>
                                </DropdownMenuItem>
                                <DropdownMenuItem>
                                    <ModeToggle />
                                </DropdownMenuItem>
                                <DropdownMenuItem>
                                    <span>Sign out</span>
                                </DropdownMenuItem>
                            </DropdownMenuContent>
                        </DropdownMenu>
                    </SidebarMenuItem>
                </SidebarMenu>
            </SidebarFooter>
        </Sidebar>
    );
}
