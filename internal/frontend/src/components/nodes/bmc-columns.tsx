import { DataTableColumnHeader } from "@/components/data-table/header";
import { Badge } from "@/components/ui/badge";
import { severityColor } from "@/lib/utils";
import { ColumnDef, Row } from "@tanstack/react-table";

// NodeRow is what every row of a per node bmc table carries about its node, whether or not the node returned any items
export type NodeRow = {
  node: string;
  success: boolean;
  node_message: string;
};

type NodeResult<T> = {
  node?: string;
  success?: boolean;
  message?: string;
  items?: Array<T>;
};

// rowsPerNode flattens a bmc response into one row per item, and gives a node with no items, including one that could not be reached, a single row of its own. That is how an empty result is told apart from a bmc that never answered, the same as the cli
export function rowsPerNode<T extends object>(
  results: Array<NodeResult<T>>,
): Array<NodeRow & Partial<T>> {
  return results.flatMap((r) => {
    const base: NodeRow = {
      node: r.node ?? "",
      success: r.success ?? false,
      node_message: r.message ?? "",
    };
    const items = r.items ?? [];
    if (items.length === 0) return [base as NodeRow & Partial<T>];
    return items.map((item) => ({ ...base, ...item }));
  });
}

// countWith counts the rows in a group that hold an item, so a node's placeholder row is not counted as one
export function countWith<T>(key: keyof T) {
  return (_columnId: string, leafRows: Array<Row<T>>) =>
    leafRows.filter((r) => r.original[key]).length;
}

function StateBadge({ success }: { success: boolean }) {
  return (
    <Badge
      variant="secondary"
      className={`rounded-sm ${severityColor(success ? "success" : "error")}`}
    >
      {success ? "success" : "failed"}
    </Badge>
  );
}

// nodeColumns are the state and message of the node's own request, shown on every row and once on the node's group row
export function nodeColumns<T extends NodeRow>(): Array<ColumnDef<T>> {
  return [
    {
      id: "node_state",
      accessorFn: (row) => (row.success ? "success" : "failed"),
      header: ({ column }) => <DataTableColumnHeader column={column} title="State" />,
      aggregationFn: "unique",
      cell: ({ row }) => <StateBadge success={row.original.success} />,
      aggregatedCell: ({ row }) => (
        <StateBadge success={row.subRows[0]?.original.success ?? false} />
      ),
    },
    {
      accessorKey: "node_message",
      header: ({ column }) => <DataTableColumnHeader column={column} title="Message" />,
      aggregationFn: "unique",
      aggregatedCell: ({ row }) => <span>{row.subRows[0]?.original.node_message}</span>,
    },
  ];
}
