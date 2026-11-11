import { Card } from "@/components/ui/card";
import {
  Sheet,
  SheetContent,
  SheetDescription,
  SheetHeader,
  SheetTitle,
  SheetTrigger,
} from "@/components/ui/sheet";
import { useRouter } from "@tanstack/react-router";
import { CopyCheckIcon, CopyIcon, Hammer } from "lucide-react";
import React, { useState } from "react";
import { Button } from "./ui/button";
import { ScrollArea } from "./ui/scroll-area";
import NodesetView from "./nodeset-view";

type props = {
  checked: string;
  length: number;
  children: React.ReactNode;
  type?: "other" | "node";
};

function SelectedItems({ checked }: { checked: string }) {
  const [copied, setCopied] = useState(false);
  return (
    <div className="grid gap-2">
      <div className="flex items-center justify-between">
        <span className="text-sm font-medium">Selected</span>
        <Button
          size="icon"
          variant="outline"
          onClick={() => {
            navigator.clipboard.writeText(checked);
            setCopied(true);
            setTimeout(() => setCopied(false), 3000);
          }}
        >
          {copied ? <CopyCheckIcon className="text-green-600" /> : <CopyIcon />}
        </Button>
      </div>
      <ScrollArea className="max-h-36 overflow-y-scroll text-sm break-all text-muted-foreground">
        {checked}
      </ScrollArea>
    </div>
  );
}

export default function ActionsSheet({ checked, length, children, type = "other" }: props) {
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
      <SheetContent className="flex max-h-dvh w-full flex-col overflow-y-auto sm:max-w-2xl">
        <SheetHeader>
          <SheetTitle>Actions</SheetTitle>
          <SheetDescription>
            {length} {type === "node" ? "node" : "item"}
            {length === 1 ? "" : "s"} selected
          </SheetDescription>
        </SheetHeader>
        {checked && (
          <Card className={`bg-muted/40 p-4 ${type === "node" ? "pt-2" : ""}`}>
            {type === "node" ? (
              <NodesetView nodes={checked} size="sm" />
            ) : (
              <SelectedItems checked={checked} />
            )}
          </Card>
        )}
        <div>{children}</div>
      </SheetContent>
    </Sheet>
  );
}
