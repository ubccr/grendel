import { BootImage } from "@/client";
import { getV1ImagesQueryKey, postV1ImagesMutation } from "@/client/@tanstack/react-query.gen";
import {
  cellInput,
  Field,
  FieldErrors,
  invalidInput,
  required,
  SaveBar,
  validateOnSave,
  Section,
} from "@/components/form-layout";
import { TemplateHelp } from "@/components/images/template-help";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Switch } from "@/components/ui/switch";
import { Textarea } from "@/components/ui/textarea";
import { cn } from "@/lib/utils";
import { useForm } from "@tanstack/react-form";
import { useMutation, useQueryClient } from "@tanstack/react-query";
import { useNavigate, useRouter } from "@tanstack/react-router";
import { Plus, X } from "lucide-react";
import { useEffect } from "react";
import { toast } from "sonner";

type Templates = NonNullable<BootImage["provision_templates"]>;

// renameTemplate gives a template a new name and keeps it in its place, so a row doesn't jump while its name is typed. A name another template already has is refused, since the two would merge into one
// incompleteTemplates refuses a template without a name or a file, since the API would store it as is and a node asking for it would fail to boot
export function incompleteTemplates({ value }: { value?: Templates }) {
  return Object.entries(value ?? {}).some(([name, file]) => !name.trim() || !file.trim())
    ? "Every template needs a name and a file."
    : undefined;
}

const nameRequired = required("Enter a name.");
const kernelRequired = required("Enter the kernel's path.");
const initrdRequired = required("Enter a path or remove this initrd.");

export function renameTemplate(
  templates: Templates,
  from: string,
  to: string,
): Templates | undefined {
  if (to !== from && to in templates) return undefined;
  return Object.fromEntries(Object.entries(templates).map(([k, v]) => [k === from ? to : k, v]));
}

export default function ImageForm({ data, reset }: { data?: BootImage; reset?: boolean }) {
  const router = useRouter();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { mutateAsync, isPending } = useMutation(postV1ImagesMutation());
  const adding = data == undefined;

  const form = useForm({
    defaultValues: data,
    onSubmit: async ({ value }) => {
      if (value == undefined) return;
      await mutateAsync(
        { body: { boot_images: [value] } },
        {
          onSuccess: (res) => {
            toast.success(res?.title, { description: res?.detail });
            if (!adding) form.reset(value);
            // a renamed image is only found under its new name, so the page moves there rather than reloading the old one
            if (!adding && value.name !== data?.name)
              navigate({ to: "/images/$image", params: { image: value.name } });
            else router.invalidate();
            queryClient.invalidateQueries({ queryKey: getV1ImagesQueryKey() });
          },
          onError: (e) => toast.error(e.title, { description: e.detail }),
        },
      );
    },
  });

  useEffect(() => {
    form.reset();
  }, [reset]);

  return (
    <form
      className="grid gap-10 pt-3"
      onSubmit={(e) => {
        e.preventDefault();
        e.stopPropagation();
        form.handleSubmit();
      }}
    >
      <div className="grid gap-x-6 gap-y-5 sm:grid-cols-2 lg:grid-cols-3">
        <form.Field name="name" validators={validateOnSave(nameRequired)}>
          {(field) => (
            <Field label="Name">
              {(id) => (
                <>
                  <Input
                    id={id}
                    className={cn("font-mono", invalidInput)}
                    aria-invalid={field.state.meta.errors.length > 0}
                    value={field.state.value ?? ""}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                  />
                  <FieldErrors errors={field.state.meta.errors} />
                </>
              )}
            </Field>
          )}
        </form.Field>
        <form.Field name="version">
          {(field) => (
            <Field label="Version">
              {(id) => (
                <Input
                  id={id}
                  value={field.state.value ?? ""}
                  onBlur={field.handleBlur}
                  onChange={(e) => field.handleChange(e.target.value)}
                />
              )}
            </Field>
          )}
        </form.Field>
        <form.Field name="verify">
          {(field) => (
            <Field label="Verify signatures">
              {(id) => (
                <div className="flex h-9 items-center gap-2">
                  <Switch
                    id={id}
                    className="data-[state=checked]:bg-emerald-600"
                    checked={field.state.value ?? false}
                    onBlur={field.handleBlur}
                    onCheckedChange={(e) => field.handleChange(e)}
                  />
                  <span className="text-sm">{field.state.value ? "On" : "Off"}</span>
                </div>
              )}
            </Field>
          )}
        </form.Field>
      </div>

      <Section title="Boot files" action={<TemplateHelp />}>
        <form.Field name="kernel" validators={validateOnSave(kernelRequired)}>
          {(field) => (
            <Field label="Kernel">
              {(id) => (
                <>
                  <Input
                    id={id}
                    className={cn("font-mono", invalidInput)}
                    aria-invalid={field.state.meta.errors.length > 0}
                    value={field.state.value ?? ""}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                  />
                  <FieldErrors errors={field.state.meta.errors} />
                </>
              )}
            </Field>
          )}
        </form.Field>
        <form.Field name="cmdline">
          {(field) => (
            <Field label="Command line">
              {(id) => (
                <Textarea
                  id={id}
                  rows={3}
                  className="font-mono text-sm"
                  value={field.state.value ?? ""}
                  onBlur={field.handleBlur}
                  onChange={(e) => field.handleChange(e.target.value)}
                />
              )}
            </Field>
          )}
        </form.Field>
        <form.Field name="initrd" mode="array">
          {(field) => {
            const initrds = field.state.value ?? [];
            return (
              <div className="grid gap-1.5">
                <div className="flex items-center justify-between gap-3">
                  <span className="text-sm font-medium">Initrds</span>
                  <Button type="button" variant="secondary" onClick={() => field.pushValue("")}>
                    <Plus />
                    Add initrd
                  </Button>
                </div>
                {initrds.length === 0 ? (
                  <p className="rounded-md border border-dashed px-4 py-4 text-center text-sm text-muted-foreground">
                    No initrds. The kernel boots on its own.
                  </p>
                ) : (
                  <ol className="grid gap-1.5" aria-label="Initrds">
                    {initrds.map((_, i) => (
                      <li key={i} className="flex items-start gap-2">
                        <span className="w-5 shrink-0 pt-2 text-right text-xs text-muted-foreground tabular-nums">
                          {i + 1}
                        </span>
                        <form.Field
                          name={`initrd[${i}]`}
                          validators={validateOnSave(initrdRequired)}
                        >
                          {(f) => (
                            <div className="grid min-w-0 flex-1 gap-1">
                              <Input
                                aria-label={`Initrd ${i + 1}`}
                                aria-invalid={f.state.meta.errors.length > 0}
                                className={cn(cellInput, "max-lg:border-transparent")}
                                value={f.state.value ?? ""}
                                onChange={(e) => f.handleChange(e.target.value)}
                              />
                              <FieldErrors errors={f.state.meta.errors} />
                            </div>
                          )}
                        </form.Field>
                        <div className="flex shrink-0">
                          <Button
                            type="button"
                            size="icon"
                            variant="ghost"
                            className="hover:bg-destructive/10 hover:text-destructive"
                            onClick={() => field.removeValue(i)}
                          >
                            <X />
                            <span className="sr-only">Remove initrd {i + 1}</span>
                          </Button>
                        </div>
                      </li>
                    ))}
                  </ol>
                )}
              </div>
            );
          }}
        </form.Field>
      </Section>
      <form.Field name="provision_templates" validators={validateOnSave(incompleteTemplates)}>
        {(field) => {
          const templates = field.state.value ?? {};
          const entries = Object.entries(templates);
          const invalid = field.state.meta.errors.length > 0;
          return (
            <Section
              title="Provision templates"
              description="File location is relative to configured template directory"
              action={
                <Button
                  type="button"
                  variant="secondary"
                  disabled={"" in templates}
                  onClick={() => field.handleChange({ ...templates, "": "" })}
                >
                  <Plus />
                  Add template
                </Button>
              }
            >
              {entries.length === 0 ? (
                <p className="rounded-md border border-dashed px-4 py-6 text-center text-sm text-muted-foreground">
                  No templates. The image boots with Grendel's default templates.
                </p>
              ) : (
                <div className="grid text-sm" role="table" aria-label="Provision templates">
                  <div
                    role="row"
                    className="hidden grid-cols-[minmax(8rem,1fr)_minmax(12rem,2fr)_2.25rem] gap-2 border-b px-2 pb-2 text-xs text-muted-foreground sm:grid"
                  >
                    <span role="columnheader" className="px-2">
                      Name
                    </span>
                    <span role="columnheader" className="px-2">
                      Template file
                    </span>
                    <span role="columnheader">
                      <span className="sr-only">Actions</span>
                    </span>
                  </div>
                  {entries.map(([name, file], i) => (
                    <div
                      key={i}
                      role="row"
                      className="grid grid-cols-[1fr_2.25rem] items-center gap-2 border-b px-2 py-3 last:border-b-0 sm:grid-cols-[minmax(8rem,1fr)_minmax(12rem,2fr)_2.25rem] sm:py-1.5"
                    >
                      <Input
                        role="cell"
                        aria-label={`Template ${i + 1} name`}
                        aria-invalid={invalid && !name.trim()}
                        className={cn(cellInput, "max-lg:border-transparent")}
                        value={name}
                        onChange={(e) => {
                          const next = renameTemplate(templates, name, e.target.value);
                          if (next) field.handleChange(next);
                        }}
                      />
                      <Input
                        role="cell"
                        aria-label={`Template ${i + 1} file`}
                        aria-invalid={invalid && !file.trim()}
                        className={cn(cellInput, "max-lg:border-transparent max-sm:col-start-1")}
                        value={file}
                        onChange={(e) =>
                          field.handleChange({ ...templates, [name]: e.target.value })
                        }
                      />
                      <Button
                        type="button"
                        size="icon"
                        variant="ghost"
                        className="hover:bg-destructive/10 hover:text-destructive max-sm:col-start-2 max-sm:row-span-2 max-sm:row-start-1"
                        onClick={() => {
                          const { [name]: _, ...rest } = templates;
                          field.handleChange(rest);
                        }}
                      >
                        <X />
                        <span className="sr-only">Remove template {name || i + 1}</span>
                      </Button>
                    </div>
                  ))}
                  <div className="px-2 pt-1">
                    <FieldErrors errors={field.state.meta.errors} />
                  </div>
                </div>
              )}
            </Section>
          );
        }}
      </form.Field>

      <form.Subscribe
        selector={(s) => [s.isDirty, s.submissionAttempts > 0 && !s.isValid] as const}
      >
        {([dirty, invalid]) =>
          (dirty || adding) && (
            <SaveBar
              noun="image"
              adding={adding}
              pending={isPending}
              invalid={invalid}
              onDiscard={() => form.reset()}
            />
          )
        }
      </form.Subscribe>
    </form>
  );
}
