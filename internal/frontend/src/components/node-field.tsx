import { useEffect, useRef } from "react";

type Props = { full: boolean };

// cell colours per theme; the field is drawn straight onto the page background
const PALETTES = {
  light: { off: "#e4e9e6", amber: "#e39b2d", blue: "#3b7fc4", ready: "#3aa776" },
  dark: { off: "#2d3531", amber: "#e8a33d", blue: "#6ea8e6", ready: "#7fd1ae" },
};
type Palette = (typeof PALETTES)["light"];

const PITCH = 23;
const CELL = 14;
const BASE_ALPHA = 0.75;
const GLOW_ALPHA = 0.85;
const LIFE = 17; // seconds a lit cell takes from orange back to the field colour
// cards float over the field, so cells under them never light up; anything else can opt in with data-node-field-avoid
// a full field is drawn at full strength; a quiet one, behind pages where the work is, has a fainter grid and fewer, dimmer groups
const QUIET = { grid: 0.3, glow: 0.25, spawn: 2 };
const EASE = 0.5; // seconds the field takes to move between full and quiet
// the colours change over seconds, so 30 frames a second looks the same as 60 for half the work
const FRAME_MS = 1000 / 30;
// the cards, sidebar and header are re-measured this often, which keeps up with scrolling and the sidebar sliding without reading layout every frame
const MEASURE_MS = 250;

// routes opt in to the full field through their static data
declare module "@tanstack/react-router" {
  interface StaticDataRouteOption {
    nodeField?: "full";
  }
}

const AVOID = '[data-slot="card"], [data-node-field-avoid]';

type Rect = DOMRect;
type Glow = { x: number; y: number; t0: number };

const smooth = (a: number, b: number, v: number) => {
  const k = Math.min(1, Math.max(0, (v - a) / (b - a)));
  return k * k * (3 - 2 * k);
};

// colour blending in OKLab; a straight blend between opposite hues such as orange and blue passes through a dull grey, so the midpoint is lifted a little in lightness and tinted toward violet, which lands on a soft lavender instead
const toLin = (c: number) => (c <= 0.04045 ? c / 12.92 : ((c + 0.055) / 1.055) ** 2.4);
const toSrgb = (c: number) => (c <= 0.0031308 ? 12.92 * c : 1.055 * c ** (1 / 2.4) - 0.055);
function oklab(hex: string): [number, number, number] {
  const [r, g, b] = [1, 3, 5].map((i) => toLin(parseInt(hex.slice(i, i + 2), 16) / 255));
  const l = Math.cbrt(0.4122214708 * r + 0.5363325363 * g + 0.0514459929 * b);
  const m = Math.cbrt(0.2119034982 * r + 0.6806995451 * g + 0.1073969566 * b);
  const s = Math.cbrt(0.0883024619 * r + 0.2817188376 * g + 0.6299787005 * b);
  return [
    0.2104542553 * l + 0.793617785 * m - 0.0040720468 * s,
    1.9779984951 * l - 2.428592205 * m + 0.4505937099 * s,
    0.0259040371 * l + 0.7827717662 * m - 0.808675766 * s,
  ];
}
function fromOklab([L, A, B]: [number, number, number]) {
  const l = (L + 0.3963377774 * A + 0.2158037573 * B) ** 3;
  const m = (L - 0.1055613458 * A - 0.0638541728 * B) ** 3;
  const s = (L - 0.0894841775 * A - 1.291485548 * B) ** 3;
  const rgb = [
    4.0767416621 * l - 3.3077115913 * m + 0.2309699292 * s,
    -1.2684380046 * l + 2.6097574011 * m - 0.3413193965 * s,
    -0.0041960863 * l - 0.7034186147 * m + 1.707656158 * s,
  ];
  return `rgb(${rgb.map((c) => Math.round(Math.min(1, Math.max(0, toSrgb(Math.max(0, c)))) * 255)).join(",")})`;
}
function mix(a: string, b: string, k: number) {
  const x = oklab(a);
  const y = oklab(b);
  const [L, A, B] = x.map((v, i) => v + (y[i] - v) * k);
  // the lift and tint scale with how far apart the hues are, so a near pair such as blue and green blends straight
  const gap = Math.max(
    0,
    -(x[1] * y[1] + x[2] * y[2]) / (Math.hypot(x[1], x[2]) * Math.hypot(y[1], y[2]) || 1),
  );
  const bump = Math.sin(Math.PI * k) * gap;
  return fromOklab([L + 0.06 * bump, A + 0.035 * bump, B - 0.03 * bump]);
}

// glow is a cell's colour and strength s seconds into its cycle: in as orange, through blue, to green, a long green hold, then back to the field
function glow(s: number, P: Palette): [string, number] {
  if (s < 1) return [P.amber, s];
  if (s < 2) return [P.amber, 1];
  if (s < 4.5) return [mix(P.amber, P.blue, (s - 2) / 2.5), 1];
  if (s < 5.5) return [P.blue, 1];
  if (s < 8) return [mix(P.blue, P.ready, (s - 5.5) / 2.5), 1];
  if (s < 14) return [P.ready, 1];
  return [P.ready, Math.max(0, 1 - (s - 14) / 3)];
}

function cell(ctx: CanvasRenderingContext2D, x: number, y: number, color: string, alpha: number) {
  if (alpha <= 0.01) return;
  ctx.globalAlpha = alpha;
  ctx.fillStyle = color;
  ctx.beginPath();
  ctx.roundRect(x, y, CELL, CELL, 3);
  ctx.fill();
  ctx.globalAlpha = 1;
}

// NodeField is a page background: a faint grid of node cells behind the page, in which small groups of nodes now and then come up together, orange then blue then green, and settle back. It runs under the sidebar and header, and fades out toward every edge, at the top and left underneath those bars. Cells under cards, the sidebar and the header never light up. It is drawn full behind the pages that ask for it and quiet behind the rest, easing between the two. Reduced motion gets a still field with a few lit groups
export function NodeField({ full }: Props) {
  const canvasRef = useRef<HTMLCanvasElement>(null);
  const fullRef = useRef(full);
  const redrawRef = useRef<() => void>(() => {});

  useEffect(() => {
    fullRef.current = full;
    redrawRef.current();
  }, [full]);

  useEffect(() => {
    const canvas = canvasRef.current;
    const ctx = canvas?.getContext("2d");
    if (!canvas || !ctx) return;

    const still = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    const root = document.documentElement;
    const palette = () => (root.classList.contains("dark") ? PALETTES.dark : PALETTES.light);
    let P = palette();
    let w = 0;
    let h = 0;
    let dpr = 1;
    let fadeLeft = 0;
    let fadeTop = 0;
    let avoid: Rect[] = [];
    let glows: Glow[] = [];
    let next = 0.5;
    let frame = 0;
    // strength runs from 0, quiet, to 1, full
    let strength = fullRef.current ? 1 : 0;
    let last = performance.now();
    const began = last;
    const lerp = (quiet: number) => quiet + (1 - quiet) * strength;
    // the static grid is drawn once into its own canvas and copied each frame, so only the lit cells are redrawn
    const base = document.createElement("canvas");
    const baseCtx = base.getContext("2d");
    let baseKey = "";

    // the top and left fades finish under the sidebar and header, so the field comes out from under them faint
    const edge = (x: number, y: number) =>
      smooth(fadeLeft - 70, fadeLeft + 110, x) *
      smooth(fadeTop - 40, fadeTop + 90, y) *
      (1 - smooth(0.62, 1, x / w)) *
      (1 - smooth(0.55, 1, y / h));

    const measure = () => {
      const side = document.querySelector('[data-slot="sidebar-inner"]')?.getBoundingClientRect();
      const header = document.querySelector('[data-slot="app-header"]')?.getBoundingClientRect();
      avoid = [...document.querySelectorAll(AVOID)].map((el) => el.getBoundingClientRect());
      if (side) avoid.push(side);
      if (header) avoid.push(header);
      avoid = avoid.filter((r) => r.width > 0 && r.height > 0);
      fadeLeft = side && side.width > 0 ? Math.round(side.right) : 0;
      fadeTop = header && header.height > 0 ? Math.round(header.bottom) : 0;
    };

    const under = (x: number, y: number) =>
      avoid.some(
        (r) => x + CELL > r.left - 6 && x < r.right + 6 && y + CELL > r.top - 6 && y < r.bottom + 6,
      );

    const drawBase = () => {
      const key = `${w}x${h}@${dpr}:${fadeLeft}:${fadeTop}:${P.off}`;
      if (key === baseKey || !baseCtx) return;
      baseKey = key;
      base.width = Math.round(w * dpr);
      base.height = Math.round(h * dpr);
      baseCtx.setTransform(dpr, 0, 0, dpr, 0, 0);
      baseCtx.clearRect(0, 0, w, h);
      for (let y = 4; y < h; y += PITCH) {
        for (let x = 4; x < w; x += PITCH) cell(baseCtx, x, y, P.off, BASE_ALPHA * edge(x, y));
      }
    };

    const resize = () => {
      dpr = window.devicePixelRatio || 1;
      w = window.innerWidth;
      h = window.innerHeight;
      canvas.width = Math.round(w * dpr);
      canvas.height = Math.round(h * dpr);
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
    };

    // a group starts from a random open cell, and 4 to 12 of its neighbours follow, nearest first with some jitter and each a little later than the one before
    const spawn = (t: number) => {
      // a group must land clear of whatever is on the page now
      measure();
      const cols = Math.floor((w - 4) / PITCH) + 1;
      const rows = Math.floor((h - 4) / PITCH) + 1;
      for (let attempt = 0; attempt < 30; attempt++) {
        const sc = Math.floor(Math.random() * cols);
        const sr = Math.floor(Math.random() * rows);
        const sx = 4 + sc * PITCH;
        const sy = 4 + sr * PITCH;
        if (under(sx, sy) || edge(sx, sy) < 0.4) continue;
        const reach = 2.2 + Math.random() * 1.2;
        const near: [number, number, number][] = [];
        for (let dr = -3; dr <= 3; dr++) {
          for (let dc = -3; dc <= 3; dc++) {
            const d = Math.hypot(dc, dr);
            const x = sx + dc * PITCH;
            const y = sy + dr * PITCH;
            if (d === 0 || d > reach || x < 0 || y < 0 || x > w || y > h || under(x, y)) continue;
            near.push([x, y, d + Math.random() * 1.2]);
          }
        }
        near.sort((a, b) => a[2] - b[2]);
        glows.push({ x: sx, y: sy, t0: t });
        for (const [x, y] of near.slice(0, 4 + Math.floor(Math.random() * 9))) {
          glows.push({
            x,
            y,
            t0: t + (Math.hypot(x - sx, y - sy) / PITCH) * 0.35 + Math.random() * 0.6,
          });
        }
        return;
      }
    };

    const render = (t: number) => {
      drawBase();
      ctx.clearRect(0, 0, w, h);
      ctx.globalAlpha = lerp(QUIET.grid);
      ctx.drawImage(base, 0, 0, w, h);
      ctx.globalAlpha = 1;
      const glowAlpha = GLOW_ALPHA * lerp(QUIET.glow);
      for (const g of glows) {
        const s = t - g.t0;
        if (s < 0) continue;
        const [color, a] = glow(s, P);
        cell(ctx, g.x, g.y, color, glowAlpha * a * edge(g.x, g.y));
      }
    };

    const drawStill = () => {
      measure();
      glows = [];
      for (let i = 0; i < 3; i++) spawn(0);
      // freeze the groups at green, blue and orange
      const stages = [10, 5, 1.5];
      let group = -1;
      glows = glows.map((g) => {
        if (g.t0 === 0) group++;
        return { ...g, t0: -stages[Math.max(0, group) % stages.length] };
      });
      render(0);
    };

    const onResize = () => {
      resize();
      measure();
      if (still) drawStill();
    };
    window.addEventListener("resize", onResize);
    resize();

    // the theme switch toggles a class on the root element
    const themeObserver = new MutationObserver(() => {
      P = palette();
      if (still) {
        measure();
        render(0);
      }
    });
    themeObserver.observe(root, { attributes: true, attributeFilter: ["class"] });

    if (still) {
      redrawRef.current = () => {
        strength = fullRef.current ? 1 : 0;
        measure();
        render(0);
      };
      drawStill();
    } else {
      let measured = 0;
      const loop = (now: number) => {
        frame = requestAnimationFrame(loop);
        if (now - last < FRAME_MS) return;
        if (now - measured > MEASURE_MS) {
          measure();
          measured = now;
        }
        const t = (now - began) / 1000;
        const step = Math.min(0.1, (now - last) / 1000) / EASE;
        last = now;
        strength = fullRef.current ? Math.min(1, strength + step) : Math.max(0, strength - step);
        if (t > next) {
          spawn(t);
          next = t + (3.8 + Math.random() * 2.6) * lerp(QUIET.spawn);
        }
        glows = glows.filter((g) => t - g.t0 < LIFE);
        render(t);
      };
      frame = requestAnimationFrame(loop);
    }

    return () => {
      window.removeEventListener("resize", onResize);
      themeObserver.disconnect();
      cancelAnimationFrame(frame);
    };
  }, []);

  return (
    <canvas
      ref={canvasRef}
      data-slot="node-field"
      aria-hidden
      className="pointer-events-none fixed inset-0 -z-10 size-full"
    />
  );
}
