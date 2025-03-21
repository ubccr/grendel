import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "./ui/card";

export function Error() {
    return (
        <div className="flex justify-center align-middle">
            <Card>
                <CardHeader>
                    <CardTitle>Oops, something has gone wrong!</CardTitle>
                    <CardDescription>
                        Seems like you've ran into the mysterious "runtime error" <br /> ooo, spooky
                    </CardDescription>
                </CardHeader>
                <CardContent></CardContent>
            </Card>
        </div>
    );
}
