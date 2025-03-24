import { ParsedLocation, redirect } from "@tanstack/react-router";
import { User } from "./hooks/user-provider";

export default function AuthRedirect({
  context,
  location,
}: {
  context: {
    auth: User;
  };
  location: ParsedLocation;
}) {
  if (!context.auth) {
    throw redirect({
      to: "/account/signin",
      search: {
        redirect: location.href,
      },
    });
  }
}
