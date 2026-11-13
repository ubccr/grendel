import type { DataTableTree, RowTone } from "@/components/data-table/data-table";
import type { DataTableColumnDef, DataTableRow } from "@/components/data-table/features";
import { DataTableColumnHeader } from "@/components/data-table/header";
import { Link } from "@tanstack/react-router";

// NodeRow is what every row of a per node bmc table carries about its node
export type NodeRow = {
  node: string;
  success: boolean;
  node_message: string;
};

// NodeTreeRow is a row of a per node bmc table: a node, with what its BMC returned as sub rows. Each item row carries its node's fields too, so a filter on the node's columns still finds the node's items
export type NodeTreeRow<T> = NodeRow & Partial<T> & { items?: NodeTreeRow<T>[] };

type NodeResult<T> = {
  node?: string;
  success?: boolean;
  message?: string;
  items?: Array<T>;
};

// nodeTree turns a bmc response into a node row per node with its items under it. A node with no items, including one whose BMC could not be reached, is a node row with nothing under it, which is how an empty result is told apart from a BMC that never answered, the same as the cli
export function nodeTree<T extends object>(results: Array<NodeResult<T>>): NodeTreeRow<T>[] {
  return results.map((r) => {
    const node: NodeRow = {
      node: r.node ?? "",
      success: r.success ?? false,
      node_message: r.message ?? "",
    };
    const items = (r.items ?? []).map((item): NodeTreeRow<T> => ({ ...node, ...item }));
    // a node row has none of its items' own fields, which an empty Partial<T> says without a cast
    const none: Partial<T> = {};
    return { ...none, ...node, items };
  });
}

// nodeTreeTable is the DataTable tree for a per node bmc table, with the columns from nodeColumns describing each node
export const nodeTreeTable: DataTableTree<NodeTreeRow<object>> = {
  getSubRows: (row) => row.items,
  groupColumns: ["node", "node_state", "node_message"],
};

// nodeRowId names a node's row by the node and an item's row by its node and its own id, so a selection follows them across a refresh
export function nodeRowId<T extends { id?: unknown }>(row: NodeTreeRow<T>) {
  return row.id === undefined || row.id === "" ? row.node : `${row.node}/${String(row.id)}`;
}

// nodeRowTone marks a node's row by whether its BMC answered, so the rail beside a node that failed shows red
export function nodeRowTone<T>(row: DataTableRow<NodeTreeRow<T>>): RowTone | undefined {
  if (row.depth > 0) return undefined;
  return row.original.success ? "ok" : "error";
}

// nodeColumns are the node's name, linked to the node, and the state and message of the request to its BMC. They lay out the node rows of a nodeTreeTable
export function nodeColumns<T>(): Array<DataTableColumnDef<NodeTreeRow<T>, any>> {
  return [
    {
      accessorKey: "node",
      header: ({ column }) => <DataTableColumnHeader column={column} title="Node" />,
      cell: ({ row }) => (
        <Link
          to="/nodes/$node/node"
          params={{ node: row.original.node }}
          className="hover:underline"
        >
          {row.original.node}
        </Link>
      ),
    },
    {
      id: "node_state",
      accessorFn: (row) => (row.success ? "success" : "failed"),
      header: ({ column }) => <DataTableColumnHeader column={column} title="State" />,
    },
    {
      accessorKey: "node_message",
      header: ({ column }) => <DataTableColumnHeader column={column} title="Message" />,
    },
  ];
}
