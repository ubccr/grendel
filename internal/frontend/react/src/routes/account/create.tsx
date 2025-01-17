import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useForm } from "@tanstack/react-form";
import { createFileRoute } from "@tanstack/react-router";

export const Route = createFileRoute("/account/create")({
    component: RouteComponent,
});

function RouteComponent() {
    const form = useForm({
        defaultValues: {
            username: "",
            password: "",
            confirmPassword: "",
        },
    });
    return (
        <div className="flex justify-center">
            <Card className="w-80">
                <CardHeader>
                    <CardTitle>Create an Account:</CardTitle>
                    <CardDescription>
                        New to Grendel? Create your account then ask your administrator to enable it
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <form className="grid grid-cols-1 gap-4">
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
                                </div>
                            )}
                        />
                        <form.Field
                            name="password"
                            children={(field) => (
                                <div>
                                    <Label>Password:</Label>
                                    <Input
                                        type="password"
                                        value={field.state.value}
                                        onBlur={field.handleBlur}
                                        onChange={(e) => field.handleChange(e.target.value)}
                                    />
                                </div>
                            )}
                        />
                        <form.Field
                            name="confirmPassword"
                            children={(field) => (
                                <div>
                                    <Label>Confirm Password:</Label>
                                    <Input
                                        type="password"
                                        value={field.state.value}
                                        onBlur={field.handleBlur}
                                        onChange={(e) => field.handleChange(e.target.value)}
                                    />
                                </div>
                            )}
                        />
                    </form>
                </CardContent>
                <CardFooter>
                    <Button>Submit</Button>
                </CardFooter>
            </Card>
        </div>
    );
}
