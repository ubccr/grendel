import { islandSpring } from "@/lib/motion";
import { MotionConfig, motion } from "motion/react";
import { AppBreadcrumbs, useCrumbs } from "./app-breadcrumbs";
import { LatestEvent } from "./latest-event";
import { Separator } from "./ui/separator";
import { SidebarTrigger } from "./ui/sidebar";

// Header is a small island at the top left of the page, beside the sidebar, holding the sidebar toggle and where you are. It is only as wide as the breadcrumb and grows or shrinks with it as you move around; the theme and account live in the sidebar footer. The latest event sits in a second island across from it
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
          // the island sits on its own layer above the page, so its shadow always falls on the card below it. Without it the card paints over the shadow at rest, and the layout animation's transform, which lifts the island above the card while it runs, made the shadow flash on and off with every navigation
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
