import type { BmcResponseModelBmcSecureBoot } from "@/client";
import KeysPanel from "@/components/secureboot/keys-panel";
import SecureBootState, { look, type Status } from "@/components/secureboot/state";
import { secureBootKeysOptions } from "@/lib/queries";
import type { SecureBootResult } from "@/lib/secureboot";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";

// fixtures name only the fields a test is about
const status = (s: Partial<Status>) => s as Status;
const results = (r: object[]) => r as SecureBootResult[];

const render = (node: ReactNode) =>
  renderToStaticMarkup(
    <QueryClientProvider client={new QueryClient()}>{node}</QueryClientProvider>,
  );

describe("look", () => {
  it("reads the setting, a staged change and the current boot as one state", () => {
    expect(
      look(status({ enabled: true, current_boot: "Enabled", mode: "DeployedMode" })).title,
    ).toBe("Secure boot is on");
    expect(
      look(status({ enabled: false, pending_enabled: true, current_boot: "Disabled" })).title,
    ).toBe("Secure boot turns on at the next boot");
    expect(
      look(status({ enabled: true, pending_enabled: false, current_boot: "Enabled" })).title,
    ).toBe("Secure boot turns off at the next boot");
    expect(look(status({ enabled: true, mode: "SetupMode" })).title).toBe(
      "Secure boot is on, but has no platform key",
    );
    expect(look(status({ enabled: false })).title).toBe("Secure boot is off");
  });
});

describe("secure boot views", () => {
  it("renders a node's state with the control to change it", () => {
    const html = render(
      <SecureBootState
        node="cpn-01"
        status={status({ enabled: false, current_boot: "Disabled", mode: "UserMode" })}
      />,
    );
    expect(html).toContain("Secure boot is off");
    expect(html).toContain("Turn on");
  });

  it("lists the db keys with how many nodes hold each", () => {
    const cert = {
      id: "1",
      fingerprint: "ab".repeat(32),
      subject: "Grendel DB",
      issuer: "Grendel DB",
      not_after: "2036-01-01T00:00:00Z",
    };
    const client = new QueryClient();
    client.setQueryData(secureBootKeysOptions("cpn-[01-02]", "trusted").queryKey, {
      results: results([
        {
          node: "cpn-01",
          success: true,
          data: { keys: [{ id: "db", certificates: [cert], signatures: [] }] },
        },
        {
          node: "cpn-02",
          success: true,
          data: {
            keys: [
              { id: "db", certificates: [], signatures: [] },
              { id: "PK", certificates: [], signatures: [] },
            ],
          },
        },
      ]),
    } as BmcResponseModelBmcSecureBoot);
    const html = renderToStaticMarkup(
      <QueryClientProvider client={client}>
        <KeysPanel nodes="cpn-[01-02]" />
      </QueryClientProvider>,
    );
    expect(html).toContain("Grendel DB");
    expect(html).toContain("On 1 of 2 nodes");
    expect(html).toContain("Missing on cpn-02");
    expect(html).toContain("None, so nothing is enforced");
    expect(html).toContain("Reading…");
  });
});
