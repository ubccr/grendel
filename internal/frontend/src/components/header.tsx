import { MotionConfig, motion } from "motion/react";
import { AppBreadcrumbs, useCrumbs } from "./app-breadcrumbs";
import { Separator } from "./ui/separator";
import { SidebarTrigger } from "./ui/sidebar";

// Header is a small island at the top left of the page, beside the sidebar, holding the sidebar toggle and where you are. It is only as wide as the breadcrumb and grows or shrinks with it as you move around; the theme and account live in the sidebar footer
export default function Header() {
  const crumbs = useCrumbs();
  return (
    <MotionConfig reducedMotion="user">
      <div className="my-2 flex">
        {/* the radius matches the sidebar and cards' rounded-xl, var(--radius) + 4px; it is set as a style in pixels because motion only keeps a radius steady while the island changes size when it can read it there */}
        <motion.header
          layout
          transition={{ type: "spring", bounce: 0.15, duration: 0.45 }}
          style={{ borderRadius: 14 }}
          data-slot="app-header"
          className="flex h-10 max-w-full min-w-0 items-center bg-sidebar px-1 text-sidebar-foreground shadow-float ring-1 ring-sidebar-border"
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
      </div>
    </MotionConfig>
  );
}
