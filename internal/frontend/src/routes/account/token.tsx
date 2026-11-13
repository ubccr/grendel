import { postV1AuthTokenMutation } from "@/client/@tanstack/react-query.gen";
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
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { FieldError } from "@/components/ui/field";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { ScrollArea } from "@/components/ui/scroll-area";
import AuthRedirect from "@/lib/auth";
import { useForm } from "@tanstack/react-form";
import { useMutation } from "@tanstack/react-query";
import { createFileRoute } from "@tanstack/react-router";
import { Copy, LoaderCircle } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import { z } from "zod";

export const Route = createFileRoute("/account/token")({
  component: RouteComponent,
  beforeLoad: AuthRedirect,
});

// one or more number-and-unit pairs, such as 30m or 1h30m, which time.ParseDuration on the server accepts, or "infinite". No g flag: zod reuses the regex, and a global regex keeps its lastIndex between tests, so a valid value would fail every other check
const EXPIRE_REGEX = /^(\d+[hms])+$|^infinite$/;

const reqSchema = z.object({
  username: z.string(),
  role: z.string(),
  expire: z.string().regex(EXPIRE_REGEX, {
    message: "Invalid duration. Follow the Go time.ParseDuration syntax, ex: 30m",
  }),
});

function RouteComponent() {
  const [resDialog, setResDialog] = useState(false);
  const { mutate, isPending, data } = useMutation(postV1AuthTokenMutation());

  const form = useForm({
    defaultValues: {
      username: "",
      role: "",
      expire: "",
    },
    validators: {
      onSubmit: reqSchema,
    },
    onSubmit: async ({ value }) => {
      mutate(
        {
          body: {
            username: value.username,
            role: value.role,
            expire: value.expire,
          },
        },
        {
          onSuccess: () => {
            toast.success("Successfully created token");
            setResDialog(true);
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
            <CardTitle>Create an API Token:</CardTitle>
            <CardDescription>
              API Tokens can be used to authenticate the CLI or custom applications that integrate
              with Grendel.
              <br />
              <br />
              Valid options:
              <br />
              Username: string, must be a valid user
              <br />
              Role: string, built in roles: "admin", "user", "read-only"
              <br />
              Expire: duration before token expires, ex "8h", "1h", "30m", "infinite"
            </CardDescription>
          </CardHeader>
          <CardContent>
            <form.Field
              name="username"
              children={(field) => (
                <div>
                  <Label>Username:</Label>
                  <Input
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                  />
                  <FieldError errors={field.state.meta.errors} />
                </div>
              )}
            />
            <form.Field
              name="role"
              children={(field) => (
                <div>
                  <Label>Role:</Label>
                  <Input
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                  />
                  <FieldError errors={field.state.meta.errors} />
                </div>
              )}
            />
            <form.Field
              name="expire"
              children={(field) => (
                <div>
                  <Label>Expire:</Label>
                  <Input
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                  />
                  <FieldError errors={field.state.meta.errors} />
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
      <Dialog open={resDialog} onOpenChange={setResDialog}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>API Token:</DialogTitle>
            <DialogDescription>
              <ScrollArea className="break-all">{data?.token}</ScrollArea>
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              type="button"
              onClick={() => {
                if (!data?.token) return;
                navigator.clipboard.writeText(data?.token);
                toast.success("Successfully copied token to clipboard");
              }}
            >
              <Copy />
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
