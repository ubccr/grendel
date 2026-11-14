import {
  useCallback,
  useEffect,
  useRef,
  useState,
  type KeyboardEvent,
  type FocusEvent,
  type PointerEvent,
} from "react";

// useHoverPin opens a disclosure two ways. Hovering with a mouse opens it while the pointer is over it, with a short delay each way so passing over it does not flash it open. Pressing its trigger pins it open until the trigger is pressed again, Escape is pressed, or a click or focus lands outside it, which is how a keyboard or a touch screen opens it. A press while hovering pins it, so it stays once the pointer leaves
export function useHoverPin<C extends HTMLElement, T extends HTMLElement>({
  openDelay = 150,
  closeDelay = 200,
} = {}) {
  const [hovered, setHovered] = useState(false);
  const [pinned, setPinned] = useState(false);
  const open = hovered || pinned;
  const container = useRef<C>(null);
  const trigger = useRef<T>(null);
  const timer = useRef<number>(undefined);

  const hoverLater = (next: boolean, ms: number) => {
    window.clearTimeout(timer.current);
    timer.current = window.setTimeout(() => setHovered(next), ms);
  };
  const close = useCallback(() => {
    window.clearTimeout(timer.current);
    setPinned(false);
    setHovered(false);
  }, []);
  useEffect(() => () => window.clearTimeout(timer.current), []);

  // a pinned disclosure closes on a click or tap anywhere else, as a menu would
  useEffect(() => {
    if (!pinned) return;
    const outside = (e: globalThis.PointerEvent) => {
      if (!container.current?.contains(e.target as Node)) close();
    };
    document.addEventListener("pointerdown", outside);
    return () => document.removeEventListener("pointerdown", outside);
  }, [pinned, close]);

  return {
    open,
    containerProps: {
      ref: container,
      onPointerEnter: (e: PointerEvent<C>) =>
        e.pointerType === "mouse" && hoverLater(true, openDelay),
      onPointerLeave: (e: PointerEvent<C>) =>
        e.pointerType === "mouse" && hoverLater(false, closeDelay),
      onBlur: (e: FocusEvent<C>) => !e.currentTarget.contains(e.relatedTarget) && setPinned(false),
      onKeyDown: (e: KeyboardEvent<C>) => {
        if (e.key !== "Escape" || !open) return;
        close();
        trigger.current?.focus();
      },
    },
    triggerProps: {
      ref: trigger,
      "aria-expanded": open,
      onClick: () => (pinned ? close() : setPinned(true)),
    },
  };
}
