import { islandSpring } from "@/lib/motion";
import { MotionConfig, motion } from "motion/react";
import { AppBreadcrumbs, useCrumbs } from "./app-breadcrumbs";
import { LatestEvent } from "./latest-event";
import { Separator } from "./ui/separator";
import { SidebarTrigger } from "./ui/sidebar";

export default function Header() {
  const crumbs = useCrumbs();
  return (
    <MotionConfig reducedMotion="user">
      <div className="my-2 flex gap-2">
        {/* the radius matches the sidebar and cards' rounded-xl, var(--radius) + 4px; it is set as a style in pixels because motion only keeps a radius steady while the island changes size when it can read it there */}
        <motion.header
          layout
          transition={islandSpring}
          style={{ borderRadius: 14 }}
          data-slot="app-header"
          // the island sits on its own layer above the page, so its shadow always falls on the card below it. Without the layer the card paints over the shadow at rest and the layout animation's transform lifts it back above, so the shadow flashes on every navigation
          className="relative z-10 flex h-10 max-w-full min-w-0 items-center bg-sidebar px-1 text-sidebar-foreground shadow-float ring-1 ring-sidebar-border"
        >
          {/* children of a layout animation are kept unstretched by giving them a layout of their own */}
          <motion.div layout="position" className="flex min-w-0 items-center gap-2">
            <SidebarTrigger />
            {/* the divider comes with the crumbs, so the island is just the toggle when there are none */}
            {crumbs.length > 0 && (
              <>
                <Separator
                  orientation="vertical"
                  className="data-vertical:h-4 data-vertical:self-center"
                />
                <AppBreadcrumbs crumbs={crumbs} className="pr-3" />
              </>
            )}
          </motion.div>
        </motion.header>
        <LatestEvent />
      </div>
    </MotionConfig>
  );
}
