import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import { useRouter } from "@tanstack/react-router";
import { Hammer } from "lucide-react";
import React from "react";
import { toast } from "sonner";
import { Button } from "./ui/button";

type props = {
  checked: string;
  length: number;
  children: React.ReactNode;
};

export default function ActionsSheet({ checked, length, children }: props) {
  const router = useRouter();
  return (
    <Sheet onOpenChange={(open) => !open && router.invalidate()}>
      <SheetTrigger asChild>
        <Button variant="secondary" className="relative">
          <Hammer />
          <span className="sr-only md:not-sr-only">Actions</span>
          {length > 0 && (
            <span className="absolute -top-1 -right-1 flex size-4">
              <span className="relative inline-flex size-full justify-center rounded-full bg-sky-400 text-xs text-black">
                {Math.abs(length).toString().length > 2 ? "-" : length}
              </span>
            </span>
          )}
        </Button>
      </SheetTrigger>
      <SheetContent className="max-h-dvh w-full overflow-x-scroll sm:max-w-2xl">
        <SheetHeader>
          <SheetTitle>Actions:</SheetTitle>
          <Button
            size="sm"
            variant="secondary"
            onClick={() => {
              navigator.clipboard.writeText(checked);
              toast.success("Successfully copied item(s)");
            }}
          >
            Copy {length} Selected item(s):
          </Button>{" "}
          <SheetDescription className="max-h-36 overflow-y-scroll rounded-md p-2">
            {checked}
          </SheetDescription>
        </SheetHeader>
        {children}
      </SheetContent>
    </Sheet>
  );
}
