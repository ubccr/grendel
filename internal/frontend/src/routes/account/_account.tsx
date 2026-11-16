import { DetailHeader, DetailTabs, Fact } from "@/components/detail-header";
import { Card, CardContent } from "@/components/ui/card";
import { useNow } from "@/hooks/use-now";
import { useUser } from "@/hooks/user-provider";
import AuthRedirect from "@/lib/auth";
import { relativeTime } from "@/lib/utils";
import { createFileRoute, Outlet } from "@tanstack/react-router";

// the signed in account's pages share this layout, which adds nothing to their paths, so they are headed by the account as a node's pages are headed by the node
export const Route = createFileRoute("/account/_account")({
  staticData: { crumb: "Account" },
  component: RouteComponent,
  beforeLoad: AuthRedirect,
});

const tabs = [
  { to: "/account/reset", label: "Password" },
  { to: "/account/token", label: "API tokens" },
] as const;

function RouteComponent() {
  const { user } = useUser();
  const now = useNow(30_000);
  const ends = user?.expire ? relativeTime(user.expire, { now }) : undefined;

  return (
    <Card>
      <CardContent className="p-0">
        <DetailHeader
          title={user?.username ?? ""}
          facts={
            <>
              <Fact label="Role">{user?.role}</Fact>
              {ends && (
                <Fact label="Session ends">
                  <span title={ends.title}>{ends.text}</span>
                </Fact>
              )}
            </>
          }
          nav={<DetailTabs label="Account" tabs={tabs} />}
        />
        <div className="px-6 py-6">
          <Outlet />
        </div>
      </CardContent>
    </Card>
  );
}
