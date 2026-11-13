import { Button } from "@/components/ui/button";
import { Popover, PopoverContent, PopoverTrigger } from "@/components/ui/popover";
import { CircleHelp } from "lucide-react";

// TemplateHelp lists what a template can use, worked through on one of the selected nodes so the parts of its name are concrete
export default function TemplateHelp({ example }: { example?: string }) {
  const name = example || "cpn-f01-04";
  const parts = name.split("-");

  const rows: Array<[string, string]> = [
    ["{{ .Name }}", name],
    ...parts.map((p, i): [string, string] => [`{{ index .Parts ${i} }}`, p]),
    ["{{ .Index }} / {{ .Count }}", "position among the selected nodes, from 0, and how many"],
    ['{{ add .Index 1 | printf "%03d" }}', "001, 002, … in name order"],
    ['{{ tag "rack" }}', "value of the node's rack= tag"],
    ['{{ index .Tags "rack" | default "none" }}', "the tag, or none when missing"],
    ['{{ index .Parts 2 | trimLeft "0" }}', "04 becomes 4"],
  ];

  return (
    <Popover>
      <PopoverTrigger
        render={
          <Button variant="ghost" size="sm" className="h-7 gap-1 px-2 text-muted-foreground" />
        }
      >
        <CircleHelp className="size-3.5" />
        Template help
      </PopoverTrigger>
      <PopoverContent align="end" className="w-[26rem] text-sm">
        <div className="grid gap-3">
          <p className="text-muted-foreground">
            Fields are Go templates, rendered for each node. Plain text works too. Shown for{" "}
            <span className="font-mono text-foreground">{name}</span>:
          </p>
          <dl className="grid grid-cols-[auto_1fr] gap-x-3 gap-y-1">
            {rows.map(([t, v]) => (
              <div key={t} className="contents">
                <dt className="font-mono text-xs">{t}</dt>
                <dd className="truncate text-xs text-muted-foreground">{v}</dd>
              </div>
            ))}
          </dl>
          <p className="text-xs text-muted-foreground">
            Functions: split, join, trimPrefix, trimSuffix, trimLeft, replace, upper, lower,
            default, atoi, add, printf, tag. A node whose template fails, such as a name with too
            few parts, is reported and left unchanged.
          </p>
        </div>
      </PopoverContent>
    </Popover>
  );
}
