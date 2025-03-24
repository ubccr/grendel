import { createFileRoute } from "@tanstack/react-router";
// import { ColumnDef } from "@tanstack/react-table";
// import { DataTable } from "@/components/data-table/data-table";
// import { DataTableColumnHeader } from "@/components/data-table/header";
// import { Checkbox } from "@/components/ui/checkbox";

export const Route = createFileRoute("/templates/")({
  component: RouteComponent,
});

// const columns: ColumnDef<unknown>[] = [
//     {
//         id: "select",
//         header: ({ table }) => (
//             <Checkbox
//                 checked={table.getIsAllPageRowsSelected() || (table.getIsSomePageRowsSelected() && "indeterminate")}
//                 onCheckedChange={(value) => table.toggleAllPageRowsSelected(!!value)}
//                 aria-label="Select all"
//             />
//         ),
//         cell: ({ row }) => (
//             <Checkbox
//                 checked={row.getIsSelected()}
//                 onCheckedChange={(value) => row.toggleSelected(!!value)}
//                 aria-label="Select row"
//             />
//         ),
//     },
//     {
//         accessorKey: "name",
//         header: ({ column }) => <DataTableColumnHeader column={column} title="Name" />,
//         cell: ({ row }) => {
//             const name = row.original.name;
//             return (
//                 <Link to={`/templates/${name}`} className="hover:underline">
//                     {name}
//                 </Link>
//             );
//         },
//     },
//     {
//         accessorKey: "images",
//         header: ({ column }) => <DataTableColumnHeader column={column} title="Images" />,
//     },
// ];

function RouteComponent() {
  // const { data, isSuccess } = useImageList();
  // const data = [
  //     {
  //         name: "frosty.tmpl",
  //         images: ["frosty", "frosty-compile"],
  //     },
  //     {
  //         name: "flatcar.tmpl",
  //         images: ["flatcar"],
  //     },
  //     {
  //         name: "test.tmpl",
  //         images: [],
  //     },
  // ];
  return (
    <div className="px-6">
      {/* <DataTable columns={columns} data={data} add={"/add/template"} /> */}
    </div>
  );
}
