import { Card, CardDescription } from "@/components/ui/card";
import "@fontsource/unifrakturcook/latin-700.css";
import type { ReactNode } from "react";

type Props = {
  title: string;
  description: ReactNode;
  children: ReactNode;
  footer: ReactNode;
};

// AuthCard sets the sign-in and sign-up forms on a card floating over the node field, under a blackletter wordmark in the standard text colour
export function AuthCard({ title, description, children, footer }: Props) {
  return (
    <div className="flex flex-1 items-start justify-center px-2 py-6 sm:items-center sm:pt-10 sm:pb-[8vh]">
      <Card className="w-full max-w-md gap-6 px-6 py-6 sm:px-10 sm:py-9">
        <div className="grid gap-1.5">
          <p
            aria-hidden
            className="mb-3 text-center font-[UnifrakturCook] text-5xl leading-none text-foreground"
          >
            Grendel
          </p>
          {/* the page heading; the wordmark above it is decoration */}
          <h1 className="text-xl leading-snug font-semibold">{title}</h1>
          <CardDescription className="text-pretty">{description}</CardDescription>
        </div>
        {children}
        <p className="text-center text-sm text-muted-foreground">{footer}</p>
      </Card>
    </div>
  );
}
