// patchOverride returns overrides with patch merged into key's entry. A field set to undefined is dropped, and so is an entry left empty, so an edit undone leaves no trace
export function patchOverride<O extends object>(
  overrides: Record<string, O>,
  key: string,
  patch: Partial<O>,
): Record<string, O> {
  const next = { ...overrides[key], ...patch } as Record<string, unknown>;
  for (const k of Object.keys(next)) if (next[k] === undefined) delete next[k];
  const out = { ...overrides };
  if (Object.keys(next).length) out[key] = next as O;
  else delete out[key];
  return out;
}

// withIncluded returns overrides whose include matches ids for each row, kept only where it differs from the row's default
export function withIncluded<O extends { include?: boolean }>(
  overrides: Record<string, O>,
  rows: { key: string; byDefault: boolean }[],
  ids: Set<string>,
): Record<string, O> {
  let out = overrides;
  for (const r of rows) {
    const on = ids.has(r.key);
    out = patchOverride(out, r.key, { include: on === r.byDefault ? undefined : on } as Partial<O>);
  }
  return out;
}
