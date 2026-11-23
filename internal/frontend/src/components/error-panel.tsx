import {
  Empty,
  EmptyDescription,
  EmptyHeader,
  EmptyMedia,
  EmptyTitle,
} from "@/components/ui/empty";
import { ServerOff, type LucideIcon } from "lucide-react";

export function ErrorPanel({
  title,
  message,
  icon: Icon = ServerOff,
}: {
  title: string;
  message?: string;
  icon?: LucideIcon;
}) {
  return (
    <Empty>
      <EmptyHeader>
        <EmptyMedia variant="icon">
          <Icon />
        </EmptyMedia>
        <EmptyTitle>{title}</EmptyTitle>
        <EmptyDescription className="break-words">
          {message || "The server gave no reason."}
        </EmptyDescription>
      </EmptyHeader>
    </Empty>
  );
}
