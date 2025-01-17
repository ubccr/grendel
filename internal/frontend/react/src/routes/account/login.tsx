import { Button } from "@/components/ui/button";
import { Card, CardContent, CardFooter, CardHeader, CardTitle } from "@/components/ui/card";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { useForm } from "@tanstack/react-form";
import { createFileRoute } from "@tanstack/react-router";
import { InputOTP, InputOTPGroup, InputOTPSeparator, InputOTPSlot } from "@/components/ui/input-otp";

export const Route = createFileRoute("/account/login")({
    component: RouteComponent,
});

function RouteComponent() {
    const form = useForm({
        defaultValues: {
            username: "",
            password: "",
            otp: "",
        },
    });
    return (
        <div className="flex justify-center">
            <Card className="w-80">
                <CardHeader>
                    <CardTitle>Login:</CardTitle>
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
                            name="otp"
                            children={(field) => (
                                <div>
                                    <Label>TOTP:</Label>
                                    <InputOTP
                                        maxLength={6}
                                        value={field.state.value}
                                        onBlur={field.handleBlur}
                                        onChange={(e) => field.handleChange(e)}>
                                        <InputOTPGroup>
                                            <InputOTPSlot index={0} />
                                            <InputOTPSlot index={1} />
                                            <InputOTPSlot index={2} />
                                        </InputOTPGroup>
                                        <InputOTPSeparator />
                                        <InputOTPGroup>
                                            <InputOTPSlot index={3} />
                                            <InputOTPSlot index={4} />
                                            <InputOTPSlot index={5} />
                                        </InputOTPGroup>
                                    </InputOTP>
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
