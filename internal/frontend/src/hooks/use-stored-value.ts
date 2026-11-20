import { useCallback, useMemo, useSyncExternalStore } from "react";

// listeners are the components reading each key, told when this tab writes it; other tabs' writes arrive as storage events
const listeners = new Map<string, Set<() => void>>();

function read(key: string) {
  try {
    return localStorage.getItem(key) ?? undefined;
  } catch {
    return undefined;
  }
}

// useStoredValue is a string kept in this browser's local storage, shared by every component and tab reading the same key, which re-render when it changes. Storage can be missing or blocked, which reads as unset and drops writes
export function useStoredValue(key: string) {
  const subscribe = useCallback(
    (listener: () => void) => {
      const set = listeners.get(key) ?? new Set();
      listeners.set(key, set.add(listener));
      const onStorage = (e: StorageEvent) => (e.key === key || e.key === null) && listener();
      window.addEventListener("storage", onStorage);
      return () => {
        set.delete(listener);
        window.removeEventListener("storage", onStorage);
      };
    },
    [key],
  );
  const value = useSyncExternalStore(subscribe, () => read(key));

  // write stores a value, or removes the key for undefined
  const write = useCallback(
    (next: string | undefined) => {
      try {
        if (next === undefined) localStorage.removeItem(key);
        else localStorage.setItem(key, next);
      } catch {
        return;
      }
      for (const listener of listeners.get(key) ?? []) listener();
    },
    [key],
  );

  return [value, write] as const;
}

// useStoredJSON is a stored object read over defaults, so a field added later falls back to its default and unreadable text reads as the defaults. defaults must keep one reference, such as a module constant
export function useStoredJSON<T extends object>(key: string, defaults: T) {
  const [raw, write] = useStoredValue(key);
  const value = useMemo(() => {
    try {
      return raw ? ({ ...defaults, ...JSON.parse(raw) } as T) : defaults;
    } catch {
      return defaults;
    }
  }, [raw, defaults]);
  const set = useCallback((next: T) => write(JSON.stringify(next)), [write]);
  return [value, set] as const;
}
