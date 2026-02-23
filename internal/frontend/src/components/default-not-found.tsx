import { NotFoundRouteProps } from "@tanstack/react-router";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "./ui/card";

export function DefaultNotFound(props: NotFoundRouteProps) {
  return (
    <Card className="text-center">
      <CardHeader>
        <CardTitle className="text-3xl">404 Not Found</CardTitle>
        <CardDescription>The requested resource could not be found.</CardDescription>
      </CardHeader>
      <CardContent>
        <pre className="rounded-xl border bg-secondary p-4 text-start text-wrap text-muted-foreground shadow-sm">
          {JSON.stringify(props.data, null, 4)}
        </pre>
      </CardContent>
    </Card>
  );
}
