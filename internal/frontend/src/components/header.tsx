import { UserRound } from "lucide-react";
import { SidebarTrigger } from "./ui/sidebar";
import { useUser } from "@/hooks/user-provider";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuGroup,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "./ui/dropdown-menu";
import { Link } from "@tanstack/react-router";
import { useDeleteV1AuthSignout } from "@/openapi/queries";
import { toast } from "sonner";
import { Button } from "./ui/button";
import { ThemeSwitch } from "./theme-switch";
import { useRouter } from "@tanstack/react-router";

export default function Header() {
  const { user, setUser } = useUser();
  const router = useRouter();
  const logout_mutation = useDeleteV1AuthSignout();

  return (
    <div className="p-2 px-0">
      <div className="border-sidebar-border bg-sidebar flex justify-between gap-3 rounded-md border p-1 align-middle">
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
                <div className="bg-secondary text-primary flex aspect-square size-9 items-center justify-center rounded-lg">
                  <UserRound className="size-4" />
                </div>
              </DropdownMenuTrigger>
              <DropdownMenuContent className="w-56" align="start">
                <div className="flex gap-3 p-2">
                  <div className="bg-secondary text-primary flex aspect-square size-11 items-center justify-center rounded-lg">
                    <UserRound className="size-4" />
                  </div>
                  <div className="flex flex-col">
                    <span className="">{user?.username}</span>
                    <span className="text-muted-foreground text-sm">
                      {user?.role}
                    </span>
                  </div>
                </div>
                <DropdownMenuSeparator />
                <DropdownMenuLabel>Account</DropdownMenuLabel>
                <DropdownMenuGroup>
                  <DropdownMenuItem className="text-muted-foreground" asChild>
                    <Link to="/account/token">API Token</Link>
                  </DropdownMenuItem>
                  <DropdownMenuItem className="text-muted-foreground" asChild>
                    <Link to="/account/reset">Change Password</Link>
                  </DropdownMenuItem>
                  <DropdownMenuItem
                    className="text-muted-foreground"
                    onClick={() => {
                      logout_mutation.mutate(
                        {},
                        {
                          onSuccess: ({ data }) => {
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
