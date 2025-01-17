import { Card, CardContent, CardHeader, CardTitle } from "./ui/card";
import { Skeleton } from "./ui/skeleton";

export function Loading() {
    return (
        <div className="mt-16 flex flex-col gap-6">
            <Card className="w-full">
                <CardHeader>
                    <CardTitle>
                        <Skeleton className="h-6 w-32" />
                    </CardTitle>
                </CardHeader>
                <CardContent className="grid grid-cols-2 gap-4">
                    <div>
                        <Skeleton className="h-6 w-full" />
                    </div>
                    <div>
                        <Skeleton className="h-6 w-full" />
                    </div>
                    <div>
                        <Skeleton className="h-6 w-full" />
                    </div>
                    <div>
                        <Skeleton className="h-6 w-full" />
                    </div>
                    <div>
                        <Skeleton className="h-6 w-full" />
                    </div>
                    <div>
                        <Skeleton className="h-6 w-full" />
                    </div>
                </CardContent>
            </Card>
            <Card className="w-full">
                <CardHeader>
                    <CardTitle>
                        <Skeleton className="h-6 w-32" />
                    </CardTitle>
                </CardHeader>
                <CardContent className="grid grid-cols-2 gap-4">
                    <div>
                        <Skeleton className="h-6 w-full" />
                    </div>
                    <div>
                        <Skeleton className="h-6 w-full" />
                    </div>
                    <div>
                        <Skeleton className="h-6 w-full" />
                    </div>
                    <div>
                        <Skeleton className="h-6 w-full" />
                    </div>
                    <div>
                        <Skeleton className="h-6 w-full" />
                    </div>
                    <div>
                        <Skeleton className="h-6 w-full" />
                    </div>
                </CardContent>
            </Card>
        </div>
    );
}
