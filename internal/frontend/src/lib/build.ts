import type { Version } from "@/client";

export const REPO = "https://github.com/ubccr/grendel";

// a release is built with its tag as the version, with or without the leading v, and anything else, such as vDEV, has no release page
export function releaseUrl(version?: string) {
  const release = version?.match(/^v?(\d+\.\d+\.\d+)$/);
  return release ? `${REPO}/releases/tag/v${release[1]}` : undefined;
}

// buildSummary is the build in one line, to paste into a bug report
export function buildSummary(build: Version) {
  const commit = build.commit && `${build.commit.slice(0, 12)}${build.modified ? "+changes" : ""}`;
  return [
    `Grendel ${build.Version ?? "unknown"}`,
    commit && `commit ${commit}`,
    build.go_version,
    build.schema && `schema ${build.schema}`,
  ]
    .filter(Boolean)
    .join(", ");
}
