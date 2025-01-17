import { useHostProvision, useHostUnprovision } from "@/openapi/queries";
import { LoaderCircle, Zap, ZapOff } from "lucide-react";
import { Button } from "../ui/button";
import { useQueryClient } from "@tanstack/react-query";
import { toast } from "sonner";
import { useEffect, useState } from "react";

type Props = {
    name?: string;
    provision?: boolean;
};

export default function ProvisionIcon({ provision, name }: Props) {
    const mutate_provision = useHostProvision();
    const mutate_unprovision = useHostUnprovision();
    const queryClient = useQueryClient();

    const [ping, setPing] = useState(false);

    useEffect(() => {
        if (ping) setTimeout(() => setPing(false), 2000);
    }, [ping]);

    return (
        <>
            {name == undefined || provision == undefined ? (
                <LoaderCircle />
            ) : (
                <Button
                    size="sm"
                    variant="outline"
                    type="button"
                    className={ping ? "animate-pulse" : ""}
                    onClick={() => {
                        setPing(true);
                        if (provision) {
                            mutate_unprovision.mutate(
                                { path: { nodeSet: name } },
                                {
                                    onSuccess: () => {
                                        queryClient.invalidateQueries();
                                    },
                                    onError: (e) =>
                                        toast.error("Failed to set host(s) to unprovision", { description: e.message }),
                                },
                            );
                        } else {
                            mutate_provision.mutate(
                                { path: { nodeSet: name } },
                                {
                                    onSuccess: () => {
                                        queryClient.invalidateQueries();
                                    },
                                    onError: (e) =>
                                        toast.error("Failed to set host(s) to provision", { description: e.message }),
                                },
                            );
                        }
                    }}>
                    {provision && <Zap className="text-green-600" />}
                    {!provision && <ZapOff className="text-red-600" />}
                </Button>
            )}
        </>
    );
}
