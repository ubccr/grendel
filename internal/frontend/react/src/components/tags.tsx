import { Badge } from "./ui/badge";

type Props = {
    tags?: string[];
};
export default function TagsList({ tags }: Props) {
    return (
        <div className="*:my-auto *:h-6 flex gap-1 overflow-y-scroll">
            {tags?.sort().map((tag, i) => (
                <Badge key={i} variant="outline" className="text-nowrap rounded-md">
                    {tag}
                </Badge>
            ))}
        </div>
    );
}
