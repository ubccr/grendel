import {
  getV1RolesOptions,
  getV1UsersOptions,
  postV1AuthTokenMutation,
} from "@/client/@tanstack/react-query.gen";
import { AccountPanel, SideNote } from "@/components/account-panel";
import { Fact } from "@/components/detail-header";
import { Button } from "@/components/ui/button";
import {
  CodeBlock,
  CodeBlockActions,
  CodeBlockContent,
  CodeBlockCopyButton,
  CodeBlockHeader,
  CodeBlockTitle,
} from "@/components/ui/code-block";
import {
  Combobox,
  ComboboxContent,
  ComboboxEmpty,
  ComboboxInput,
  ComboboxItem,
  ComboboxList,
} from "@/components/ui/combobox";
import { Field, FieldDescription, FieldGroup, FieldLabel } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useUser } from "@/hooks/user-provider";
import { authErrorMessage } from "@/lib/auth-form";
import {
  clientConfig,
  DEFAULT_EXPIRY,
  expiryLabel,
  expiryOptions,
  expirySeconds,
  NEVER,
  parseExpiry,
  tokenClaims,
  tokenParts,
  tokenRoles,
} from "@/lib/tokens";
import { relativeTime } from "@/lib/utils";
import { useForm, useStore } from "@tanstack/react-form";
import { useMutation, useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { FileCog, LoaderCircle } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import { toast } from "sonner";

export const Route = createFileRoute("/account/_account/token")({
  staticData: { crumb: "API tokens" },
  component: RouteComponent,
});

const dateFormat = new Intl.DateTimeFormat(undefined, { dateStyle: "medium", timeStyle: "short" });

// expiryNote says when a token made now with the expiry would stop working
function expiryNote(expiry: string) {
  if (expiry === NEVER) return "The token never expires.";
  const seconds = expirySeconds(expiry) ?? 0;
  return `The token stops working ${dateFormat.format(Date.now() + seconds * 1000)}.`;
}

function RouteComponent() {
  const { user } = useUser();
  const isAdmin = user?.role === "admin";
  const mint = useMutation(postV1AuthTokenMutation());
  // a new request clears the last result, so the token shown is always the one just made
  const token = mint.data?.token;

  // only an administrator may list users or make a token for someone else, so everyone else makes tokens for themselves. Roles are listed for anyone allowed to; for anyone else the choice is their own role or a read-only one, and the server refuses what their account can't hold
  const users = useQuery({ ...getV1UsersOptions(), enabled: isAdmin });
  const roles = useQuery(getV1RolesOptions());

  const form = useForm({
    defaultValues: {
      username: user?.username ?? "",
      role: user?.role ?? "",
      expiry: DEFAULT_EXPIRY,
    },
    onSubmit: ({ value }) =>
      mint.mutate({
        body: { username: value.username, role: value.role, expire: value.expiry },
      }),
  });

  const username = useStore(form.store, (s) => s.values.username);

  // until the users load, an administrator's list holds just themselves, so the field doesn't change shape when it arrives
  const people = users.data
    ? users.data.filter((u) => u.username && u.enabled).map((u) => u.username!)
    : [user?.username ?? ""];
  const userRole = users.data?.find((u) => u.username === username)?.role ?? user?.role ?? "";
  const roleChoices = roles.data?.roles
    ? tokenRoles(roles.data.roles, userRole)
    : [...new Set([userRole, "read-only"])];

  return (
    <AccountPanel
      title="Make a token"
      description="A token authenticates the grendel CLI or a script to the API"
      aside={token ? <TokenSlip token={token} /> : <AboutTokens />}
    >
      <form
        noValidate
        onSubmit={(e) => {
          e.preventDefault();
          e.stopPropagation();
          form.handleSubmit();
        }}
      >
        <FieldGroup className="gap-5">
          <form.Field
            name="username"
            children={(field) => (
              <Field>
                <FieldLabel htmlFor={field.name}>Acts as</FieldLabel>
                {isAdmin ? (
                  <Select
                    value={field.state.value}
                    items={people.map((p) => ({ value: p, label: p }))}
                    onValueChange={(v) => {
                      if (!v) return;
                      field.handleChange(v);
                      // a token can't carry more than its user's role, so choosing someone starts from theirs
                      const role = users.data?.find((u) => u.username === v)?.role;
                      if (role) form.setFieldValue("role", role);
                    }}
                  >
                    <SelectTrigger id={field.name} className="w-full font-mono">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {people.map((p) => (
                        <SelectItem key={p} value={p} className="font-mono">
                          {p}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                ) : (
                  <Input id={field.name} value={field.state.value} disabled className="font-mono" />
                )}
                <FieldDescription>
                  {isAdmin
                    ? "Requests made with the token are logged under this account."
                    : "Only an administrator can make a token for another account."}
                </FieldDescription>
              </Field>
            )}
          />
          <form.Field
            name="role"
            children={(field) => (
              <Field>
                <FieldLabel htmlFor={field.name}>Role</FieldLabel>
                <Select
                  value={field.state.value}
                  items={roleChoices.map((r) => ({ value: r, label: r }))}
                  onValueChange={(v) => v && field.handleChange(v)}
                >
                  <SelectTrigger id={field.name} className="w-full">
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {roleChoices.map((r) => (
                      <SelectItem key={r} value={r}>
                        {r}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
                <FieldDescription>
                  A token can do no more than its account
                  {roles.data ? (
                    <>
                      , so only roles within <span className="text-foreground">{userRole}</span> are
                      offered.
                    </>
                  ) : (
                    "."
                  )}
                </FieldDescription>
              </Field>
            )}
          />
          <form.Field
            name="expiry"
            children={(field) => (
              <Field>
                <FieldLabel htmlFor={field.name}>Expires after</FieldLabel>
                <ExpiryCombobox
                  id={field.name}
                  value={field.state.value}
                  onValueChange={field.handleChange}
                />
                <FieldDescription aria-live="polite">
                  {expiryNote(field.state.value)}
                </FieldDescription>
              </Field>
            )}
          />
          {mint.isError && (
            <p role="alert" className="text-sm text-destructive">
              {authErrorMessage(mint.error, {})}
            </p>
          )}
          <div>
            <Button type="submit" disabled={mint.isPending}>
              {mint.isPending && <LoaderCircle className="animate-spin" />}
              {mint.isPending ? "Making token" : "Make token"}
            </Button>
          </div>
        </FieldGroup>
      </form>
    </AccountPanel>
  );
}

// ExpiryCombobox picks how long a token lasts from common lengths, or takes one typed in, like 45d or 1h30m. Typing a length chooses it straight away, so it counts without being picked from the list, and when the list closes the combobox puts the chosen expiry's name back in the input, which also drops any text that wasn't a length
function ExpiryCombobox({
  id,
  value,
  onValueChange,
}: {
  id: string;
  value: string;
  onValueChange: (value: string) => void;
}) {
  const [query, setQuery] = useState(() => expiryLabel(value));

  return (
    <Combobox
      filteredItems={expiryOptions(query, value)}
      value={value}
      onValueChange={(v) => v && onValueChange(v)}
      inputValue={query}
      onInputValueChange={(text, details) => {
        setQuery(text);
        const typed = details.reason === "input-change" && parseExpiry(text);
        if (typed) onValueChange(typed);
      }}
      itemToStringLabel={expiryLabel}
      autoHighlight
    >
      <ComboboxInput id={id} className="w-full" spellCheck={false} autoCapitalize="none" />
      <ComboboxContent>
        <ComboboxEmpty>Type a length of time, like 45d, 12h or 1h30m.</ComboboxEmpty>
        <ComboboxList>
          {(v: string) => (
            <ComboboxItem key={v} value={v}>
              {expiryLabel(v)}
            </ComboboxItem>
          )}
        </ComboboxList>
      </ComboboxContent>
    </Combobox>
  );
}

// AboutTokens stands where a new token appears, saying what to know before making one
function AboutTokens() {
  return (
    <SideNote title="Info">
      <p>The token is shown once after creation, copy and store it somewhere safe.</p>
      <p>Tokens are signed with the [api.secret] key, changing it will revoke all tokens.</p>
    </SideNote>
  );
}

// TokenSlip shows a token just made, once: who it acts as, until when, the token itself, and the lines that point the CLI at this server with it
function TokenSlip({ token }: { token: string }) {
  const ref = useRef<HTMLElement>(null);
  const claims = tokenClaims(token);
  const parts = tokenParts(token);
  const config = clientConfig(window.location.origin, token);
  const expires = claims.exp ? claims.exp * 1000 : undefined;
  const expiresIn = expires ? relativeTime(expires) : undefined;
  const copyFailed = () =>
    toast.error("Couldn't copy to the clipboard", { description: "Select the text and copy it." });

  // on a narrow screen the slip lands below the form, so it is brought into view
  useEffect(() => {
    ref.current?.scrollIntoView({ block: "nearest" });
    ref.current?.focus({ preventScroll: true });
  }, [token]);

  return (
    <section
      ref={ref}
      tabIndex={-1}
      aria-labelledby="token-heading"
      className="relative grid content-start gap-5 self-start overflow-hidden rounded-lg border bg-muted/30 py-5 pr-5 pl-7 outline-none before:absolute before:inset-y-0 before:left-0 before:w-1 before:bg-(--node-ready) motion-safe:animate-in motion-safe:duration-300 motion-safe:fade-in motion-safe:slide-in-from-bottom-2"
    >
      <div className="flex flex-wrap items-start justify-between gap-3">
        <div className="grid gap-1">
          <h2 id="token-heading" className="text-base font-semibold tracking-tight">
            Your new token
          </h2>
        </div>
        <CodeBlockCopyButton value={token} variant="default" size="sm" onCopyError={copyFailed}>
          Copy token
        </CodeBlockCopyButton>
      </div>

      <dl className="flex flex-wrap gap-x-4 gap-y-3">
        <Fact label="Acts as">
          <span className="font-mono">{claims.username}</span>
        </Fact>
        <Fact label="Role">{claims.role}</Fact>
        <Fact label="Expires">
          {expires && expiresIn ? (
            <span title={expiresIn.title}>
              {dateFormat.format(expires)}, {expiresIn.text}
            </span>
          ) : (
            "never"
          )}
        </Fact>
      </dl>

      {/* the parts of the token are told apart by tone: the readable claims in full colour, the header and signature dimmer */}
      <p
        aria-label="Token"
        className="rounded-md border bg-background px-3 py-2.5 font-mono text-xs leading-relaxed break-all select-all"
      >
        <span className="text-muted-foreground">{parts.header}</span>
        <span className="text-(--node-ready)">.</span>
        <span>{parts.claims}</span>
        <span className="text-(--node-ready)">.</span>
        <span className="text-muted-foreground">{parts.signature}</span>
      </p>
    </section>
  );
}
