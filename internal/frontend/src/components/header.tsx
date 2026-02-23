import { deleteV1AuthSignoutMutation } from "@/client/@tanstack/react-query.gen";
import { useUser } from "@/hooks/user-provider";
import { useMutation } from "@tanstack/react-query";
import { Link, useRouter } from "@tanstack/react-router";
import { UserRound } from "lucide-react";
import { toast } from "sonner";
import { ThemeSwitch } from "./theme-switch";
import { Button } from "./ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "./ui/dropdown-menu";
import { SidebarTrigger } from "./ui/sidebar";

export default function Header() {
  const { user, setUser } = useUser();
  const router = useRouter();
  const { mutate } = useMutation(deleteV1AuthSignoutMutation());

  return (
    <div className="p-2 px-0">
      <div className="flex justify-between gap-3 rounded-md border border-sidebar-border bg-sidebar p-1 align-middle">
        <SidebarTrigger className="my-auto" />
        <div className="flex gap-2">
          <ThemeSwitch />
          {!user ? (
            <div className="flex gap-2">
              <Button variant="secondary" asChild>
                <Link to="/account/signin">Login</Link>
              </Button>
              <Button variant="secondary" asChild>
                <Link to="/account/signup">Signup</Link>
              </Button>
            </div>
          ) : (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <div className="flex aspect-square size-9 cursor-pointer items-center justify-center rounded-lg bg-secondary text-primary">
                  <UserRound className="size-4" />
                </div>
              </DropdownMenuTrigger>
              <DropdownMenuContent className="w-56" align="start">
                <div className="flex gap-3 p-2">
                  <div className="flex aspect-square size-11 items-center justify-center rounded-lg bg-secondary text-primary">
                    <UserRound className="size-4" />
                  </div>
                  <div className="flex flex-col">
                    <span className="">{user?.username}</span>
                    <span className="text-sm text-muted-foreground">{user?.role}</span>
                  </div>
                </div>
                <DropdownMenuSeparator />
                <DropdownMenuLabel>Account</DropdownMenuLabel>
                <DropdownMenuGroup>
                  <DropdownMenuItem className="cursor-pointer text-muted-foreground" asChild>
                    <Link to="/account/token">API Token</Link>
                  </DropdownMenuItem>
                  <DropdownMenuItem className="cursor-pointer text-muted-foreground" asChild>
                    <Link to="/account/reset">Change Password</Link>
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    className="cursor-pointer text-muted-foreground"
                    onClick={() => {
                      mutate(
                        {},
                        {
                          onSuccess: (data) => {
                            setUser(null);
                            toast.success(data?.title, {
                              description: data?.detail,
                            });
                            router.invalidate();
                          },
                          onError: (e) => {
                            toast.error(e.title, {
                              description: e.detail,
                            });
                          },
                        },
                      );
                    }}
                  >
                    Logout
                  </DropdownMenuItem>
                </DropdownMenuGroup>
              </DropdownMenuContent>
            </DropdownMenu>
          )}
        </div>
      </div>
    </div>
  );
}
