import { getV1NodesTagsOptions } from "@/client/@tanstack/react-query.gen";
import {
  Combobox,
  ComboboxChip,
  ComboboxChips,
  ComboboxChipsInput,
  ComboboxCollection,
  ComboboxContent,
  ComboboxEmpty,
  ComboboxGroup,
  ComboboxItem,
  ComboboxLabel,
  ComboboxList,
  ComboboxValue,
  useComboboxAnchor,
} from "@/components/ui/combobox";
import { cn } from "@/lib/utils";
import { useQuery } from "@tanstack/react-query";
import { Plus } from "lucide-react";
import { useMemo, useState, type ClipboardEvent } from "react";

type TagGroup = { value: string; items: string[] };

// newTag names the group holding the text being typed when it isn't a tag yet; it has no heading
const newTag = "\u0000new";

// groupTags groups a namespaced tag under its namespace, grendel:brand=dell under grendel, and any other key=value tag under its key, rack=f01 under rack. Key only tags like gpu go last under an empty key. Groups and the tags in them are sorted
export function groupTags(tags: string[]): TagGroup[] {
  const byKey = new Map<string, string[]>();
  for (const tag of tags) {
    const colon = tag.indexOf(":");
    const eq = tag.indexOf("=");
    // a colon after the = is part of the value, as in url=http://host, not a namespace
    const namespaced = colon > 0 && (eq < 0 || colon < eq);
    const key = namespaced ? tag.slice(0, colon) : eq > 0 ? tag.slice(0, eq) : "";
    if (!byKey.has(key)) byKey.set(key, []);
    byKey.get(key)!.push(tag);
  }
  const keys = [...byKey.keys()].filter(Boolean).sort();
  if (byKey.has("")) keys.push("");
  return keys.map((key) => ({ value: key, items: byKey.get(key)!.sort() }));
}

// a pasted list of tags may be separated by commas or whitespace
const pasteSeparator = /[\s,]+/;

// TagCombobox edits a list of tags as chips, suggesting the tags already on nodes with how many nodes have each. Text that isn't an existing tag can still be added, since a tag is created by putting it on a node and a filter can use operators like !gpu or rack=
export function TagCombobox({
  id,
  value,
  onValueChange,
  placeholder = "Add a tag",
  label,
  className,
}: {
  id?: string;
  value: string[];
  onValueChange: (tags: string[]) => void;
  placeholder?: string;
  // label names the input for screen readers, which don't read a placeholder as one; it defaults to the placeholder
  label?: string;
  className?: string;
}) {
  const anchor = useComboboxAnchor();
  const [query, setQuery] = useState("");
  const { data } = useQuery({ ...getV1NodesTagsOptions(), staleTime: 60_000 });

  const counts = useMemo(
    () => new Map((data ?? []).map((t) => [t.tag ?? "", t.nodes ?? 0])),
    [data],
  );
  const typed = query.trim();
  // the known tags, any selected ones the server doesn't know yet, and what is being typed when it is new, grouped by key
  const groups = useMemo(() => {
    const known = [...counts.keys()].filter(Boolean);
    const all = [...known, ...value.filter((t) => !counts.has(t))];
    const grouped = groupTags(all);
    if (typed && !all.includes(typed)) grouped.push({ value: newTag, items: [typed] });
    return grouped;
  }, [counts, value, typed]);

  function set(tags: string[]) {
    onValueChange([...new Set(tags.map((t) => t.trim()).filter(Boolean))]);
    setQuery("");
  }

  function onPaste(e: ClipboardEvent<HTMLInputElement>) {
    const text = e.clipboardData.getData("text");
    if (!pasteSeparator.test(text.trim())) return;
    e.preventDefault();
    set([...value, ...text.split(pasteSeparator)]);
  }

  return (
    <Combobox
      items={groups}
      multiple
      autoHighlight
      value={value}
      onValueChange={set}
      inputValue={query}
      onInputValueChange={setQuery}
    >
      <ComboboxChips ref={anchor} className={cn("w-full", className)}>
        <ComboboxValue>
          {value.map((tag) => (
            <ComboboxChip key={tag} aria-label={tag}>
              {tag}
            </ComboboxChip>
          ))}
        </ComboboxValue>
        <ComboboxChipsInput
          id={id}
          placeholder={value.length ? undefined : placeholder}
          aria-label={label ?? placeholder}
          onPaste={onPaste}
        />
      </ComboboxChips>
      <ComboboxContent anchor={anchor}>
        <ComboboxEmpty>No tags yet. Type one to add it.</ComboboxEmpty>
        <ComboboxList>
          {(group: TagGroup) => (
            <ComboboxGroup key={group.value} items={group.items} className="not-first:mt-1">
              {group.value !== newTag && <ComboboxLabel>{group.value || "Other"}</ComboboxLabel>}
              <ComboboxCollection>
                {(tag: string) =>
                  group.value === newTag ? (
                    <ComboboxItem key={tag} value={tag}>
                      <Plus className="text-muted-foreground" />
                      <span className="truncate">Add “{tag}”</span>
                    </ComboboxItem>
                  ) : (
                    <ComboboxItem key={tag} value={tag} className="pl-4">
                      <span className="truncate">{tag}</span>
                      <span className="ml-auto text-xs text-muted-foreground tabular-nums">
                        {counts.get(tag)}
                      </span>
                    </ComboboxItem>
                  )
                }
              </ComboboxCollection>
            </ComboboxGroup>
          )}
        </ComboboxList>
      </ComboboxContent>
    </Combobox>
  );
}
