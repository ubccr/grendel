import { Button } from "@/components/ui/button";
import { Label } from "@/components/ui/label";
import { cn } from "@/lib/utils";
import { LoaderCircle } from "lucide-react";
import { useId, type ReactNode } from "react";

// cellInput styles an input in a dense table of rows, like a node's ports or an image's initrds: a faint filled box until it is hovered or focused, so empty cells are still visible without the table becoming a wall of outlines
export const cellInput =
  "h-8 border-transparent bg-muted/60 px-2 font-mono shadow-none hover:border-input focus-visible:border-input focus-visible:bg-transparent max-lg:border-input aria-invalid:border-destructive";

// invalidInput outlines an input whose field failed validation, which it marks with aria-invalid
export const invalidInput =
  "aria-invalid:border-destructive aria-invalid:focus-visible:ring-destructive";

// required checks that a field isn't empty or blank, refusing it with message
export function required(message: string) {
  return ({ value }: { value: unknown }) =>
    typeof value === "string" && value.trim() !== "" ? undefined : message;
}

type Check<T> = (props: { value: T }) => string | undefined;
type FieldProps<T> = { value: T; fieldApi: { form: { state: { submissionAttempts: number } } } };

// validateOnSave runs check when the form is saved, and again on every change once a save has been refused, so a field is only marked once the person has tried to save and is cleared as soon as they fix it. A row that was just added isn't marked while it is still being filled in
export function validateOnSave<T>(check: Check<T>) {
  return {
    onSubmit: check,
    onChange: ({ value, fieldApi }: FieldProps<T>) =>
      fieldApi.form.state.submissionAttempts > 0 ? check({ value }) : undefined,
  };
}

// FieldErrors shows why a field failed validation, the first reason being enough to fix it
export function FieldErrors({ errors }: { errors: unknown[] }) {
  const first = errors.find((e) => typeof e === "string" && e);
  if (!first) return null;
  return <p className="text-xs text-destructive">{first as string}</p>;
}

// Section is a titled part of an edit form, with an optional action such as an add button at its end
export function Section({
  title,
  description,
  action,
  children,
}: {
  title: string;
  description?: ReactNode;
  action?: ReactNode;
  children: ReactNode;
}) {
  return (
    <section className="grid gap-4">
      <div className="flex flex-wrap items-end justify-between gap-3">
        <div className="grid gap-1">
          <h2 className="text-base font-semibold tracking-tight">{title}</h2>
          {description && <p className="text-sm text-muted-foreground">{description}</p>}
        </div>
        {action}
      </div>
      {children}
    </section>
  );
}

// Field labels one input, handing it the id the label points at
export function Field({
  label,
  className,
  children,
}: {
  label: string;
  className?: string;
  children: (id: string) => ReactNode;
}) {
  const id = useId();
  return (
    <div className={cn("grid content-start gap-1.5", className)}>
      <Label htmlFor={id}>{label}</Label>
      {children(id)}
    </div>
  );
}

// SaveBar sits at the bottom of an edit form while it has unsaved changes, or always when the form adds something new. noun names what is saved, as in "Add node". invalid says a save was refused because fields failed validation
export function SaveBar({
  noun,
  adding,
  pending,
  invalid,
  onDiscard,
}: {
  noun: string;
  adding: boolean;
  pending: boolean;
  invalid?: boolean;
  onDiscard: () => void;
}) {
  return (
    <div className="sticky bottom-3 z-10 flex flex-wrap items-center justify-between gap-3 rounded-lg border bg-background/95 px-4 py-2.5 shadow-lg backdrop-blur supports-[backdrop-filter]:bg-background/80">
      {invalid ? (
        <span className="text-sm text-destructive" role="alert">
          Fill in or remove the fields marked in red.
        </span>
      ) : (
        <span className="text-sm text-muted-foreground">
          {adding ? `The ${noun} is added when you save it.` : "You have unsaved changes."}
        </span>
      )}
      <div className="flex gap-2">
        {!adding && (
          <Button type="button" variant="ghost" onClick={onDiscard} disabled={pending}>
            Discard
          </Button>
        )}
        <Button type="submit" disabled={pending}>
          {pending && <LoaderCircle className="animate-spin" />}
          {adding ? `Add ${noun}` : "Save changes"}
        </Button>
      </div>
    </div>
  );
}
