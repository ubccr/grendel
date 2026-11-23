import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import type { Compiled, Row, Settings } from "@/lib/import-rows";
import {
  parseTemplate,
  renderTemplate,
  type Template,
  type TemplateVars,
} from "@/lib/name-template";
import { cn } from "@/lib/utils";
import { TagCombobox } from "@/components/tag-combobox";
import { Collapsible, CollapsibleContent, CollapsibleTrigger } from "@/components/ui/collapsible";
import { useStoredValue } from "@/hooks/use-stored-value";
import { ChevronRight } from "lucide-react";
import { useId, type ReactNode } from "react";

export type Field = { token: string; help: string };

const fields: Field[] = [
  { token: "{port}", help: "The port's number on the faceplate" },
  { token: "{port:02}", help: "Any number padded, here to 2 digits" },
  { token: "{port/2}", help: "Arithmetic with + - * or /. Division rounds up" },
  { token: "{n}", help: "Counts the ports being imported, from Count {n} from" },
  { token: "{sw}", help: "The switch's name" },
  { token: "{sw.1}", help: "The first number in the switch's name. {sw.2} is the second" },
  { token: "{name}", help: "The node's name, typed or templated. FQDN and tags only" },
  { token: "{name/cpn/bmc}", help: "The name with its first cpn swapped for bmc. {sw/…/…} too" },
];

const example = { port: 7, sw: "swe-a01", name: "cpn-a01-07", ifname: "Ethernet7" };

export function Naming({
  settings,
  compiled,
  rows,
  onChange,
}: {
  settings: Settings;
  compiled: Compiled;
  rows: Row[];
  onChange: (s: Settings) => void;
}) {
  const id = useId();
  const set = <K extends keyof Settings>(k: K, v: Settings[K]) => onChange({ ...settings, [k]: v });
  const named = rows.filter((r) => r.include && r.name && !r.nameOverridden);
  const sample = named.length
    ? named[0].name + (named.length > 1 ? ` to ${named[named.length - 1].name}` : "")
    : undefined;
  const fqdnSample = rows.find((r) => r.include && r.fqdn)?.fqdn;
  const badTag = compiled.tags.find((t) => t.error);

  const summary = [
    { label: "Name", value: settings.name },
    { label: "FQDN", value: settings.fqdn },
    { label: "VLAN", value: settings.portVlan },
    { label: "MTU", value: settings.mtu },
  ].filter((f) => f.value);

  return (
    <SettingsPanel
      title="Templates and interface"
      storageKey="grendel.import.settings-collapsed"
      summary={summary}
      empty="No name template yet"
    >
      <div className="grid max-lg:divide-y lg:grid-cols-2 lg:divide-x">
        <Group title="Node" className="grid content-start gap-5">
          <TemplateField
            label="Name"
            value={settings.name}
            template={compiled.name}
            placeholder="cpn-a{sw.1}-{port:02}"
            sample={sample}
            onChange={(v) => set("name", v)}
            autoFocus
          />
          <div className="flex flex-wrap items-start gap-4">
            <div className="grid min-w-48 flex-1 content-start gap-1.5">
              <Label htmlFor={`${id}-tags`} className="text-xs text-muted-foreground">
                Tags
                <span className="font-normal">each can use the fields below</span>
              </Label>
              <TagCombobox
                id={`${id}-tags`}
                label="Tags"
                placeholder="rack=a{sw.1}"
                className="bg-background font-mono"
                value={settings.tags}
                onValueChange={(v) => set("tags", v)}
              />
              {badTag && (
                <p className="text-xs text-destructive">
                  <span className="font-mono">{badTag.source}</span> {badTag.error!.message}
                </p>
              )}
            </div>
            <Plain
              label="Count {n} from"
              value={String(settings.start)}
              inputMode="numeric"
              className="w-28"
              onChange={(v) => set("start", Math.max(0, Number(v.replace(/\D/g, "")) || 0))}
            />
          </div>
        </Group>

        <Group title="BMC interface" className="grid content-start gap-5">
          <TemplateField
            label="FQDN"
            value={settings.fqdn}
            template={compiled.fqdn}
            placeholder="{name/cpn/bmc}.cluster"
            sample={fqdnSample}
            optional
            onChange={(v) => set("fqdn", v)}
          />
          <div className="flex flex-wrap gap-4">
            <Plain
              label="Interface name"
              placeholder="None"
              value={settings.ifname}
              className="w-40"
              onChange={(v) => set("ifname", v.trim())}
            />
            <Plain
              label="VLAN"
              placeholder="None"
              value={settings.portVlan}
              className="w-28"
              onChange={(v) => set("portVlan", v.trim())}
            />
            <Plain
              label="MTU"
              placeholder="Default"
              value={settings.mtu}
              inputMode="numeric"
              className="w-28"
              onChange={(v) => set("mtu", v.replace(/\D/g, ""))}
            />
          </div>
        </Group>
      </div>
      <NameLegend rows={rows} start={settings.start} />
    </SettingsPanel>
  );
}

export function SettingsPanel({
  title,
  storageKey,
  summary,
  empty,
  children,
}: {
  title: string;
  // storageKey remembers the panel collapsed in this browser, so it stays out of the way on the next import
  storageKey: string;
  summary: { label: string; value: string }[];
  empty: string;
  children: ReactNode;
}) {
  const [collapsed, setCollapsed] = useStoredValue(storageKey);
  const open = !collapsed;
  return (
    <Collapsible
      open={open}
      onOpenChange={(o) => setCollapsed(o ? undefined : "1")}
      className="grid overflow-hidden rounded-lg border bg-muted/30"
    >
      <CollapsibleTrigger className="group flex w-full items-center gap-3 px-4 py-3 text-left outline-none hover:bg-muted/40 focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-inset data-panel-open:border-b">
        <ChevronRight className="size-4 shrink-0 text-muted-foreground transition-transform group-data-panel-open:rotate-90 motion-reduce:transition-none" />
        <span className="text-sm font-semibold tracking-tight">{title}</span>
        {!open && (
          <span className="flex min-w-0 flex-wrap gap-x-4 gap-y-1 text-xs text-muted-foreground">
            {summary.length === 0
              ? empty
              : summary.map((f) => (
                  <span key={f.label} className="truncate">
                    {f.label} <span className="font-mono text-foreground">{f.value}</span>
                  </span>
                ))}
          </span>
        )}
      </CollapsibleTrigger>
      <CollapsibleContent>{children}</CollapsibleContent>
    </Collapsible>
  );
}

export function Group({
  title,
  className,
  children,
}: {
  title: string;
  className?: string;
  children: ReactNode;
}) {
  const id = useId();
  return (
    <section role="group" aria-labelledby={id} className="grid content-start gap-3 p-4">
      <h3 id={id} className="text-sm font-semibold tracking-tight">
        {title}
      </h3>
      <div className={className}>{children}</div>
    </section>
  );
}

export function TemplateField({
  label,
  value,
  template,
  placeholder,
  sample,
  optional,
  autoFocus,
  onChange,
}: {
  label: string;
  value: string;
  template: Template;
  placeholder: string;
  sample?: string;
  optional?: boolean;
  autoFocus?: boolean;
  onChange: (v: string) => void;
}) {
  const id = useId();
  const error = template.error;

  return (
    <div className="grid content-start gap-1.5">
      <Label htmlFor={id} className="text-xs text-muted-foreground">
        {label}
        {optional && <span className="font-normal">optional</span>}
      </Label>
      <Input
        id={id}
        value={value}
        placeholder={placeholder}
        spellCheck={false}
        autoComplete="off"
        autoFocus={autoFocus}
        aria-invalid={!!error}
        aria-describedby={`${id}-note`}
        className="h-10 font-mono text-[15px]"
        onChange={(e) => onChange(e.target.value)}
      />
      <div id={`${id}-note`} className="min-h-5 text-xs">
        {error ? (
          <p className="text-destructive">
            <span className="font-mono">
              {value.slice(0, error.at)}
              <span className="rounded-sm bg-destructive/15 px-0.5 underline decoration-wavy underline-offset-2">
                {value.slice(error.at, error.at + 1) || " "}
              </span>
            </span>{" "}
            {error.message}
          </p>
        ) : sample ? (
          <p className="truncate text-muted-foreground">
            Names <span className="font-mono text-foreground">{sample}</span>
          </p>
        ) : null}
      </div>
    </div>
  );
}

export function Plain({
  label,
  value,
  placeholder,
  error,
  inputMode,
  className,
  onChange,
}: {
  label: string;
  value: string;
  placeholder?: string;
  error?: string;
  inputMode?: "numeric";
  className?: string;
  onChange: (v: string) => void;
}) {
  const id = useId();
  return (
    <div className={cn("grid content-start gap-1.5", className)}>
      <Label htmlFor={id} className="text-xs text-muted-foreground">
        {label}
      </Label>
      <Input
        id={id}
        value={value}
        placeholder={placeholder}
        inputMode={inputMode}
        spellCheck={false}
        aria-invalid={!!error}
        className={cn("font-mono", error && "border-destructive")}
        onChange={(e) => onChange(e.target.value)}
      />
      {error && <p className="text-xs text-destructive">{error}</p>}
    </div>
  );
}

// NameLegend shows the fields with their values for the first port being imported, or for a made up port before any is
function NameLegend({ rows, start }: { rows: Row[]; start: number }) {
  const included = rows.filter((r) => r.include);
  const first = included.find((r) => r.number !== undefined);
  const vars = first
    ? {
        port: first.number,
        sw: first.sw,
        n: start + included.indexOf(first),
        name: first.name || undefined,
      }
    : { ...example, n: start };
  return (
    <Legend
      fields={fields}
      vars={vars}
      subject={first ? `${first.sw} ${first.ifname}` : `${example.sw} ${example.ifname}`}
      note={first ? undefined : ", an example until a port is imported"}
    />
  );
}

export function Legend({
  fields,
  vars,
  subject,
  note,
}: {
  fields: Field[];
  vars: TemplateVars;
  subject: string;
  note?: string;
}) {
  return (
    <section
      aria-label="Template fields"
      className="grid gap-3 border-t bg-background/60 px-4 py-3"
    >
      <p className="text-xs text-muted-foreground">
        <span className="font-semibold text-foreground">Fields</span> for{" "}
        <span className="font-mono text-foreground">{subject}</span>
        {note}
      </p>
      <dl className="grid gap-x-6 gap-y-3 sm:grid-cols-2 xl:grid-cols-4">
        {fields.map((f) => {
          const value = renderTemplate(parseTemplate(f.token), vars).value;
          return (
            <div key={f.token} className="grid min-w-0 content-start gap-0.5">
              <dt className="flex min-w-0 items-baseline gap-1.5 font-mono text-xs">
                <code className="shrink-0 text-sky-700 dark:text-sky-300">{f.token}</code>
                <span aria-hidden className="text-muted-foreground">
                  =
                </span>
                <span className="truncate text-foreground" title={value}>
                  {value || <span className="text-muted-foreground">nothing here</span>}
                </span>
              </dt>
              <dd className="text-xs text-muted-foreground">{f.help}</dd>
            </div>
          );
        })}
      </dl>
    </section>
  );
}
