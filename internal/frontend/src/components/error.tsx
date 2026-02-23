import { ErrorComponentProps } from "@tanstack/react-router";
import { Button } from "./ui/button";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "./ui/card";

export function DefaultErrorComponent(props: ErrorComponentProps) {
  return (
    <Card className="text-center">
      <CardHeader>
        <CardTitle className="text-3xl">At least it's not a blue screen...</CardTitle>
        <CardDescription>
          Unexpected response from the server. Here's some debug information:
        </CardDescription>
      </CardHeader>
      <CardContent>
        <pre className="rounded-xl border bg-secondary p-4 text-start text-wrap text-muted-foreground shadow-sm">
          {JSON.stringify(props.error, null, 4)}
        </pre>
      </CardContent>
      <CardFooter>
        <Button onClick={props.reset}>Try Again</Button>
      </CardFooter>
    </Card>
  );
}
