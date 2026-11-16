import { useEffect, useState } from "react";

// useNow is the time, updated every ms milliseconds, so a component showing a relative time like "in 5 minutes" re-renders to keep it current
export function useNow(ms: number) {
  const [now, setNow] = useState(() => Date.now());
  useEffect(() => {
    const id = setInterval(() => setNow(Date.now()), ms);
    return () => clearInterval(id);
  }, [ms]);
  return now;
}
