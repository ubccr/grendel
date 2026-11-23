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
import { useMutation, type UseMutationOptions } from "@tanstack/react-query";
import { LoaderCircle } from "lucide-react";
import React from "react";
import { toast } from "sonner";

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

type confirmActionProps<TData, TError, TVariables> = {
  title: string;
  description: React.ReactNode;
  label: string;
  destructive?: boolean;
  mutation: UseMutationOptions<TData, TError, TVariables>;
  variables: TVariables;
  children: React.ReactNode;
};

// The server's answer is toasted, unless the mutation is a BMC one, whose per-node outcome is toasted for it
export function ConfirmAction<TData, TError, TVariables>({
  title,
  description,
  label,
  destructive,
  mutation,
  variables,
  children,
}: confirmActionProps<TData, TError, TVariables>) {
  const { mutate, isPending } = useMutation(mutation);
  const announce = (data: unknown) => {
    if (mutation.meta?.bmc !== undefined) return;
    const said = data as { title?: string; detail?: string } | undefined;
    toast.success(said?.title ?? title, { description: said?.detail });
  };

  return (
    <ActionCard
      title={title}
      description={description}
      footer={
        <ConfirmDialog
          trigger={
            <Button variant={destructive ? "destructive" : "default"} disabled={isPending}>
              {isPending && <LoaderCircle className="animate-spin" />}
              {label}
            </Button>
          }
          onConfirm={() => mutate(variables, { onSuccess: announce })}
        >
          {children}
        </ConfirmDialog>
      }
    />
  );
}

type linkActionProps = {
  title: string;
  description: React.ReactNode;
  link: React.ReactElement;
};

export function LinkAction({ title, description, link }: linkActionProps) {
  return (
    <ActionCard
      title={title}
      description={description}
      footer={
        <Button render={link} nativeButton={false}>
          Open
        </Button>
      }
    />
  );
}
