import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import { Button } from "./ui/button";
import { Hammer } from "lucide-react";
import React from "react";

export default function ActionsSheet({
  checked,
  length,
  children,
}: {
  checked: string;
  length: number;
  children: React.ReactNode;
}) {
  return (
    <Sheet>
      <SheetTrigger asChild>
        <Button variant="outline" size="sm" className="relative">
          <Hammer />
          <span className="sr-only md:not-sr-only">Actions</span>
          {length > 0 && (
            <span className="size-4 absolute -right-1 -top-1 flex">
              <span className="size-full relative inline-flex justify-center rounded-full bg-sky-500 text-xs text-black">
                {Math.abs(length).toString().length > 2 ? "-" : length}
              </span>
            </span>
          )}
        </Button>
      </SheetTrigger>
      <SheetContent className="w-full sm:max-w-2xl max-h-dvh overflow-x-scroll">
        <SheetHeader>
          <SheetTitle>Actions:</SheetTitle>
          {/* TODO: add copy button */}
          <SheetDescription className="max-h-36 overflow-y-scroll rounded-md border p-2">
            {length} Selected item(s): <br />
            {checked}
          </SheetDescription>
        </SheetHeader>
        {children}
      </SheetContent>
    </Sheet>
  );
}
