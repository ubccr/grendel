import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Info } from "lucide-react";
import type { ReactNode } from "react";

// data is what internal/provision/handler.go verifyClaims hands RenderString for an image's kernel, initrd and cmdline, as docs/pages/bootImages/images.md describes it
const data: [string, ReactNode][] = [
  [
    "$.host",
    <>
      The node booting: <Code>.Name</Code>, <Code>.UID</Code>, <Code>.Arch</Code>,{" "}
      <Code>.Tags</Code>, <Code>.BootImage</Code>, <Code>.Interfaces</Code>, <Code>.Bonds</Code>
    </>,
  ],
  [
    "$.nic",
    <>
      The interface it is booting from: <Code>.MAC</Code>, <Code>.Name</Code>, <Code>.IP</Code>{" "}
      (with its prefix length), <Code>.AddrString</Code> (without), <Code>.FQDN</Code>,{" "}
      <Code>.VLAN</Code>, <Code>.MTU</Code>
    </>,
  ],
  [
    "$.bootimage",
    <>
      This image as saved, before rendering: <Code>.Name</Code>, <Code>.Version</Code>,{" "}
      <Code>.KernelPath</Code>, <Code>.InitrdPaths</Code>, <Code>.CommandLine</Code>
    </>,
  ],
  [
    "$.endpoints",
    <>
      This node&apos;s provision URLs, such as <Code>.RepoURL</Code>, <Code>.KickstartURL</Code>,{" "}
      <Code>.CloudInitURL</Code> and <Code>.KernelURL</Code>
    </>,
  ],
  [
    "$.extra_vars",
    <>
      Values set under <Code>provision.extra_vars</Code> in the Grendel config
    </>,
  ],
  ["$.headers", "The boot request's HTTP headers"],
  ["$.token", "The node's boot token"],
];

const examples: [string, string][] = [
  ["/var/lib/grendel/images/{{ $.host.Name }}/vmlinuz", "A kernel per node"],
  [
    "/var/lib/grendel/images/rocky-{{ $.bootimage.Version }}/initrd.img",
    "An initrd that follows the image's version, so a new release is a change to one field",
  ],
  ["inst.ks={{ $.endpoints.KickstartURL }}", "Point the installer at this node's kickstart"],
  [
    '{{ if eq $.host.Arch "aarch64" }}console=ttyAMA0{{ else }}console=ttyS0{{ end }}',
    "A console that depends on the node's architecture",
  ],
];

function Code({ children }: { children: ReactNode }) {
  return <code className="rounded bg-muted px-1 py-0.5 font-mono text-[0.85em]">{children}</code>;
}

// TemplateHelp opens a summary of the Go template syntax the kernel, initrd and command line fields of an image are rendered with
export function TemplateHelp() {
  return (
    <Dialog>
      <DialogTrigger asChild>
        <Button type="button" variant="secondary" size="icon">
          <Info />
          <span className="sr-only">Template syntax</span>
        </Button>
      </DialogTrigger>
      <DialogContent className="max-h-[85dvh] overflow-y-auto sm:max-w-2xl">
        <DialogHeader>
          <DialogTitle>Templates in boot files</DialogTitle>
          <DialogDescription>
            The kernel, initrd and command line fields are Go templates, filled in for each node
            when it boots.
          </DialogDescription>
        </DialogHeader>

        <div className="grid gap-6 text-sm">
          <section className="grid gap-2">
            <h3 className="font-medium">Examples</h3>
            <dl className="grid gap-5">
              {examples.map(([code, what]) => (
                <div key={code} className="grid gap-1">
                  <dt>
                    <code className="block rounded-md bg-muted px-3 py-2 font-mono text-xs break-all">
                      {code}
                    </code>
                  </dt>
                  <dd className="text-muted-foreground">{what}</dd>
                </div>
              ))}
            </dl>
          </section>

          <section className="grid gap-2">
            <h3 className="font-medium">Values</h3>
            <dl className="grid gap-x-4 gap-y-2 sm:grid-cols-[auto_1fr]">
              {data.map(([key, what]) => (
                <div key={key} className="contents">
                  <dt className="font-mono text-xs leading-5">{key}</dt>
                  <dd className="text-muted-foreground max-sm:mb-1">{what}</dd>
                </div>
              ))}
            </dl>
          </section>

          <section className="grid gap-2">
            <h3 className="font-medium">Functions</h3>
            <p className="text-muted-foreground">
              Go&apos;s built in actions and functions work, like <Code>if</Code>, <Code>eq</Code>,{" "}
              <Code>printf</Code> and <Code>index</Code>, along with Grendel&apos;s own:{" "}
              <Code>Split</Code>, <Code>Join</Code>, <Code>Contains</Code>, <Code>Add</Code> and{" "}
              <Code>Mul</Code>. The same values and functions are available in provision templates.
            </p>
          </section>
        </div>
      </DialogContent>
    </Dialog>
  );
}
