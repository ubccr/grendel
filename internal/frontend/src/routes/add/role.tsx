import { postV1RolesMutation } from "@/client/@tanstack/react-query.gen";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useForm } from "@tanstack/react-form";
import { useMutation } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { LoaderCircle } from "lucide-react";
import { toast } from "sonner";

export const Route = createFileRoute("/add/role")({
  component: RouteComponent,
});

function RouteComponent() {
  const navigate = Route.useNavigate();
  const { mutate, isPending } = useMutation(postV1RolesMutation());

  const form = useForm({
    defaultValues: {
      name: "",
      inherit: "",
    },
    onSubmit: async ({ value }) => {
      mutate(
        { body: { role: value.name, inherited_role: value.inherit } },
        {
          onSuccess: (data) => {
            toast.success(data?.title, {
              description: data?.detail,
            });

            navigate({
              to: "/account/roles",
            });
          },
          onError: (e) => {
            toast.error(e.title, {
              description: e.detail,
            });
          },
        },
      );
    },
  });
  return (
    <div className="flex justify-center">
      <form
        className="grid grid-cols-1 gap-4"
        onSubmit={(e) => {
          e.preventDefault();
          e.stopPropagation();
          form.handleSubmit();
        }}
      >
        <Card>
          <CardHeader>
            <CardTitle>Add a Role:</CardTitle>
            <CardDescription>
              New Roles can be created from an inherited role, which will copy all permissions into
              the new role.
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form.Field
              name="name"
              children={(field) => (
                <div>
                  <Label>Name:</Label>
                  <Input
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                  />
                </div>
              )}
            />
            <form.Field
              name="inherit"
              children={(field) => (
                <div>
                  <Label>Inherited role:</Label>
                  <Input
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                  />
                </div>
              )}
            />
          </CardContent>
          <CardFooter>
            <Button type="submit">
              {isPending ? <LoaderCircle className="animate-spin" /> : <span>Submit</span>}
            </Button>
          </CardFooter>
        </Card>
      </form>
    </div>
  );
}
