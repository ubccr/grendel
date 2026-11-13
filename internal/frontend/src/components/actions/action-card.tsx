import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import {
  Dialog,
  DialogClose,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import React from "react";

type props = {
  title: string;
  description: React.ReactNode;
  children?: React.ReactNode;
  footer: React.ReactNode;
};

// ActionCard pins the footer to the bottom of the card, so the buttons line up across a row of the actions grid however much each card holds
export default function ActionCard({ title, description, children, footer }: props) {
  return (
    <Card className="flex flex-col">
      <CardHeader>
        <CardTitle>{title}</CardTitle>
        <CardDescription>{description}</CardDescription>
      </CardHeader>
      {children && <CardContent className="grid gap-3">{children}</CardContent>}
      <CardFooter className="mt-auto justify-end gap-2">{footer}</CardFooter>
    </Card>
  );
}

export function ActionSection({ title }: { title: string }) {
  return (
    <h3 className="col-span-full mt-2 text-xs font-medium tracking-wide text-muted-foreground uppercase first:mt-0">
      {title}
    </h3>
  );
}

type confirmProps = {
  trigger: React.ReactElement;
  children: React.ReactNode;
  onConfirm: () => void;
};

// ConfirmDialog asks before running onConfirm, with children saying what will happen
export function ConfirmDialog({ trigger, children, onConfirm }: confirmProps) {
  return (
    <Dialog>
      <DialogTrigger render={trigger} />
      <DialogContent className="max-h-dvh overflow-scroll">
        <DialogHeader>
          <DialogTitle>Are you sure?</DialogTitle>
          <DialogDescription render={<div />}>{children}</DialogDescription>
        </DialogHeader>
        <DialogFooter>
          <DialogClose render={<Button variant="destructive" onClick={onConfirm} />}>
            Confirm
          </DialogClose>
          <DialogClose render={<Button />}>Cancel</DialogClose>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
