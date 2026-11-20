import type { BootRow, BootSettings } from "@/lib/boot-nics";
import type { Template } from "@/lib/name-template";
import { Legend, Plain, SettingsPanel, TemplateField, type Field } from "./naming";

const fields: Field[] = [
  { token: "{name}", help: "The node's name" },
  { token: "{name/cpn/ib}", help: "The name with its first cpn swapped for ib" },
  { token: "{sw}", help: "The switch that has seen the interface" },
  { token: "{sw.1}", help: "The first number in that switch's name" },
  { token: "{port}", help: "That switch port's number on the faceplate" },
  { token: "{port:02}", help: "Any number padded, here to 2 digits" },
];

const example = { name: "cpn-a01-07", sw: "swe-a01", port: 7 };

// BootSettingsPanel holds the settings every interface is stored with: its FQDN template and its interface
export function BootSettingsPanel({
  settings,
  fqdn,
  rows,
  onChange,
}: {
  settings: BootSettings;
  fqdn: Template;
  rows: BootRow[];
  onChange: (s: BootSettings) => void;
}) {
  const set = (k: keyof BootSettings, v: string) => onChange({ ...settings, [k]: v });
  const first = rows.find((r) => r.include && r.nic);
  const sample = rows.find((r) => r.include && r.fqdn)?.fqdn;
  const summary = [
    { label: "FQDN", value: settings.fqdn },
    { label: "Interface", value: settings.ifname },
    { label: "VLAN", value: settings.vlan },
    { label: "MTU", value: settings.mtu },
  ].filter((f) => f.value);

  return (
    <SettingsPanel
      title="Interface"
      storageKey="grendel.boot-nics.settings-collapsed"
      summary={summary}
      empty="No FQDN or interface settings"
    >
      <div className="grid gap-5 p-4 lg:grid-cols-[minmax(0,1fr)_auto]">
        <TemplateField
          label="FQDN"
          value={settings.fqdn}
          template={fqdn}
          placeholder="{name}.cluster"
          sample={sample}
          optional
          onChange={(v) => set("fqdn", v)}
        />
        <div className="flex flex-wrap items-start gap-4">
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
            value={settings.vlan}
            className="w-28"
            onChange={(v) => set("vlan", v.trim())}
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
      </div>
      <Legend
        fields={fields}
        vars={first ? { name: first.node, sw: first.seen?.sw, port: first.seen?.port } : example}
        subject={first ? `${first.node} ${first.nic?.id}` : example.name}
        note={
          first
            ? first.seen
              ? undefined
              : ". No scanned switch has seen this port, so it has no {sw} or {port}"
            : ", an example until a port is picked"
        }
      />
    </SettingsPanel>
  );
}
