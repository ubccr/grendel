import { Manifest } from "@/components/boot/manifest";
import type { NodeCheck } from "@/lib/boot-check";
import { renderToStaticMarkup } from "react-dom/server";
import { describe, expect, it } from "vitest";

const check: NodeCheck = {
  name: "cpn-01",
  image: "flatcar",
  files: [
    { name: "kernel", path: "/images/vmlinuz", size: 12 * 1024 * 1024 },
    {
      name: "initrd-0",
      path: "/images/initrd.img",
      error: "open /images/initrd.img: no such file or directory",
    },
  ],
  templates: [
    { kind: "butane", template: "site.tmpl" },
    { kind: "ipxe", template: "ipxe.tmpl" },
  ],
  on_request: [{ kind: "kickstart", template: "kickstart.tmpl" }],
};

const render = (open: string) =>
  renderToStaticMarkup(<Manifest check={check} open={open} onOpen={() => {}} />);

const openStep = (html: string) =>
  /aria-current="true"[^>]*>((?:(?!<\/button>).)*)/.exec(html)?.[1] ?? "";

describe("Manifest", () => {
  it("lists the iPXE script, then the files it loads, then the configs", () => {
    const html = render("ipxe");
    const at = (text: string) => html.indexOf(text);
    expect(at("iPXE script")).toBeGreaterThan(-1);
    expect(at("iPXE script")).toBeLessThan(at("Kernel"));
    expect(at("Kernel")).toBeLessThan(at("Initrd 0"));
    expect(at("Initrd 0")).toBeLessThan(at("Butane config"));
  });

  it("shows a readable file's size, and keeps the reason a file failed for when it is opened", () => {
    const html = render("ipxe");
    expect(html).toContain("12.0 MiB");
    expect(html).not.toContain("no such file");
  });

  it("opens a file as a step of its own", () => {
    expect(openStep(render("file:initrd-0"))).toContain("Initrd 0");
  });

  it("keeps the butane step open on either side of its translation", () => {
    expect(openStep(render("ignition"))).toContain("Butane config");
    expect(openStep(render("butane"))).toContain("Butane config");
  });

  it("lists what the server says is served on request as steps that were not checked", () => {
    const html = render("ipxe");
    expect(html).toContain("kickstart.tmpl");
    expect(html).toContain("Not checked");
  });
});
