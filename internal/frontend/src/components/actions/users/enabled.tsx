import { patchV1UsersUsernamesEnableMutation } from "@/client/@tanstack/react-query.gen";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { useMutation } from "@tanstack/react-query";
import { LoaderCircle } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";

export default function UsersEnabledAction({ users }: { users: string }) {
  const [userEnabled, setUserEnabled] = useState("");
  const { mutate, isPending } = useMutation(patchV1UsersUsernamesEnableMutation());
  return (
    <Card>
      <CardHeader>
        <CardTitle>Change Enabled flag</CardTitle>
      </CardHeader>
      <CardContent className="grid grid-cols-1 gap-2">
        <Select onValueChange={(e) => setUserEnabled(e)}>
          <SelectTrigger className="w-[180px]">
            <SelectValue placeholder="Enabled" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="true">True</SelectItem>
            <SelectItem value="false">False</SelectItem>
          </SelectContent>
        </Select>
      </CardContent>
      <CardFooter>
        <Button
          onClick={() =>
            mutate(
              {
                path: { usernames: users },
                body: { enabled: userEnabled === "true" },
              },
              {
                onSuccess: (data) => {
                  toast.success(data?.title, {
                    description: data?.detail,
                  });
                },
                onError: (e) =>
                  toast.error(e.title, {
                    description: e.detail,
                  }),
              },
            )
          }
        >
          {isPending ? <LoaderCircle className="animate-spin" /> : <span>Submit</span>}
        </Button>
      </CardFooter>
    </Card>
  );
}
