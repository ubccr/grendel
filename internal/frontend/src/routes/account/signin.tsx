import { postV1AuthSigninMutation } from "@/client/@tanstack/react-query.gen";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useUser } from "@/hooks/user-provider";
import { useForm } from "@tanstack/react-form";
import { useMutation } from "@tanstack/react-query";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { LoaderCircle } from "lucide-react";
import { toast } from "sonner";
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
  const { mutate, isPending } = useMutation(postV1AuthSigninMutation());
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
          onSuccess: (data) => {
            User.setUser({
              username: data?.username ?? "",
              role: data?.role ?? "",
              expire: data?.expire ?? 0,
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
              {isPending ? <LoaderCircle className="animate-spin" /> : <span>Submit</span>}
            </Button>
          </CardFooter>
        </Card>
      </form>
    </div>
  );
}
