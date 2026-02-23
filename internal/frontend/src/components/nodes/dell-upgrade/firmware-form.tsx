import { BmcDellInstallFromRepoRequest } from "@/client";
import { postV1BmcUpgradeDellInstallfromrepoMutation } from "@/client/@tanstack/react-query.gen";
import { useAppForm } from "@/components/form/form-context";
import { Button } from "@/components/ui/button";
import { Field, FieldGroup, FieldSeparator, FieldSet } from "@/components/ui/field";
import { useMutation } from "@tanstack/react-query";
import { useRouter } from "@tanstack/react-router";
import { LoaderCircle } from "lucide-react";
import { toast } from "sonner";

export default function FirmwareForm({ nodes }: { nodes: string }) {
  const router = useRouter();
  const mutation = useMutation(postV1BmcUpgradeDellInstallfromrepoMutation());

  const defaultValues: BmcDellInstallFromRepoRequest = {
    IgnoreCertWarning: true,
    IPAddress: "downloads.dell.com",
    ShareType: "HTTPS",
  };
  const form = useAppForm({
    defaultValues: defaultValues,
    onSubmit: (data) => {
      mutation.mutate(
        {
          query: { nodeset: nodes },
          body: data.value,
        },
        {
          onSuccess: () => {
            toast.success("Success", {
              description: "successfully sent request",
            });
            router.invalidate();
          },
          onError: (e) =>
            toast.error(e.title, {
              description: e.detail,
            }),
        },
      );
    },
  });

  const defaultShareTypes = new Map<string, string>([
    ["HTTP", "HTTP"],
    ["HTTPS", "HTTPS"],
    ["NFS", "NFS"],
    ["CIFS", "CIFS"],
    ["FTP", "FTP"],
    ["TFTP", "TFTP"],
  ]);

  return (
    <form
      onSubmit={(e) => {
        e.preventDefault();
        form.handleSubmit(e);
      }}
    >
      <FieldSet>
        <FieldGroup>
          <form.AppField
            name="ShareType"
            children={(field) => (
              <field.SelectField
                label="Share Type"
                description="Type of Network Share"
                items={defaultShareTypes}
              />
            )}
          />
          <form.AppField
            name="IPAddress"
            children={(field) => (
              <field.TextField label="IP Address" description="IP address for the remote share." />
            )}
          />
          <form.AppField
            name="ShareName"
            children={(field) => (
              <field.TextField
                label="Share Name"
                description="Name of the CIFS share or full path to the NFS share. Optional for HTTP/HTTPS share, this may be treated as the path of the directory containing the file."
              />
            )}
          />
          <form.AppField
            name="CatalogFile"
            children={(field) => (
              <field.TextField
                label="Catalog File"
                description="Name of the catalog file on the repository. Default is Catalog.xml."
              />
            )}
          />
          <FieldSeparator />
          <form.AppField
            name="ApplyUpdate"
            children={(field) => (
              <field.SwitchField
                label="Apply Update"
                description="True will start / queue the install jobs, False will only check for updates which will be queriable."
              />
            )}
          />
          <form.AppField
            name="RebootNeeded"
            children={(field) => (
              <field.SwitchField
                label="Reboot"
                description="Automatically reboot the node if needed. Leaving this unchecked will queue the updates for the next reboot."
              />
            )}
          />
          <form.AppField
            name="ClearJobQueue"
            children={(field) => (
              <field.SwitchField
                label="Clear Jobs"
                description="Remove all jobs from the job queue before upgrading. Only applicable when Apply Update = true"
              />
            )}
          />
          <form.AppField
            name="IgnoreCertWarning"
            children={(field) => (
              <field.SwitchField
                label="Ignore Certificate Warning"
                description="Enable to ignore invalid HTTPS certificates."
              />
            )}
          />
          <Field className="justify-end" orientation="responsive">
            <Button type="submit">
              {mutation.isPending ? <LoaderCircle className="animate-spin" /> : <span>Submit</span>}
            </Button>
          </Field>
        </FieldGroup>
      </FieldSet>
    </form>
  );
}
