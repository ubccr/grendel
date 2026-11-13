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
import { Link, useRouter } from "@tanstack/react-router";
import { ChevronsUpDown, KeyRound, LockKeyhole, LogIn, LogOut } from "lucide-react";
import { toast } from "sonner";
import { ThemeIcon, ThemeRadioGroup } from "./theme-switch";

// NavUser sits in the sidebar footer: the signed in user and their account menu, which holds the theme, or a way to sign in with the theme beside it
export function NavUser() {
  const { user, setUser } = useUser();
  const { isMobile } = useSidebar();
  const router = useRouter();
  const { mutate } = useMutation(deleteV1AuthSignoutMutation());

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
          setUser(null);
          toast.success(data?.title, { description: data?.detail });
          router.invalidate();
        },
        onError: (e) => toast.error(e.title, { description: e.detail }),
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
              <DropdownMenuItem render={<Link to="/account/token" />}>
                <KeyRound />
                API token
              </DropdownMenuItem>
              <DropdownMenuItem render={<Link to="/account/reset" />}>
                <LockKeyhole />
                Change password
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
