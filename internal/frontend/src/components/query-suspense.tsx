import { ErrorBoundary, FallbackProps } from "react-error-boundary";
import {
  Card,
  CardContent,
  CardDescription,
  CardFooter,
  CardHeader,
  CardTitle,
} from "./ui/card";
import { Button } from "./ui/button";
import { QueryErrorResetBoundary } from "@tanstack/react-query";
import React from "react";

export function QuerySuspense({ children }: { children: React.ReactNode }) {
  return (
    <QueryErrorResetBoundary>
      {({ reset }) => (
        <ErrorBoundary
          onReset={reset}
          fallbackRender={({ error, resetErrorBoundary }: FallbackProps) => (
            <div className="flex justify-center align-middle">
              <Card className="text-center">
                <CardHeader>
                  <CardTitle>Oops, something has gone wrong!</CardTitle>
                  <CardDescription>
                    Seems like you've ran into the mysterious "runtime error"{" "}
                    <br /> ooo, spooky
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <h1 className="text-md text-muted-foreground">
                    Here's some debug information you can forward to your local
                    Wizard:
                  </h1>
                  <pre className="text-left text-sm font-mono text-muted-foreground">
                    {JSON.stringify(error, null, 4)}
                  </pre>
                </CardContent>
                <CardFooter>
                  <Button type="button" size="sm" onClick={resetErrorBoundary}>
                    Retry
                  </Button>
                </CardFooter>
              </Card>
            </div>
          )}
        >
          {children}
        </ErrorBoundary>
      )}
    </QueryErrorResetBoundary>
  );
}
