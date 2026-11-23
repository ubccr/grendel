import {
  deleteV1AuthSignoutMutation,
  patchV1AuthResetMutation,
} from "@/client/@tanstack/react-query.gen";
import { AccountPanel, SideNote } from "@/components/account-panel";
import { Button } from "@/components/ui/button";
import { Field, FieldDescription, FieldError, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { useUser } from "@/hooks/user-provider";
import { authErrorMessage } from "@/lib/auth-form";
import { useForm } from "@tanstack/react-form";
import { useMutation } from "@tanstack/react-query";
import { createFileRoute, useRouter } from "@tanstack/react-router";
import { LoaderCircle } from "lucide-react";
import { toast } from "sonner";
import { z } from "zod";

export const Route = createFileRoute("/account/_account/reset")({
  staticData: { crumb: "Password" },
  component: RouteComponent,
});

// the API refuses a new password shorter than this
const MIN_PASSWORD = 8;

const schema = z
  .object({
    current: z.string().min(1, "Enter your current password"),
    password: z.string().min(MIN_PASSWORD, `Use at least ${MIN_PASSWORD} characters`),
    confirm: z.string(),
  })
  .refine((v) => v.password === v.confirm, {
    message: "The new passwords don't match",
    path: ["confirm"],
  })
  .refine((v) => !v.current || v.password !== v.current, {
    message: "Choose a password different from your current one",
    path: ["password"],
  });

type Name = "current" | "password" | "confirm";

const fields: { name: Name; label: string; autoComplete: string; hint?: string }[] = [
  { name: "current", label: "Current password", autoComplete: "current-password" },
  {
    name: "password",
    label: "New password",
    autoComplete: "new-password",
    hint: `At least ${MIN_PASSWORD} characters.`,
  },
  { name: "confirm", label: "Confirm new password", autoComplete: "new-password" },
];

function RouteComponent() {
  const reset = useMutation({
    ...patchV1AuthResetMutation(),
    meta: { quietError: true, changesNothing: true },
  });
  const signout = useMutation({
    ...deleteV1AuthSignoutMutation(),
    meta: { quietError: true, changesNothing: true },
  });
  const { user, signOut } = useUser();
  const router = useRouter();

  const form = useForm({
    defaultValues: { current: "", password: "", confirm: "" },
    validators: { onSubmit: schema },
    onSubmit: ({ value }) => {
      reset.mutate(
        { body: { current_password: value.current, new_password: value.password } },
        {
          // the session ends with the old password, so the person signs in again with the new one. If the server can't end it, the session still works until it expires, so they're told rather than shown as signed out
          onSuccess: () =>
            signout.mutate(
              {},
              {
                onSuccess: () => {
                  signOut();
                  toast.success("Password changed", {
                    description: "Sign in with your new password.",
                  });
                  router.navigate({ to: "/account/signin", replace: true });
                },
                onError: (e) =>
                  toast.warning("Password changed, but you're still signed in", {
                    description: `Sign out from the account menu. ${e.detail ?? ""}`.trim(),
                  }),
              },
            ),
        },
      );
    },
  });

  const pending = reset.isPending || signout.isPending;

  return (
    <AccountPanel
      title="Change password"
      description=""
      aside={
        <SideNote title="Info">
          <p>Changing your password will sign you out immediately.</p>
          <p>API tokens are not affected by password changes.</p>
        </SideNote>
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
        {/* lets a password manager tie the new password to this account */}
        <input
          type="text"
          name="username"
          autoComplete="username"
          value={user?.username ?? ""}
          readOnly
          hidden
        />
        <FieldGroup className="gap-5">
          {fields.map((f) => (
            <form.Field
              key={f.name}
              name={f.name}
              children={(field) => {
                const invalid = field.state.meta.errors.length > 0;
                return (
                  <Field data-invalid={invalid}>
                    <FieldLabel htmlFor={field.name}>{f.label}</FieldLabel>
                    <Input
                      id={field.name}
                      name={field.name}
                      type="password"
                      autoComplete={f.autoComplete}
                      autoFocus={f.name === "current"}
                      aria-invalid={invalid}
                      aria-describedby={f.hint && !invalid ? `${field.name}-hint` : undefined}
                      value={field.state.value}
                      onBlur={field.handleBlur}
                      onChange={(e) => field.handleChange(e.target.value)}
                    />
                    {invalid ? (
                      <FieldError errors={field.state.meta.errors} />
                    ) : (
                      f.hint && (
                        <FieldDescription id={`${field.name}-hint`}>{f.hint}</FieldDescription>
                      )
                    )}
                  </Field>
                );
              }}
            />
          ))}
          {reset.isError && (
            <p role="alert" className="text-sm text-destructive">
              {authErrorMessage(reset.error, { 401: "Your current password is incorrect." })}
            </p>
          )}
          <div>
            <Button type="submit" disabled={pending}>
              {pending && <LoaderCircle className="animate-spin" />}
              {pending ? "Changing password" : "Change password"}
            </Button>
          </div>
        </FieldGroup>
      </form>
    </AccountPanel>
  );
}
