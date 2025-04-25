import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useGetV1Roles, usePatchV1Roles } from "@/openapi/queries";
import { GetRolesResponse } from "@/openapi/requests";
import { useForm } from "@tanstack/react-form";
import { useQueryClient } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { LoaderCircle } from "lucide-react";
import { toast } from "sonner";

export const Route = createFileRoute("/account/roles/$role")({
  component: RouteComponent,
});

function RouteComponent() {
  const { role } = Route.useParams();
  const role_query = useGetV1Roles({ query: { name: role } });

  return (
    <div>
      {role_query.data?.roles?.[0] != undefined && (
        <PermissionForm role={role_query.data?.roles[0]} />
      )}
    </div>
  );
}

function PermissionForm({
  role,
}: {
  role: NonNullable<GetRolesResponse["roles"]>[number];
}) {
  const { mutate, isPending } = usePatchV1Roles();
  const queryClient = useQueryClient();

  const form = useForm({
    defaultValues: {
      name: role.name ?? "",
      permission_list: role.permission_list ?? [],
      unassigned_permission_list: role.unassigned_permission_list ?? [],
    },
    onSubmit(props) {
      const v = props.value;
      mutate(
        { body: { role: v.name, permission_list: v.permission_list } },
        {
          onSuccess: ({ data }) => {
            toast.success(data?.title, {
              description: data?.detail,
            });
            queryClient.invalidateQueries();
          },
          onError: (e) => {
            toast.error(e.title, {
              description: e.detail,
            });
          },
        }
      );
    },
  });

  return (
    <div className="p-4 flex justify-center">
      <Card>
        <CardContent>
          <form
            id="permissionForm"
            onSubmit={(e) => {
              e.preventDefault();
              e.stopPropagation();
              form.handleSubmit();
            }}
          >
            <div className="grid grid-cols-2 gap-4 p-4">
              <Card>
                <CardHeader>
                  <CardTitle className="mx-auto">Permissions:</CardTitle>
                </CardHeader>
                <CardContent>
                  <form.Field
                    name="permission_list"
                    mode="array"
                    children={(field) => (
                      <>
                        {field.state.value?.map((v, i) => (
                          <div key={i}>
                            <button
                              type="button"
                              className="hover:font-medium"
                              onClick={() => {
                                form.pushFieldValue(
                                  "unassigned_permission_list",
                                  v
                                );
                                field.removeValue(i);
                              }}
                            >
                              {v.method} {v.path}
                            </button>
                          </div>
                        ))}
                      </>
                    )}
                  />
                </CardContent>
              </Card>
              <Card>
                <CardHeader>
                  <CardTitle className="mx-auto">
                    Unassigned Permissions:
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <form.Field
                    name="unassigned_permission_list"
                    mode="array"
                    children={(field) => (
                      <>
                        {field.state.value?.map((v, i) => (
                          <div key={i}>
                            <button
                              type="button"
                              className="hover:font-medium"
                              onClick={() => {
                                form.pushFieldValue("permission_list", v);
                                field.removeValue(i);
                              }}
                            >
                              {v.method} {v.path}
                            </button>
                          </div>
                        ))}
                      </>
                    )}
                  />
                </CardContent>
              </Card>
            </div>
          </form>
        </CardContent>
        <CardFooter>
          <Button
            type="submit"
            size="sm"
            // variant="outline"
            form="permissionForm"
          >
            {isPending ? (
              <LoaderCircle className="animate-spin" />
            ) : (
              <span>Submit</span>
            )}
          </Button>
        </CardFooter>
      </Card>
    </div>
  );
}
