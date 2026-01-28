import { Row, Table } from "@tanstack/react-table";
import { Checkbox } from "../ui/checkbox";

type Props<T> = {
  row: Row<T>;
  table: Table<T>;
  lastSelectedID: number;
  setLastSelectedID: React.Dispatch<React.SetStateAction<number>>;
};

/**
 * This component allows for many rows in a datatable to be selected together using a shift click.
 * @param {Row<T>} row - Tanstack table row object
 * @param {Table<T>} table - Tanstack table object
 * @param {number} lastSelectedID - state to store last checked checkbox ID
 * @param {React.Dispatch<React.SetStateAction<number>>} setLastSelectedID
 */
export default function SelectableCheckbox<T>({
  row,
  table,
  lastSelectedID,
  setLastSelectedID,
}: Props<T>) {
  return (
    <Checkbox
      checked={row.getIsSelected()}
      onClick={(e) => {
        if (e.shiftKey) {
          const { rows, rowsById } = table.getPrePaginationRowModel();

          const rangeStart = lastSelectedID > row.index ? row.index : lastSelectedID;
          const rangeEnd = rangeStart === row.index ? lastSelectedID : row.index;
          const rowsToToggle = rows.filter(
            (_row) => rangeStart <= _row.index && rangeEnd >= _row.index,
          );

          const isCellSelected = rowsById[row.id].getIsSelected();
          rowsToToggle.forEach((_row) => _row.toggleSelected(!isCellSelected));
        } else {
          row.toggleSelected();
        }

        setLastSelectedID(row.index);
      }}
      aria-label="Select row"
    />
  );
}
