import { Error } from "@/components/error";
import { Loading } from "@/components/loading";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table";
import { toast } from "sonner";
import { useHostListSuspense } from "@/openapi/queries/suspense";
import { createFileRoute, Link } from "@tanstack/react-router";
import { Suspense, useState } from "react";
import { ErrorBoundary } from "react-error-boundary";
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuLabel,
    DropdownMenuRadioGroup,
    DropdownMenuRadioItem,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
import { Button } from "@/components/ui/button";

export const Route = createFileRoute("/floorplan")({
    component: Floorplan,
});

function TableComponent() {
    const { data, isSuccess } = useHostListSuspense();
    const rows = Array.from("fghijklmnopqrstuv");
    const cols: string[] = [];
    const [view, setView] = useState("rackName");

    if (isSuccess) {
        for (let x = 28; x >= 5; x--) {
            cols.push(`${x}`);
        }
    }

    const populated: Set<string> = new Set([]);
    const size: Map<string, number> = new Map();

    data?.forEach((element) => {
        const parts = element.name.split("-");
        if (parts.length < 2) {
            return;
        }
        populated.add(parts[1]);

        const currentSize = size.get(parts[1]) ?? 0;
        size.set(parts[1], currentSize + 1);
    });

    return (
        <div className="flex justify-center">
            <Table>
                <TableHeader className="*:text-center">
                    <TableRow>
                        <TableHead className="w-12 border">
                            <DropdownMenu>
                                <DropdownMenuTrigger asChild>
                                    <Button variant="outline" size="sm">
                                        View
                                    </Button>
                                </DropdownMenuTrigger>
                                <DropdownMenuContent className="w-56">
                                    <DropdownMenuLabel>Display:</DropdownMenuLabel>
                                    <DropdownMenuSeparator />
                                    <DropdownMenuRadioGroup value={view} onValueChange={setView}>
                                        <DropdownMenuRadioItem value="rackName">Rack Name</DropdownMenuRadioItem>
                                        <DropdownMenuRadioItem value="nodeCount">Node Count</DropdownMenuRadioItem>
                                    </DropdownMenuRadioGroup>
                                </DropdownMenuContent>
                            </DropdownMenu>
                        </TableHead>
                        {cols.map((col, i) => (
                            <TableHead key={i} className="border text-center">
                                {col}
                            </TableHead>
                        ))}
                    </TableRow>
                </TableHeader>
                <TableBody>
                    {rows.map((row, i) => (
                        <TableRow key={i} className="*:text-center">
                            <TableHead className="border">{row}</TableHead>
                            {cols.map((col, i) => {
                                const rack = row + col;
                                return (
                                    <TableCell key={i} className="border p-0">
                                        {populated.has(rack) && (
                                            <Link to={`/rack/${rack}`} className="hover:font-bold">
                                                {view === "rackName" && rack}
                                                {view === "nodeCount" && size.get(rack)}
                                            </Link>
                                        )}
                                    </TableCell>
                                );
                            })}
                        </TableRow>
                    ))}
                </TableBody>
            </Table>
        </div>
    );
}

function Floorplan() {
    return (
        <div className="p-4">
            <Suspense fallback={<Loading />}>
                <ErrorBoundary
                    fallback={<Error />}
                    onError={(error) =>
                        toast.error("Error loading response", {
                            description: error.message,
                        })
                    }>
                    <TableComponent />
                </ErrorBoundary>
            </Suspense>
        </div>
    );
}
