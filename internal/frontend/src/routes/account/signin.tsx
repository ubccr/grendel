import { postV1AuthSigninMutation } from "@/client/@tanstack/react-query.gen";
import { AuthCard } from "@/components/auth-card";
import { Button } from "@/components/ui/button";
import { Field, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { useUser } from "@/hooks/user-provider";
import { authErrorMessage } from "@/lib/auth-form";
import { useForm } from "@tanstack/react-form";
import { useMutation } from "@tanstack/react-query";
import { createFileRoute, Link, useRouter } from "@tanstack/react-router";
import { LoaderCircle } from "lucide-react";
import { toast } from "sonner";
import { z } from "zod";

export const LOGIN_REDIRECT_FALLBACK = "/";

const schema = z.object({
  username: z.string().trim().min(1, "Enter your username"),
  password: z.string().min(1, "Enter your password"),
});

export const Route = createFileRoute("/account/signin")({
  staticData: { nodeField: "full", crumb: "Sign in" },
  component: RouteComponent,
  validateSearch: z.object({
    redirect: z.string().optional(),
  }),
});

function RouteComponent() {
  const signin = useMutation(postV1AuthSigninMutation());
  const User = useUser();
  const search = Route.useSearch();
  const router = useRouter();
  const navigate = Route.useNavigate();

  const form = useForm({
    defaultValues: {
      username: "",
      password: "",
    },
    validators: { onSubmit: schema },
    onSubmit: async ({ value }) => {
      signin.mutate(
        { body: { username: value.username, password: value.password } },
        {
          onSuccess: (data) => {
            User.setUser({
              username: data?.username ?? "",
              role: data?.role ?? "",
              expire: data?.expire ?? 0,
            });
            toast.success(`Signed in as ${data?.username ?? value.username}`);
            router.invalidate().then(() => {
              navigate({
                to: search.redirect ?? LOGIN_REDIRECT_FALLBACK,
                replace: true,
              });
            });
          },
        },
      );
    },
  });

  return (
    <AuthCard
      title="Sign in"
      description={"Use your Grendel account to manage bare\u2011metal resources."}
      footer={
        <>
          No account yet?{" "}
          <Link
            to="/account/signup"
            search={{ redirect: search.redirect }}
            className="font-medium text-foreground underline underline-offset-4 hover:text-primary"
          >
            Create one
          </Link>
        </>
      }
    >
      <form
        noValidate
        onSubmit={(e) => {
          e.preventDefault();
          e.stopPropagation();
          form.handleSubmit();
        }}
      >
        <FieldGroup className="gap-4">
          <form.Field
            name="username"
            children={(field) => {
              const invalid = field.state.meta.errors.length > 0;
              return (
                <Field data-invalid={invalid}>
                  <FieldLabel htmlFor={field.name}>Username</FieldLabel>
                  <Input
                    id={field.name}
                    name={field.name}
                    autoComplete="username"
                    autoCapitalize="none"
                    spellCheck={false}
                    autoFocus
                    aria-invalid={invalid}
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                  />
                  <FieldError errors={field.state.meta.errors} />
                </Field>
              );
            }}
          />
          <form.Field
            name="password"
            children={(field) => {
              const invalid = field.state.meta.errors.length > 0;
              return (
                <Field data-invalid={invalid}>
                  <FieldLabel htmlFor={field.name}>Password</FieldLabel>
                  <Input
                    id={field.name}
                    name={field.name}
                    type="password"
                    autoComplete="current-password"
                    aria-invalid={invalid}
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                  />
                  <FieldError errors={field.state.meta.errors} />
                </Field>
              );
            }}
          />
          {signin.isError && (
            <p role="alert" className="text-sm text-destructive">
              {authErrorMessage(signin.error, { 401: "The username or password is incorrect." })}
            </p>
          )}
          <Button type="submit" size="lg" className="mt-1 w-full" disabled={signin.isPending}>
            {signin.isPending && <LoaderCircle className="animate-spin" />}
            {signin.isPending ? "Signing in" : "Sign in"}
          </Button>
        </FieldGroup>
      </form>
    </AuthCard>
  );
}
