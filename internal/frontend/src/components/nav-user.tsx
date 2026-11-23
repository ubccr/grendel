import { deleteV1AuthSignoutMutation } from "@/client/@tanstack/react-query.gen";
import { Avatar, AvatarFallback } from "@/components/ui/avatar";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuSub,
  DropdownMenuSubContent,
  DropdownMenuSubTrigger,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import {
  SidebarMenu,
  SidebarMenuAction,
  SidebarMenuButton,
  SidebarMenuItem,
  useSidebar,
} from "@/components/ui/sidebar";
import { useUser } from "@/hooks/user-provider";
import { useMutation } from "@tanstack/react-query";
import { Link } from "@tanstack/react-router";
import { ChevronsUpDown, Info, LogIn, LogOut, UserRound } from "lucide-react";
import { toast } from "sonner";
import { ThemeIcon, ThemeRadioGroup } from "./theme-switch";

export function NavUser() {
  const { user, signOut: endSession } = useUser();
  const { isMobile } = useSidebar();
  const { mutate } = useMutation({
    ...deleteV1AuthSignoutMutation(),
    meta: { changesNothing: true },
  });

  if (!user) {
    return (
      <SidebarMenu>
        <SidebarMenuItem>
          <SidebarMenuButton tooltip="Sign in" render={<Link to="/account/signin" />}>
            <LogIn />
            <span>Sign in</span>
          </SidebarMenuButton>
          <DropdownMenu>
            <DropdownMenuTrigger render={<SidebarMenuAction title="Theme" />}>
              <ThemeIcon />
              <span className="sr-only">Theme</span>
            </DropdownMenuTrigger>
            <DropdownMenuContent side={isMobile ? "bottom" : "right"} align="end" sideOffset={8}>
              <ThemeRadioGroup />
            </DropdownMenuContent>
          </DropdownMenu>
        </SidebarMenuItem>
      </SidebarMenu>
    );
  }

  const signOut = () =>
    mutate(
      {},
      {
        onSuccess: (data) => {
          endSession();
          toast.success(data?.title, { description: data?.detail });
        },
      },
    );

  const identity = (
    <>
      <Avatar className="rounded-lg after:rounded-lg">
        <AvatarFallback className="rounded-lg bg-sidebar-accent font-medium text-sidebar-accent-foreground">
          {user.username?.slice(0, 2).toUpperCase()}
        </AvatarFallback>
      </Avatar>
      <div className="grid flex-1 text-left text-sm leading-tight">
        <span className="truncate font-medium">{user.username}</span>
        <span className="truncate text-xs text-muted-foreground">{user.role}</span>
      </div>
    </>
  );

  return (
    <SidebarMenu>
      <SidebarMenuItem>
        <DropdownMenu>
          <DropdownMenuTrigger
            render={<SidebarMenuButton size="lg" className="data-popup-open:bg-sidebar-accent" />}
          >
            {identity}
            <ChevronsUpDown className="ml-auto" />
          </DropdownMenuTrigger>
          <DropdownMenuContent
            className="min-w-56"
            side={isMobile ? "bottom" : "right"}
            align="end"
            sideOffset={8}
          >
            <DropdownMenuGroup>
              <DropdownMenuLabel className="flex items-center gap-2 px-1 py-1.5 font-normal text-foreground">
                {identity}
              </DropdownMenuLabel>
            </DropdownMenuGroup>
            <DropdownMenuSeparator />
            <DropdownMenuGroup>
              <DropdownMenuItem render={<Link to="/account" />}>
                <UserRound />
                Account
              </DropdownMenuItem>
              <DropdownMenuSub>
                <DropdownMenuSubTrigger>
                  <ThemeIcon />
                  Theme
                </DropdownMenuSubTrigger>
                <DropdownMenuSubContent>
                  <ThemeRadioGroup />
                </DropdownMenuSubContent>
              </DropdownMenuSub>
              <DropdownMenuItem render={<Link to="/about" />}>
                <Info />
                About Grendel
              </DropdownMenuItem>
            </DropdownMenuGroup>
            <DropdownMenuSeparator />
            <DropdownMenuItem onClick={signOut}>
              <LogOut />
              Sign out
            </DropdownMenuItem>
          </DropdownMenuContent>
        </DropdownMenu>
      </SidebarMenuItem>
    </SidebarMenu>
  );
}
