import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useUser } from "@/hooks/user-provider";
import { usePostV1AuthSignup } from "@/openapi/queries";
import { useForm } from "@tanstack/react-form";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { LoaderCircle } from "lucide-react";
import { toast } from "sonner";
import { LOGIN_REDIRECT_FALLBACK } from "./signin";

export const Route = createFileRoute("/account/signup")({
  component: RouteComponent,
});

function RouteComponent() {
  const { mutate, isPending } = usePostV1AuthSignup();
  const User = useUser();
  const router = useRouter();

  const form = useForm({
    defaultValues: {
      username: "",
      password: "",
      confirmPassword: "",
    },
    onSubmit: async ({ value }) => {
      if (value.password !== value.confirmPassword) {
        toast.error("Passwords do not match");
        return;
      }
      mutate(
        { body: { username: value.username, password: value.password } },
        {
          onSuccess: (e) => {
            User.setUser({
              username: e.data?.username ?? "",
              role: e.data?.role ?? "",
              expire: 0,
            });
            toast.success("Successfully created an account");
            router.history.push(LOGIN_REDIRECT_FALLBACK);
          },
          onError: () => {
            toast.error("Failed to create an account");
          },
        }
      );
    },
  });
  return (
    <div className="flex justify-center">
      <form
        className="grid grid-cols-1 gap-4"
        onSubmit={(e) => {
          e.preventDefault();
          e.stopPropagation();
          form.handleSubmit();
        }}
      >
        <Card>
          <CardHeader>
            <CardTitle>Create an Account:</CardTitle>
            <CardDescription>
              New to Grendel? Create your account then ask your administrator to
              enable it
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form.Field
              name="username"
              children={(field) => (
                <div>
                  <Label>Username:</Label>
                  <Input
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                  />
                </div>
              )}
            />
            <form.Field
              name="password"
              children={(field) => (
                <div>
                  <Label>Password:</Label>
                  <Input
                    type="password"
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                  />
                </div>
              )}
            />
            <form.Field
              name="confirmPassword"
              children={(field) => (
                <div>
                  <Label>Confirm Password:</Label>
                  <Input
                    type="password"
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                  />
                </div>
              )}
            />
          </CardContent>
          <CardFooter>
            <Button type="submit">
              {isPending ? (
                <LoaderCircle className="animate-spin" />
              ) : (
                <span>Submit</span>
              )}
            </Button>
          </CardFooter>
        </Card>
      </form>
    </div>
  );
}
