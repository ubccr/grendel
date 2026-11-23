import { ToggleGroup, ToggleGroupItem } from "@/components/ui/toggle-group";
import { useId } from "react";

// ArchToggle picks the arch a node with none set is taken to have, which is otherwise only known once it network boots
export function ArchToggle<T extends string>({
  options,
  value,
  onChange,
}: {
  options: readonly T[];
  value: T;
  onChange: (arch: T) => void;
}) {
  const label = useId();

  return (
    <div className="flex items-center gap-2">
      <span id={label} className="text-sm text-muted-foreground">
        Boot as
      </span>
      <ToggleGroup
        aria-labelledby={label}
        variant="outline"
        size="sm"
        spacing={0}
        value={[value]}
        onValueChange={(v) => v[0] && onChange(v[0] as T)}
      >
        {options.map((a) => (
          <ToggleGroupItem key={a} value={a} className="px-2.5 font-mono text-xs">
            {a === "all" ? "both" : a}
          </ToggleGroupItem>
        ))}
      </ToggleGroup>
    </div>
  );
}
