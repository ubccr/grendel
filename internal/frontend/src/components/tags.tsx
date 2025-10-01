import { Badge } from "./ui/badge";

type Props = {
  tags?: string[];
};
export default function TagsList({ tags }: Props) {
  return (
    <div className="flex gap-1 overflow-y-scroll *:my-auto *:h-6">
      {tags?.sort().map((tag, i) => (
        <Badge key={i} variant="secondary" className="rounded-md text-nowrap">
          {tag}
        </Badge>
      ))}
    </div>
  );
}
