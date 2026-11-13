import { postV1AuthSignupMutation } from "@/client/@tanstack/react-query.gen";
import { AuthCard } from "@/components/auth-card";
import { Button } from "@/components/ui/button";
import { Field, FieldDescription, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { useUser } from "@/hooks/user-provider";
import { authErrorMessage } from "@/lib/auth-form";
import { useForm } from "@tanstack/react-form";
import { useMutation } from "@tanstack/react-query";
import { createFileRoute, Link, useRouter } from "@tanstack/react-router";
import { LoaderCircle } from "lucide-react";
import { toast } from "sonner";
import { z } from "zod";
import { LOGIN_REDIRECT_FALLBACK } from "./signin";

// the limits match the API's own validation of a signup, which rejects anything shorter
const MIN_USERNAME = 2;
const MIN_PASSWORD = 8;

const schema = z
  .object({
    username: z.string().trim().min(MIN_USERNAME, `Use at least ${MIN_USERNAME} characters`),
    password: z.string().min(MIN_PASSWORD, `Use at least ${MIN_PASSWORD} characters`),
    confirmPassword: z.string(),
  })
  .refine((v) => v.password === v.confirmPassword, {
    message: "The passwords don't match",
    path: ["confirmPassword"],
  });

export const Route = createFileRoute("/account/signup")({
  staticData: { nodeField: "full", crumb: "Sign up" },
  component: RouteComponent,
  validateSearch: z.object({
    redirect: z.string().optional(),
  }),
});

function RouteComponent() {
  const signup = useMutation(postV1AuthSignupMutation());
  const User = useUser();
  const search = Route.useSearch();
  const router = useRouter();
  const navigate = Route.useNavigate();

  const form = useForm({
    defaultValues: {
      username: "",
      password: "",
      confirmPassword: "",
    },
    validators: { onSubmit: schema },
    onSubmit: async ({ value }) => {
      signup.mutate(
        { body: { username: value.username.trim(), password: value.password } },
        {
          onSuccess: (data) => {
            User.setUser({
              username: data?.username ?? "",
              role: data?.role ?? "",
              expire: data?.expire ?? 0,
            });
            // the first account on a new install becomes the administrator; every later one waits for an administrator to enable it
            if (data?.role === "admin") {
              toast.success("Account created. You're the administrator.");
            } else {
              toast.success("Account created", {
                description: "Ask a Grendel administrator to enable it.",
              });
            }
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
      title="Create an account"
      description="New accounts start disabled. After you sign up, ask a Grendel administrator to enable yours."
      footer={
        <>
          Already have an account?{" "}
          <Link
            to="/account/signin"
            search={{ redirect: search.redirect }}
            className="font-medium text-foreground underline underline-offset-4 hover:text-primary"
          >
            Sign in
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
                    autoComplete="new-password"
                    aria-invalid={invalid}
                    aria-describedby={invalid ? undefined : `${field.name}-hint`}
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                  />
                  {invalid ? (
                    <FieldError errors={field.state.meta.errors} />
                  ) : (
                    <FieldDescription id={`${field.name}-hint`}>
                      At least {MIN_PASSWORD} characters.
                    </FieldDescription>
                  )}
                </Field>
              );
            }}
          />
          <form.Field
            name="confirmPassword"
            children={(field) => {
              const invalid = field.state.meta.errors.length > 0;
              return (
                <Field data-invalid={invalid}>
                  <FieldLabel htmlFor={field.name}>Confirm password</FieldLabel>
                  <Input
                    id={field.name}
                    name={field.name}
                    type="password"
                    autoComplete="new-password"
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
          {signup.isError && (
            <p role="alert" className="text-sm text-destructive">
              {authErrorMessage(signup.error, { 409: "That username is taken. Choose another." })}
            </p>
          )}
          <Button type="submit" size="lg" className="mt-1 w-full" disabled={signup.isPending}>
            {signup.isPending && <LoaderCircle className="animate-spin" />}
            {signup.isPending ? "Creating account" : "Create account"}
          </Button>
        </FieldGroup>
      </form>
    </AuthCard>
  );
}
