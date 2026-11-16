import { useEffect, useState } from "react";

// useDebounced follows value once it has stopped changing for ms. The setter it returns takes a value at once, for a change that shouldn't wait, like a button press, while typing still waits
export function useDebounced<T>(value: T, ms: number): [T, (v: T) => void] {
  const [settled, setSettled] = useState(value);
  useEffect(() => {
    const t = setTimeout(() => setSettled(value), ms);
    return () => clearTimeout(t);
  }, [value, ms]);
  return [settled, setSettled];
}
