import {
  postV1BmcBiosDiscardMutation,
  postV1BmcBiosMutation,
} from "@/client/@tanstack/react-query.gen";
import { ConfirmDialog } from "@/components/actions/action-card";
import ReviewDialog from "@/components/bios/review-dialog";
import NodesetView from "@/components/nodeset-view";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Switch } from "@/components/ui/switch";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import {
  BiosAttribute,
  BiosValue,
  BiosValues,
  editable,
  inputKind,
  menuSection,
  parseInput,
  sameValue,
  showValue,
} from "@/lib/bios";
import { toastBmcResponse } from "@/lib/bmc";
import { cn } from "@/lib/utils";
import { useMutation } from "@tanstack/react-query";
import {
  CircleHelp,
  LoaderCircle,
  Lock,
  RotateCw,
  Search,
  TriangleAlert,
  Undo2,
} from "lucide-react";
import { ReactNode, useMemo, useState } from "react";
import { toast } from "sonner";

type Props = {
  // nodes is the nodeset a change is applied to
  nodes: string;
  values: BiosValues;
  pending: BiosValues;
  // registry is undefined when the bmc publishes none, and the editor then works from the raw names and value types
  registry?: BiosAttribute[];
  applyTimes: string[];
  notice?: ReactNode;
  onApplied: () => void;
};

type Row = { name: string; attr?: BiosAttribute; section: string };

// BiosEditor lists a node's BIOS attributes by setup menu, and stages the changes made to them on every node in nodes
export default function BiosEditor({
  nodes,
  values,
  pending,
  registry,
  applyTimes,
  notice,
  onApplied,
}: Props) {
  const [search, setSearch] = useState("");
  const [section, setSection] = useState<string>();
  const [pendingOnly, setPendingOnly] = useState(false);
  const [changedOnly, setChangedOnly] = useState(false);
  const [showHidden, setShowHidden] = useState(false);
  const [edits, setEdits] = useState<BiosValues>({});
  const [applyTime, setApplyTime] = useState(
    applyTimes.includes("OnReset") ? "OnReset" : (applyTimes[0] ?? ""),
  );

  const setBios = useMutation(postV1BmcBiosMutation());
  const discard = useMutation(postV1BmcBiosDiscardMutation());

  // the registry orders and groups the attributes. Any attribute it leaves out still gets a row, under Other
  const rows = useMemo(() => {
    const out: Row[] = [];
    const seen = new Set<string>();
    for (const attr of registry ?? []) {
      if (!attr.name || !(attr.name in values) || seen.has(attr.name)) continue;
      seen.add(attr.name);
      out.push({ name: attr.name, attr, section: menuSection(attr.menu_path) });
    }
    for (const name of Object.keys(values).sort()) {
      if (!seen.has(name)) out.push({ name, section: "Other" });
    }
    return out;
  }, [registry, values]);

  // effective is what the node runs once anything staged applies, which is what an edit is compared against
  const effective = (name: string) => (name in pending ? pending[name] : values[name]);

  const visible = rows.filter(
    (r) =>
      (showHidden || !r.attr?.hidden) &&
      (!pendingOnly || r.name in pending) &&
      (!changedOnly ||
        r.attr?.default_value === undefined ||
        !sameValue(effective(r.name), r.attr.default_value as BiosValue)),
  );

  const sections = [...new Set(visible.map((r) => r.section))];
  const counts = new Map<string, number>();
  for (const r of visible) counts.set(r.section, (counts.get(r.section) ?? 0) + 1);

  const query = search.trim().toLowerCase();
  const current = section && sections.includes(section) ? section : sections[0];
  const shown = query
    ? visible.filter((r) =>
        `${r.name} ${r.attr?.display_name ?? ""} ${r.attr?.help_text ?? ""}`
          .toLowerCase()
          .includes(query),
      )
    : visible.filter((r) => r.section === current);

  const edit = (name: string, value: BiosValue) => {
    setEdits((prev) => {
      const next = { ...prev };
      if (sameValue(value, effective(name))) delete next[name];
      else next[name] = value;
      return next;
    });
  };

  const changes = Object.entries(edits);
  const pendingCount = Object.keys(pending).length;

  const apply = () =>
    setBios.mutate(
      {
        query: { nodeset: nodes },
        body: { attributes: edits, apply_time: applyTime || undefined },
      },
      {
        onSuccess: (data) => {
          toastBmcResponse("Stage BIOS changes", data);
          if (!data.summary?.failed) setEdits({});
          onApplied();
        },
        onError: (e) => toast.error(e.title, { description: e.detail }),
      },
    );

  return (
    <div className="grid gap-4">
      {notice}
      {!registry && (
        <div className="flex items-start gap-2 rounded-md border border-dashed p-3 text-sm text-muted-foreground">
          <TriangleAlert className="mt-0.5 size-4 shrink-0" />
          This BMC publishes no attribute registry, so attributes are shown by their raw names with
          no allowed values. The BMC still rejects a value it does not accept.
        </div>
      )}

      <div className="flex flex-wrap items-center gap-x-6 gap-y-3">
        <div className="relative min-w-56 flex-1">
          <Search className="absolute top-2.5 left-2.5 size-4 text-muted-foreground" />
          <Input
            className="pl-8"
            placeholder="Search every menu"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
        <Filter
          id="bios-pending"
          label="Pending only"
          checked={pendingOnly}
          onChange={setPendingOnly}
        />
        {registry && (
          <>
            <Filter
              id="bios-changed"
              label="Changed from default"
              checked={changedOnly}
              onChange={setChangedOnly}
            />
            <Filter
              id="bios-hidden"
              label="Show hidden"
              checked={showHidden}
              onChange={setShowHidden}
            />
          </>
        )}
      </div>

      <div className="grid gap-4 md:grid-cols-[14rem_1fr]">
        <nav className={cn("grid content-start gap-1", query && "opacity-50")}>
          {sections.map((s) => (
            <button
              key={s}
              type="button"
              onClick={() => {
                setSection(s);
                setSearch("");
              }}
              className={cn(
                "flex items-center justify-between rounded-md px-3 py-1.5 text-left text-sm hover:bg-muted",
                s === current && !query && "bg-muted font-medium",
              )}
            >
              <span className="truncate">{s}</span>
              <span className="text-xs text-muted-foreground">{counts.get(s)}</span>
            </button>
          ))}
        </nav>

        <Card>
          <CardContent className="divide-y p-0">
            {shown.length === 0 && (
              <div className="p-6 text-center text-sm text-muted-foreground">
                No attributes match.
              </div>
            )}
            {shown.map((r) => (
              <AttributeRow
                key={r.name}
                row={r}
                applied={values[r.name]}
                staged={pending[r.name]}
                isStaged={r.name in pending}
                value={r.name in edits ? edits[r.name] : effective(r.name)}
                edited={r.name in edits}
                onChange={(v) => edit(r.name, v)}
                onReset={() => edit(r.name, effective(r.name))}
              />
            ))}
          </CardContent>
        </Card>
      </div>

      <div className="sticky bottom-3 z-10 flex flex-wrap items-center justify-end gap-3 rounded-lg border bg-card/90 px-4 py-3 shadow-md backdrop-blur">
        <span className="mr-auto text-sm text-muted-foreground">
          {changes.length} change(s)
          {pendingCount > 0 && ` · ${pendingCount} staged on the BMC`}
        </span>
        {pendingCount > 0 && (
          <ConfirmDialog
            trigger={
              <Button variant="secondary" disabled={discard.isPending}>
                {discard.isPending && <LoaderCircle className="animate-spin" />}
                Discard pending
              </Button>
            }
            onConfirm={() =>
              discard.mutate(
                { query: { nodeset: nodes } },
                {
                  onSuccess: (data) => {
                    toastBmcResponse("Discard BIOS changes", data);
                    onApplied();
                  },
                  onError: (e) => toast.error(e.title, { description: e.detail }),
                },
              )
            }
          >
            Every staged BIOS change will be dropped on:
            <NodesetView nodes={nodes} />
          </ConfirmDialog>
        )}
        {applyTimes.length > 0 && (
          <Select
            value={applyTime}
            onValueChange={(v) => setApplyTime(v ?? "")}
            items={applyTimes.map((t) => ({ value: t, label: applyTimeLabels[t] ?? t }))}
          >
            <SelectTrigger className="w-56" aria-label="Apply time">
              <SelectValue placeholder="Apply time" />
            </SelectTrigger>
            <SelectContent>
              {applyTimes.map((t) => (
                <SelectItem key={t} value={t}>
                  {applyTimeLabels[t] ?? t}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        )}
        <Button variant="ghost" disabled={changes.length === 0} onClick={() => setEdits({})}>
          Reset
        </Button>
        <ReviewDialog
          trigger={
            <Button disabled={changes.length === 0 || setBios.isPending}>
              {setBios.isPending && <LoaderCircle className="animate-spin" />}
              Review & apply
            </Button>
          }
          changes={changes.map(([name, to]) => ({
            name,
            attr: rows.find((r) => r.name === name)?.attr,
            from: effective(name),
            to,
          }))}
          nodes={nodes}
          applyTime={applyTime}
          applyTimeLabel={applyTimeLabels[applyTime]}
          onConfirm={apply}
        />
      </div>
    </div>
  );
}

const applyTimeLabels: Record<string, string> = {
  Immediate: "Immediately",
  OnReset: "On next reboot",
  AtMaintenanceWindowStart: "At maintenance window",
  InMaintenanceWindowOnReset: "Reboot in maintenance window",
};

function Filter({
  id,
  label,
  checked,
  onChange,
}: {
  id: string;
  label: string;
  checked: boolean;
  onChange: (v: boolean) => void;
}) {
  return (
    <div className="flex items-center gap-2">
      <Switch id={id} checked={checked} onCheckedChange={onChange} />
      <Label htmlFor={id}>{label}</Label>
    </div>
  );
}

type RowProps = {
  row: Row;
  applied: BiosValue;
  staged: BiosValue;
  isStaged: boolean;
  value: BiosValue;
  edited: boolean;
  onChange: (v: BiosValue) => void;
  onReset: () => void;
};

function AttributeRow({
  row,
  applied,
  staged,
  isStaged,
  value,
  edited,
  onChange,
  onReset,
}: RowProps) {
  const { name, attr } = row;
  const kind = inputKind(attr, applied);
  const canEdit = editable(attr);

  return (
    <div
      className={cn(
        "grid gap-3 p-3 sm:grid-cols-[1fr_18rem] sm:items-center",
        edited && "bg-muted/40",
      )}
    >
      <div className="min-w-0">
        <div className="flex flex-wrap items-center gap-2">
          <span className="font-medium">{attr?.display_name || name}</span>
          {attr?.help_text && (
            <Tooltip>
              <TooltipTrigger render={<CircleHelp className="size-3.5 text-muted-foreground" />} />
              <TooltipContent className="max-w-sm border bg-popover text-sm text-popover-foreground shadow-md">
                {attr.help_text}
              </TooltipContent>
            </Tooltip>
          )}
          {isStaged && (
            <Badge variant="outline" className="border-sky-500/50 text-sky-600 dark:text-sky-400">
              pending → {showValue(staged)}
            </Badge>
          )}
          {attr?.system_unique && <Badge variant="secondary">per system</Badge>}
        </div>
        {attr?.display_name && (
          <div className="truncate font-mono text-xs text-muted-foreground">{name}</div>
        )}
        {attr?.warning_text && edited && (
          <div className="mt-1 text-xs text-amber-600 dark:text-amber-400">{attr.warning_text}</div>
        )}
      </div>

      <div className="flex items-center gap-1">
        <div className="min-w-0 flex-1">
          <AttributeInput
            attr={attr}
            kind={kind}
            value={value}
            disabled={!canEdit}
            onChange={onChange}
          />
        </div>
        {/* the slot is kept when empty, so every input in the list lines up */}
        <span className="ml-2 flex w-5 shrink-0 justify-center">
          {!canEdit && kind !== "password" ? (
            <Tooltip>
              <TooltipTrigger
                render={<Lock className="size-4 text-muted-foreground" aria-label="Read only" />}
              />
              <TooltipContent className="border bg-popover text-sm text-popover-foreground shadow-md">
                Read only
              </TooltipContent>
            </Tooltip>
          ) : (
            attr?.reset_required && (
              <Tooltip>
                <TooltipTrigger
                  render={
                    <RotateCw
                      className="size-4 text-muted-foreground"
                      aria-label="Reboot required"
                    />
                  }
                />
                <TooltipContent className="border bg-popover text-sm text-popover-foreground shadow-md">
                  Reboot required
                </TooltipContent>
              </Tooltip>
            )
          )}
        </span>
        <Button
          variant="ghost"
          size="icon"
          className={cn(!edited && "invisible")}
          onClick={onReset}
          aria-label={`Undo ${name}`}
        >
          <Undo2 />
        </Button>
      </div>
    </div>
  );
}

function AttributeInput({
  attr,
  kind,
  value,
  disabled,
  onChange,
}: {
  attr?: BiosAttribute;
  kind: ReturnType<typeof inputKind>;
  value: BiosValue;
  disabled: boolean;
  onChange: (v: BiosValue) => void;
}) {
  switch (kind) {
    case "password":
      return <span className="text-sm text-muted-foreground">hidden</span>;
    case "boolean":
      return <Switch checked={value === true} disabled={disabled} onCheckedChange={onChange} />;
    case "enum": {
      const options = attr?.values ?? [];
      const current = String(value ?? "");
      return (
        <Select
          value={current}
          disabled={disabled}
          onValueChange={onChange}
          items={[
            { value: current, label: current },
            ...options.map((o) => ({ value: o.name ?? "", label: o.display_name || o.name })),
          ]}
        >
          <SelectTrigger className="w-full">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {!options.some((o) => o.name === current) && current && (
              <SelectItem value={current}>{current}</SelectItem>
            )}
            {options.map((o) => (
              <SelectItem key={o.name} value={o.name ?? ""}>
                {o.display_name || o.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      );
    }
    default:
      return (
        <Input
          type={kind === "integer" ? "number" : "text"}
          className="font-mono"
          value={value == null ? "" : String(value)}
          disabled={disabled}
          min={attr?.lower_bound ?? undefined}
          max={attr?.upper_bound ?? undefined}
          maxLength={attr?.max_length ?? undefined}
          onChange={(e) => onChange(parseInput(kind, e.target.value))}
        />
      );
  }
}
