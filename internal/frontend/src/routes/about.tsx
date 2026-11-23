import type { Version } from "@/client";
import { getV1GrendelVersionOptions } from "@/client/@tanstack/react-query.gen";
import { ErrorPanel } from "@/components/error-panel";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Skeleton } from "@/components/ui/skeleton";
import AuthRedirect from "@/lib/auth";
import { buildSummary, releaseUrl, REPO } from "@/lib/build";
import { copyText } from "@/lib/clipboard";
import { relativeTime } from "@/lib/utils";
import "@fontsource/unifrakturcook/latin-700.css";
import { useQuery } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { Check, Copy } from "lucide-react";
import { type ReactNode, useState } from "react";

export const Route = createFileRoute("/about")({
  staticData: { nodeField: "full", crumb: "About" },
  component: RouteComponent,
  beforeLoad: AuthRedirect,
});

const links = [
  { title: "Documentation", href: "https://grendel.readthedocs.io/en/latest/" },
  { title: "Source code", href: REPO },
  { title: "Report a problem", href: `${REPO}/issues` },
];

function External({ href, children }: { href: string; children: ReactNode }) {
  return (
    <a
      className="font-medium text-foreground underline decoration-border underline-offset-4 hover:decoration-foreground"
      href={href}
      target="_blank"
      rel="noreferrer"
    >
      {children}
    </a>
  );
}

function Fact({ label, children }: { label: string; children: ReactNode }) {
  return (
    <div className="flex items-baseline justify-between gap-6 border-t py-2.5">
      <dt className="shrink-0 text-muted-foreground">{label}</dt>
      <dd className="min-w-0 text-right wrap-anywhere tabular-nums">{children}</dd>
    </div>
  );
}

function CopyBuild({ build }: { build: Version }) {
  const [copied, setCopied] = useState(false);
  const copy = async () => {
    if (!(await copyText(buildSummary(build)))) return;
    setCopied(true);
    setTimeout(() => setCopied(false), 2000);
  };
  return (
    <Button type="button" variant="outline" size="sm" onClick={copy}>
      {copied ? <Check /> : <Copy />}
      {copied ? "Copied" : "Copy build details"}
    </Button>
  );
}

function RouteComponent() {
  const { data: build, error, isPending } = useQuery(getV1GrendelVersionOptions());
  const release = releaseUrl(build?.Version);
  const committed = build?.commit_time && relativeTime(build.commit_time);

  return (
    <div className="flex flex-1 items-center justify-center px-2 pb-[8vh]">
      <Card className="w-full max-w-xl">
        <CardContent className="flex flex-col gap-6 px-8 py-4">
          <header className="flex flex-col gap-3">
            <div className="flex flex-wrap items-baseline gap-x-5 gap-y-1">
              <h1 className="font-[UnifrakturCook] text-6xl leading-none sm:text-7xl">Grendel</h1>
              {isPending ? (
                <Skeleton className="h-8 w-24 self-center" />
              ) : (
                <p className="text-3xl font-light tracking-tight text-muted-foreground tabular-nums">
                  {build?.Version ?? "version unknown"}
                </p>
              )}
            </div>
            <p className="max-w-[60ch] text-sm text-muted-foreground">
              Grendel provisions bare-metal servers for high-performance computing clusters. It is
              developed by the University at Buffalo Center for Computational Research.
            </p>
          </header>

          {error ? (
            <ErrorPanel title="Could not read the server's version" message={error.detail} />
          ) : (
            <dl className="border-b text-sm">
              <Fact label="Version">
                {isPending ? (
                  <Skeleton className="h-4 w-20" />
                ) : release ? (
                  <External href={release}>{build.Version}</External>
                ) : (
                  (build.Version ?? "Unknown")
                )}
              </Fact>
              {build?.commit && (
                <Fact label="Commit">
                  <External href={`${REPO}/commit/${build.commit}`}>
                    {build.commit.slice(0, 12)}
                  </External>
                  {build.modified && (
                    <span className="text-muted-foreground"> with uncommitted changes</span>
                  )}
                </Fact>
              )}
              {committed && (
                <Fact label="Committed">
                  <time dateTime={build.commit_time} title={committed.title}>
                    {committed.text}
                  </time>
                </Fact>
              )}
              {build?.go_version && (
                <Fact label="Built with">Go {build.go_version.replace(/^go/, "")}</Fact>
              )}
              {!!build?.schema && <Fact label="Database schema">{build.schema}</Fact>}
              <Fact label="License">
                <External href={`${REPO}/blob/main/LICENSES/GPL-3.0-or-later.txt`}>
                  GPL-3.0-or-later
                </External>
              </Fact>
            </dl>
          )}

          <footer className="flex flex-wrap items-center justify-between gap-x-6 gap-y-3 text-sm">
            <nav aria-label="Grendel elsewhere" className="flex flex-wrap gap-x-5 gap-y-1">
              {links.map((link) => (
                <External key={link.href} href={link.href}>
                  {link.title}
                </External>
              ))}
            </nav>
            {build && <CopyBuild build={build} />}
          </footer>
        </CardContent>
      </Card>
    </div>
  );
}
