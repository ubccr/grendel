import type { BmcResponseModelBmcSecureBoot } from "@/client";
import SecureBootToggle from "@/components/secureboot/toggle";
import { cn } from "@/lib/utils";
import { ShieldAlert, ShieldCheck, ShieldEllipsis, ShieldOff, type LucideIcon } from "lucide-react";

export type Status = NonNullable<
  NonNullable<BmcResponseModelBmcSecureBoot["results"]>[number]["data"]
>;

// modes say in plain words what each UEFI secure boot mode means for what the node will boot
const modes: Record<string, string> = {
  SetupMode: "No platform key is enrolled, so the firmware checks nothing and accepts any new key.",
  UserMode: "A platform key is enrolled. Boot loaders and kernels are checked against db and dbx.",
  AuditMode: "Boot images are checked and the results logged, but nothing is blocked.",
  DeployedMode:
    "The keys are locked in. Boot loaders and kernels are checked against db and dbx, and the mode cannot change without the platform key.",
};

// Tone is the colour of a state, kept as separate classes so a caller can take the text colour alone
export type Tone = { text: string; border: string };

const tones = {
  ok: { text: "text-green-600", border: "border-green-400/60" },
  warn: { text: "text-yellow-600", border: "border-yellow-400/60" },
  off: { text: "text-muted-foreground", border: "border-border" },
} satisfies Record<string, Tone>;

type Look = { icon: LucideIcon; title: string; tone: Tone; detail?: string };

// next is what the node uses from its next boot, a staged change if there is one
export function nextBoot(s: Status) {
  return s.pending_enabled ?? s.enabled ?? false;
}

// look reads the three things the BMC reports, the setting, a staged change and the current boot, as the one sentence that matters: is the node protected now, and will it be after a reboot
export function look(s: Status): Look {
  const next = nextBoot(s);
  const now = s.current_boot === "Enabled";
  if (s.mode === "SetupMode" && next)
    return {
      icon: ShieldAlert,
      title: "Secure boot is on, but has no platform key",
      tone: tones.warn,
      detail: "Nothing is enforced until a platform key is enrolled.",
    };
  if (next && now)
    return {
      icon: ShieldCheck,
      title: "Secure boot is on",
      tone: tones.ok,
    };
  if (next)
    return {
      icon: ShieldEllipsis,
      title: "Secure boot turns on at the next boot",
      tone: tones.warn,
      detail: "The node is running now without it.",
    };
  if (now)
    return {
      icon: ShieldEllipsis,
      title: "Secure boot turns off at the next boot",
      tone: tones.warn,
      detail: "The node is protected until it reboots.",
    };
  return {
    icon: ShieldOff,
    title: "Secure boot is off",
    tone: tones.off,
  };
}

type Props = {
  node: string;
  status: Status;
  onChanged: () => void;
};

// SecureBootState is the headline of a node's secure boot tab, with the control that flips it
export default function SecureBootState({ node, status, onChanged }: Props) {
  const l = look(status);
  const Icon = l.icon;
  const next = nextBoot(status);
  const mode = status.mode ? modes[status.mode] : undefined;

  return (
    <div
      className={cn(
        "flex flex-wrap items-start justify-between gap-4 rounded-lg border-l-4 bg-muted/30 px-5 py-4",
        l.tone.text,
        l.tone.border,
      )}
    >
      <div className="flex min-w-0 items-start gap-3">
        <Icon className="mt-0.5 size-6 shrink-0" />
        <div className="min-w-0 space-y-1">
          <h2 className="text-lg font-semibold tracking-tight text-foreground">{l.title}</h2>
          <p className="max-w-prose text-sm text-muted-foreground">
            {[l.detail, mode].filter(Boolean).join(" ")}
          </p>
          {status.mode && (
            <p className="text-xs text-muted-foreground">
              Mode {status.mode.replace(/Mode$/, "").toLowerCase()}, this boot ran with secure boot{" "}
              {status.current_boot === "Enabled" ? "on" : "off"}
            </p>
          )}
        </div>
      </div>
      <div className="flex flex-wrap gap-2">
        <SecureBootToggle
          nodes={node}
          enable={!next}
          variant={next ? "secondary" : "default"}
          onDone={onChanged}
        />
      </div>
    </div>
  );
}
