import { getV1Roles, patchV1Roles } from "@/client";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import ApiToast from "@/lib/api-toast";
import AuthRedirect from "@/lib/auth";
import { useForm } from "@tanstack/react-form";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/account/roles/$role")({
  component: RouteComponent,
  beforeLoad: AuthRedirect,
  loader: ({ params: { role } }) => getV1Roles({ query: { name: role } }),
});

function RouteComponent() {
  const role = Route.useLoaderData();

  const form = useForm({
    defaultValues: {
      name: role.data?.roles?.[0].name ?? "",
      permission_list: role.data?.roles?.[0].permission_list ?? [],
      unassigned_permission_list: role.data?.roles?.[0].unassigned_permission_list ?? [],
    },
    async onSubmit(props) {
      const v = props.value;
      const res = await patchV1Roles({
        body: { role: v.name, permission_list: v.permission_list },
      });

      ApiToast(res);
    },
  });

  return (
    <div className="flex justify-center p-4">
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
                                form.pushFieldValue("unassigned_permission_list", v);
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
                  <CardTitle className="mx-auto">Unassigned Permissions:</CardTitle>
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
          <Button type="submit" form="permissionForm">
            {/*{form. ? (
              <LoaderCircle className="animate-spin" />
            ) : (*/}
            <span>Submit</span>
            {/*)}*/}
          </Button>
        </CardFooter>
      </Card>
    </div>
  );
}
