// matchesTags says whether a node with tags passes every filter, as the nodes table filters them: !gpu for a node without gpu, a filter starting or ending with = or : for a node with any tag holding it, like rack= or grendel:, and any other filter for a node with exactly that tag
export function matchesTags(tags: string[] | undefined, filters: string[]): boolean {
  return filters.every((f) => {
    if (!tags) return false;
    if (f.startsWith("!")) return !tags.includes(f.substring(1));
    if (f.startsWith("=") || f.endsWith("=") || f.startsWith(":") || f.endsWith(":"))
      return tags.some((t) => t.includes(f));
    return tags.includes(f);
  });
}
