import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useUser } from "@/hooks/user-provider";
import { useDeleteV1AuthSignout, usePatchV1AuthReset } from "@/openapi/queries";
import { useForm } from "@tanstack/react-form";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { toast } from "sonner";

export const Route = createFileRoute("/account/reset")({
  component: RouteComponent,
});

function RouteComponent() {
  const { mutate } = usePatchV1AuthReset();
  const logout = useDeleteV1AuthSignout();
  const User = useUser();
  const router = useRouter();

  const form = useForm({
    defaultValues: {
      current_password: "",
      new_password: "",
      confirm_password: "",
    },
    onSubmit(props) {
      const v = props.value;

      if (v.new_password !== v.confirm_password) {
        toast.error("Passwords do not match");
        return;
      }

      mutate(
        {
          body: {
            current_password: v.current_password,
            new_password: v.new_password,
          },
        },
        {
          onSuccess: ({ data }) => {
            toast.success(data?.title, {
              description: data?.detail,
            });
            logout.mutate(
              {},
              {
                onError: (e) => toast.error(e.title, { description: e.detail }),
              }
            );
            User.setUser(null);
            router.navigate({ to: "/" });
          },
          onError: (e) => {
            toast.error(e.title, {
              description: e.detail,
            });
          },
        }
      );
    },
  });
  return (
    <div className="p-4 flex justify-center">
      <Card>
        <CardHeader>
          <CardTitle>Change Password</CardTitle>
          <CardContent>
            <form
              id="resetForm"
              className="grid grid-cols-1 gap-4 mt-3"
              onSubmit={(e) => {
                e.preventDefault();
                e.stopPropagation();
                form.handleSubmit();
              }}
            >
              <form.Field
                name="current_password"
                children={(field) => (
                  <div>
                    <Label>Current Password:</Label>
                    <Input
                      value={field.state.value}
                      onBlur={field.handleBlur}
                      onChange={(e) => field.handleChange(e.target.value)}
                      type="password"
                      autoComplete="current-password"
                    />
                  </div>
                )}
              />
              <form.Field
                name="new_password"
                children={(field) => (
                  <div>
                    <Label>Password:</Label>
                    <Input
                      value={field.state.value}
                      onBlur={field.handleBlur}
                      onChange={(e) => field.handleChange(e.target.value)}
                      type="password"
                      autoComplete="new-password"
                    />
                  </div>
                )}
              />
              <form.Field
                name="confirm_password"
                children={(field) => (
                  <div>
                    <Label>Confirm Password:</Label>
                    <Input
                      value={field.state.value}
                      onBlur={field.handleBlur}
                      onChange={(e) => field.handleChange(e.target.value)}
                      type="password"
                      autoComplete="new-password"
                    />
                  </div>
                )}
              />
            </form>
          </CardContent>
          <CardFooter>
            <Button type="submit" form="resetForm" size="sm">
              Submit
            </Button>
          </CardFooter>
        </CardHeader>
      </Card>
    </div>
  );
}
