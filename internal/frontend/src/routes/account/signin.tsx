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
import { useForm } from "@tanstack/react-form";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { toast } from "sonner";
import { LoaderCircle } from "lucide-react";
import { useUser } from "@/hooks/user-provider";
import { usePostV1AuthSignin } from "@/openapi/queries";
import { z } from "zod";

export const LOGIN_REDIRECT_FALLBACK = "/";

export const Route = createFileRoute("/account/signin")({
  component: RouteComponent,
  validateSearch: z.object({
    redirect: z.string().optional(),
  }),
  // beforeLoad: ({ context, location})
});

function RouteComponent() {
  const { mutate, isPending } = usePostV1AuthSignin();
  const User = useUser();
  const search = Route.useSearch();
  const router = useRouter();
  const navigate = Route.useNavigate();

  const form = useForm({
    defaultValues: {
      username: "",
      password: "",
    },
    onSubmit: async ({ value }) => {
      mutate(
        { body: { username: value.username, password: value.password } },
        {
          onSuccess: (e) => {
            User.setUser({
              username: e.data?.username ?? "",
              role: e.data?.role ?? "",
              expire: e.data?.expire ?? 0,
            });
            toast.success("Successfully authenticated");
            router.invalidate().then(() => {
              navigate({
                to: search.redirect ?? LOGIN_REDIRECT_FALLBACK,
                replace: true,
              });
            });
          },
          onError: (e) => {
            toast.error(e.title, {
              description: e.detail,
            });
          },
        },
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
            <CardTitle>Login:</CardTitle>
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
