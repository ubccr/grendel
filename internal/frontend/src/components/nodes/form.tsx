import { useNavigate, useRouter } from "@tanstack/react-router";

import { Host } from "@/client";
import {
  getV1ImagesOptions,
  getV1NodesTagsQueryKey,
  postV1NodesMutation,
} from "@/client/@tanstack/react-query.gen";
import {
  cellInput,
  Field,
  FieldErrors,
  invalidInput,
  SaveBar,
  Section,
} from "@/components/form-layout";
import { TagCombobox } from "@/components/tag-combobox";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
import { Input } from "@/components/ui/input";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { cn } from "@/lib/utils";
import { useForm, useStore } from "@tanstack/react-form";
import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import { ChevronRight, ExternalLink, Plus, X } from "lucide-react";
import { useEffect, type ReactNode } from "react";
import { toast } from "sonner";

type Nic = NonNullable<Host["interfaces"]>[number];

// firmwareBuilds match firmware.BuildToStringMap; an empty firmware lets the DHCP server pick the build from what the node asks for
const firmwareBuilds = [
  "ipxe.pxe",
  "ipxe-i386.efi",
  "ipxe-x86_64.efi",
  "snponly-x86_64.efi",
  "snponly-arm64.efi",
  "undionly.kpxe",
];
// archs are the ones firmware.DetectArch reports
const archs = ["x86_64", "aarch64", "i386"];

// unset stands in for an empty value, since a select item can't have an empty value
const unset = "__unset__";

// OptionSelect picks one of options or leaves the value empty, shown as unsetLabel. A saved value that isn't one of the options is still listed, so opening the form never changes it
function OptionSelect({
  id,
  value,
  onChange,
  options,
  unsetLabel,
}: {
  id: string;
  value?: string;
  onChange: (v: string) => void;
  options: string[];
  unsetLabel: string;
}) {
  const all = value && !options.includes(value) ? [value, ...options] : options;
  return (
    <Select value={value || unset} onValueChange={(v) => onChange(v === unset ? "" : v)}>
      <SelectTrigger id={id} className="w-full">
        <SelectValue />
      </SelectTrigger>
      <SelectContent>
        <SelectItem value={unset} className="text-muted-foreground">
          {unsetLabel}
        </SelectItem>
        {all.map((o) => (
          <SelectItem key={o} value={o}>
            {o}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}

// the port table's columns on a wide screen; below lg each port stacks into a labelled grid
const portColumns =
  "lg:grid-cols-[3.25rem_minmax(4.5rem,0.6fr)_10.5rem_minmax(8rem,0.9fr)_minmax(12rem,2fr)_4rem_4.5rem_2.75rem_4.5rem]";

// Cell labels a port input; the label is only shown when the ports stack on a narrow screen, the column header names it otherwise
function Cell({
  label,
  className,
  children,
}: {
  label: string;
  className?: string;
  children: ReactNode;
}) {
  return (
    <label className={cn("grid min-w-0 gap-1", className)}>
      <span className="text-xs text-muted-foreground lg:sr-only">{label}</span>
      {children}
    </label>
  );
}

// the port fields that say which port it is; a port with none of them set is empty, and the API would store it as is
const portIdentity = ["ifname", "mac", "ip", "fqdn"] as const;

export function emptyPort(nic?: Nic) {
  return !portIdentity.some((k) => String(nic?.[k] ?? "").trim());
}

// checkNode refuses a node with no name or with an empty port. The name's reason goes on its field; an empty port is a form error, and its row marks itself from its own values so it clears as soon as any of its fields is filled in
export function checkNode({ value }: { value?: Host }) {
  const name = value?.name?.trim() ? undefined : "Enter a name.";
  const ports = value?.interfaces?.some(emptyPort)
    ? "Fill in or remove the empty ports."
    : undefined;
  if (!name && !ports) return undefined;
  return { form: ports, fields: name ? { name } : {} };
}

// role names what the server uses a port for, by the same rule as Host.BootInterface and Host.InterfaceBMC
function role(nics: Nic[], i: number) {
  if (nics[i]?.bmc) return nics.findIndex((n) => n.bmc) === i ? "BMC" : undefined;
  return nics.findIndex((n) => !n.bmc) === i ? "boot" : undefined;
}

export default function NodeForm({ data, reset }: { data?: Host; reset?: boolean }) {
  const { mutateAsync, isPending } = useMutation(postV1NodesMutation());
  const images = useQuery(getV1ImagesOptions());
  const router = useRouter();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const adding = data == undefined;

  const form = useForm({
    defaultValues: data,
    // checked on save, then on every change once a save was refused, so a port isn't marked while it is still being filled in
    validators: {
      onSubmit: checkNode,
      onChange: ({ value, formApi }) =>
        formApi.state.submissionAttempts > 0 ? checkNode({ value }) : undefined,
    },
    onSubmit: async ({ value }) => {
      if (value == undefined) return;
      await mutateAsync(
        { body: { node_list: [value] } },
        {
          onSuccess: (res) => {
            toast.success(res?.title, { description: res?.detail });
            if (!adding) form.reset(value);
            // a renamed node is only found under its new name, so the page moves there rather than reloading the old one
            if (!adding && value.name && value.name !== data?.name)
              navigate({ to: "/nodes/$node/node", params: { node: value.name } });
            else router.invalidate();
            queryClient.invalidateQueries({ queryKey: getV1NodesTagsQueryKey() });
          },
          onError: (e) => toast.error(e.title, { description: e.detail }),
        },
      );
    },
  });

  // fields are only marked once a save has been tried, so a port isn't red while it is still being filled in
  const attempted = useStore(form.store, (st) => st.submissionAttempts > 0);

  useEffect(() => {
    form.reset();
  }, [reset]);

  return (
    <form
      className="grid gap-10 pt-3"
      onSubmit={(e) => {
        e.preventDefault();
        e.stopPropagation();
        form.handleSubmit();
      }}
    >
      <div className="grid gap-4">
        <div className="grid gap-x-6 gap-y-5 sm:grid-cols-2 lg:grid-cols-3">
          <form.Field name="name">
            {(field) => (
              <Field label="Name">
                {(id) => (
                  <>
                    <Input
                      id={id}
                      className={cn("font-mono", invalidInput)}
                      aria-invalid={field.state.meta.errors.length > 0}
                      value={field.state.value ?? ""}
                      onBlur={field.handleBlur}
                      onChange={(e) => field.handleChange(e.target.value)}
                    />
                    <FieldErrors errors={field.state.meta.errors} />
                  </>
                )}
              </Field>
            )}
          </form.Field>
          <form.Field name="boot_image">
            {(field) => (
              <Field label="Image">
                {(id) => (
                  <OptionSelect
                    id={id}
                    value={field.state.value}
                    onChange={field.handleChange}
                    options={images.data?.map((image) => image.name) ?? []}
                    unsetLabel="Default image"
                  />
                )}
              </Field>
            )}
          </form.Field>
          <form.Field name="provision">
            {(field) => (
              <Field label="Provision">
                {(id) => (
                  <div className="flex h-9 items-center gap-2">
                    <Switch
                      id={id}
                      className="data-[state=checked]:bg-emerald-600"
                      checked={field.state.value ?? false}
                      onBlur={field.handleBlur}
                      onCheckedChange={(e) => field.handleChange(e)}
                    />
                    <span className="text-sm">{field.state.value ? "On" : "Off"}</span>
                  </div>
                )}
              </Field>
            )}
          </form.Field>
        </div>
        {/* arch and firmware are detected when the node boots, so they start collapsed */}
        <Collapsible>
          <CollapsibleTrigger className="group inline-flex items-center gap-1 rounded-sm text-sm text-muted-foreground hover:text-foreground focus-visible:ring-2 focus-visible:ring-ring focus-visible:outline-none">
            <ChevronRight className="size-4 transition-transform group-data-[state=open]:rotate-90 motion-reduce:transition-none" />
            Advanced
          </CollapsibleTrigger>
          <CollapsibleContent className="p-4">
            <div className="grid gap-x-6 gap-y-5 sm:grid-cols-2">
              <form.Field name="arch">
                {(field) => (
                  <Field label="Arch">
                    {(id) => (
                      <OptionSelect
                        id={id}
                        value={field.state.value}
                        onChange={field.handleChange}
                        options={archs}
                        unsetLabel="Detect automatically"
                      />
                    )}
                  </Field>
                )}
              </form.Field>
              <form.Field name="firmware">
                {(field) => (
                  <Field label="Firmware">
                    {(id) => (
                      <OptionSelect
                        id={id}
                        value={field.state.value}
                        onChange={field.handleChange}
                        options={firmwareBuilds}
                        unsetLabel="Detect automatically"
                      />
                    )}
                  </Field>
                )}
              </form.Field>
            </div>
          </CollapsibleContent>
        </Collapsible>
      </div>

      <Section
        title="Tags"
        description={
          <>
            A key like <code className="font-mono">dell</code> or a key and value like{" "}
            <code className="font-mono">brand=dell</code>. Nodesets and actions can select nodes by
            tag.
          </>
        }
      >
        <form.Field name="tags">
          {(field) => (
            <TagCombobox
              label="Tags"
              className="bg-card dark:bg-card"
              value={field.state.value ?? []}
              onValueChange={(e) => field.handleChange(e)}
            />
          )}
        </form.Field>
      </Section>

      <form.Field name="interfaces" mode="array">
        {(field) => {
          const nics = field.state.value ?? [];
          return (
            <Section
              title="Interfaces"
              description="Grendel answers DHCP and network boot requests from any of these ports by its MAC."
              action={
                <Button type="button" variant="secondary" onClick={() => field.pushValue({})}>
                  <Plus />
                  Add interface
                </Button>
              }
            >
              {nics.length === 0 ? (
                <p className="rounded-md border border-dashed px-4 py-6 text-center text-sm text-muted-foreground">
                  No interfaces yet. Add the port the node boots from, with its MAC, so Grendel can
                  answer its network boot.
                </p>
              ) : (
                <div className="grid text-sm" role="table" aria-label="Interfaces">
                  <div
                    role="row"
                    className={cn(
                      "hidden gap-2 border-b px-2 pb-2 text-xs text-muted-foreground lg:grid",
                      portColumns,
                    )}
                  >
                    <span role="columnheader">
                      <span className="sr-only">Role</span>
                    </span>
                    {["Name", "MAC", "IP", "FQDN", "VLAN", "MTU", "BMC"].map((h) => (
                      <span key={h} role="columnheader" className="px-2">
                        {h}
                      </span>
                    ))}
                    <span role="columnheader">
                      <span className="sr-only">Actions</span>
                    </span>
                  </div>
                  {nics.map((nic, i) => {
                    const r = role(nics, i);
                    const missing = attempted && emptyPort(nic);
                    return (
                      <div
                        key={i}
                        role="row"
                        className={cn(
                          "grid grid-cols-2 items-end gap-2 border-b px-2 py-3 last:border-b-0 sm:grid-cols-6 lg:items-center lg:py-1.5",
                          portColumns,
                        )}
                      >
                        <div
                          className="col-span-full flex items-center justify-between lg:col-span-1"
                          role="cell"
                        >
                          {r ? (
                            <Badge
                              variant="outline"
                              className={cn(
                                "font-normal",
                                r === "boot" &&
                                  "border-emerald-500/50 text-emerald-700 dark:text-emerald-400",
                              )}
                            >
                              {r}
                            </Badge>
                          ) : (
                            <span className="text-xs text-muted-foreground lg:sr-only">
                              Port {i + 1}
                            </span>
                          )}
                        </div>
                        <form.Field name={`interfaces[${i}].ifname`}>
                          {(f) => (
                            <Cell label="Name" className="col-span-2 lg:col-span-1">
                              <Input
                                aria-invalid={missing}
                                className={cellInput}
                                value={f.state.value ?? ""}
                                onChange={(e) => f.handleChange(e.target.value)}
                              />
                            </Cell>
                          )}
                        </form.Field>
                        <form.Field name={`interfaces[${i}].mac`}>
                          {(f) => (
                            <Cell label="MAC" className="col-span-2 lg:col-span-1">
                              <Input
                                aria-invalid={missing}
                                className={cellInput}
                                value={f.state.value ?? ""}
                                onChange={(e) => f.handleChange(e.target.value)}
                              />
                            </Cell>
                          )}
                        </form.Field>
                        <form.Field name={`interfaces[${i}].ip`}>
                          {(f) => (
                            <Cell label="IP" className="col-span-2 lg:col-span-1">
                              <Input
                                aria-invalid={missing}
                                className={cellInput}
                                value={f.state.value ?? ""}
                                onChange={(e) => f.handleChange(e.target.value)}
                              />
                            </Cell>
                          )}
                        </form.Field>
                        <form.Field name={`interfaces[${i}].fqdn`}>
                          {(f) => (
                            <Cell label="FQDN" className="col-span-2 sm:col-span-4 lg:col-span-1">
                              <Input
                                aria-invalid={missing}
                                className={cellInput}
                                value={f.state.value ?? ""}
                                onChange={(e) => f.handleChange(e.target.value)}
                              />
                              {missing && (
                                <p className="text-xs text-destructive">
                                  Set a name, MAC, IP or FQDN, or remove this port.
                                </p>
                              )}
                            </Cell>
                          )}
                        </form.Field>
                        <form.Field name={`interfaces[${i}].vlan`}>
                          {(f) => (
                            <Cell label="VLAN">
                              <Input
                                className={cellInput}
                                value={f.state.value ?? ""}
                                onChange={(e) => f.handleChange(e.target.value)}
                              />
                            </Cell>
                          )}
                        </form.Field>
                        <form.Field name={`interfaces[${i}].mtu`}>
                          {(f) => (
                            <Cell label="MTU">
                              <Input
                                className={cn(cellInput, "tabular-nums")}
                                type="number"
                                value={f.state.value ?? ""}
                                onChange={(e) =>
                                  f.handleChange(
                                    e.target.value === "" ? undefined : +e.target.value,
                                  )
                                }
                              />
                            </Cell>
                          )}
                        </form.Field>
                        <form.Field name={`interfaces[${i}].bmc`}>
                          {(f) => (
                            <Cell
                              label="BMC"
                              className="sm:max-lg:col-span-3 lg:justify-items-center"
                            >
                              <Switch
                                aria-label={`Port ${i + 1} is the BMC`}
                                checked={f.state.value ?? false}
                                onBlur={f.handleBlur}
                                onCheckedChange={(e) => f.handleChange(e)}
                              />
                            </Cell>
                          )}
                        </form.Field>
                        <div className="flex justify-end gap-1 sm:max-lg:col-span-3" role="cell">
                          <Button
                            type="button"
                            size="icon"
                            variant="ghost"
                            disabled={!nic?.fqdn}
                            asChild={!!nic?.fqdn}
                          >
                            {nic?.fqdn ? (
                              <a href={"https://" + nic.fqdn} target="_blank" rel="noreferrer">
                                <ExternalLink />
                                <span className="sr-only">Open {nic.fqdn}</span>
                              </a>
                            ) : (
                              <>
                                <ExternalLink />
                                <span className="sr-only">Open FQDN</span>
                              </>
                            )}
                          </Button>
                          <Button
                            type="button"
                            size="icon"
                            variant="ghost"
                            className="hover:bg-destructive/10 hover:text-destructive"
                            onClick={() => field.removeValue(i)}
                          >
                            <X />
                            <span className="sr-only">Remove port {i + 1}</span>
                          </Button>
                        </div>
                      </div>
                    );
                  })}
                </div>
              )}
            </Section>
          );
        }}
      </form.Field>

      <form.Subscribe
        selector={(s) => [s.isDirty, s.submissionAttempts > 0 && !s.isValid] as const}
      >
        {([dirty, invalid]) =>
          (dirty || adding) && (
            <SaveBar
              noun="node"
              adding={adding}
              pending={isPending}
              invalid={invalid}
              onDiscard={() => form.reset()}
            />
          )
        }
      </form.Subscribe>
    </form>
  );
}
