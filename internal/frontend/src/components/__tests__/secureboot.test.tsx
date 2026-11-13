import KeysPanel from "@/components/secureboot/keys-panel";
import SecureBootState, { look } from "@/components/secureboot/state";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import type { ReactNode } from "react";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";

const render = (node: ReactNode) =>
  renderToStaticMarkup(
    <QueryClientProvider client={new QueryClient()}>{node}</QueryClientProvider>,
  );

describe("look", () => {
  it("reads the setting, a staged change and the current boot as one state", () => {
    expect(look({ enabled: true, current_boot: "Enabled", mode: "DeployedMode" }).title).toBe(
      "Secure boot is on",
    );
    expect(look({ enabled: false, pending_enabled: true, current_boot: "Disabled" }).title).toBe(
      "Secure boot turns on at the next boot",
    );
    expect(look({ enabled: true, pending_enabled: false, current_boot: "Enabled" }).title).toBe(
      "Secure boot turns off at the next boot",
    );
    expect(look({ enabled: true, mode: "SetupMode" }).title).toBe(
      "Secure boot is on, but has no platform key",
    );
    expect(look({ enabled: false }).title).toBe("Secure boot is off");
  });
});

describe("secure boot views", () => {
  it("renders a node's state with the control to change it", () => {
    const html = render(
      <SecureBootState
        node="cpn-01"
        status={{ enabled: false, current_boot: "Disabled", mode: "UserMode" }}
        onChanged={() => {}}
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
    const html = render(
      <KeysPanel
        nodes="cpn-[01-02]"
        onChanged={() => {}}
        results={[
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
        ]}
      />,
    );
    expect(html).toContain("Grendel DB");
    expect(html).toContain("On 1 of 2 nodes");
    expect(html).toContain("Missing on cpn-02");
    expect(html).toContain("None, so nothing is enforced");
  });
});
