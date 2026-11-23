import {
  fileLabel,
  formatBytes,
  groupProblems,
  languageOf,
  problems,
  readable,
  stepFile,
  fileStep,
  translationFailure,
  viewKind,
  type NodeCheck,
} from "@/lib/boot-check";
import { describe, expect, it } from "vitest";

const failing = (name: string): NodeCheck => ({
  name,
  ok: false,
  files: [
    { name: "kernel", path: "/images/vmlinuz", size: 10 },
    { name: "initrd-0", path: "/images/initrd.img", error: "no such file or directory" },
  ],
  templates: [{ kind: "kickstart", template: "rocky.tmpl", error: "line 4: bad" }],
});

describe("problems", () => {
  it("names the file or template each problem is about", () => {
    expect(problems(failing("cpn-01"))).toEqual([
      "Initrd 0: no such file or directory",
      "Kickstart (rocky.tmpl): line 4: bad",
    ]);
  });

  it("is the node's own error when nothing could be checked", () => {
    expect(problems({ name: "cpn-01", error: "no boot image" })).toEqual(["no boot image"]);
  });

  it("groups nodes under a shared problem, the most widespread first", () => {
    const groups = groupProblems([
      failing("cpn-01"),
      failing("cpn-02"),
      { name: "cpn-03", error: "no boot image" },
      { name: "cpn-04", ok: true },
    ]);
    expect(groups.map((g) => g.nodes)).toEqual([
      ["cpn-01", "cpn-02"],
      ["cpn-01", "cpn-02"],
      ["cpn-03"],
    ]);
  });

  it("counts a node checked as two arches once", () => {
    const groups = groupProblems([failing("cpn-01"), failing("cpn-01")]);
    expect(groups[0].nodes).toEqual(["cpn-01"]);
  });
});

describe("labels", () => {
  it("names files as the boot does", () => {
    expect(fileLabel("kernel")).toBe("Kernel");
    expect(fileLabel("initrd-1")).toBe("Initrd 1");
    expect(fileLabel("kernel.sig")).toBe("Kernel signature");
  });

  it("opens a butane template as the Ignition config a node gets", () => {
    expect(viewKind("butane")).toBe("ignition");
    expect(viewKind("ipxe")).toBe("ipxe");
  });

  it("formats sizes", () => {
    expect(formatBytes(512)).toBe("512 B");
    expect(formatBytes(12 * 1024 * 1024)).toBe("12.0 MiB");
    expect(formatBytes(300 * 1024 * 1024)).toBe("300 MiB");
  });
});

describe("rendered output", () => {
  it("indents an Ignition config and nothing else", () => {
    expect(readable("ignition", '{"a":1}')).toBe('{\n  "a": 1\n}');
    expect(readable("ipxe", '{"a":1}')).toBe('{"a":1}');
    expect(readable("ignition", "not json")).toBe("not json");
  });
});

describe("languageOf", () => {
  it("knows what each kind always is", () => {
    expect(languageOf("ignition", "site.tmpl")).toBe("json");
    expect(languageOf("user_data", "ubuntu-autoinstall.tmpl")).toBe("yaml");
    expect(languageOf("butane", "butane.tmpl")).toBe("yaml");
    expect(languageOf("kickstart", "rocky.tmpl")).toBe("shellscript");
  });

  it("reads the extension under .tmpl for anything else", () => {
    expect(languageOf("bmc", "idrac-config.json.tmpl")).toBe("json");
    expect(languageOf("post", "post.sh.tmpl")).toBe("shellscript");
    expect(languageOf("netplan.yaml.tmpl")).toBe("yaml");
  });

  it("leaves the rest plain", () => {
    expect(languageOf("ipxe", "ipxe.tmpl")).toBeUndefined();
    expect(languageOf("ipxe", "boot.sh.tmpl")).toBeUndefined();
    expect(languageOf("post", "post.tmpl")).toBeUndefined();
  });
});

describe("translationFailure", () => {
  it("finds the line of the rendered Butane a failed translation names", () => {
    expect(
      translationFailure(
        "failed to translate butane config: template site.tmpl: butane warnings: warning at $.storge, line 3 col 1: unused key storge",
      ),
    ).toEqual({ line: 3 });
  });

  it("is a failed translation without a line when none is named", () => {
    expect(
      translationFailure("failed to translate butane config: template site.tmpl: bad"),
    ).toEqual({ line: undefined });
  });

  it("is nothing for a template that failed to render", () => {
    expect(
      translationFailure('template: site.tmpl:3:15: executing "site.tmpl" at <.host.Nope>: no'),
    ).toBeUndefined();
  });
});

describe("steps", () => {
  it("names a boot file apart from a template kind", () => {
    expect(stepFile(fileStep("initrd-0"))).toBe("initrd-0");
    expect(stepFile("ipxe")).toBeUndefined();
  });
});
