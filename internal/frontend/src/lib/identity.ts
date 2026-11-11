export type Location = {
  datacenter?: string;
  room?: string;
  row?: string;
  rack?: string;
  rack_unit?: number | null;
};

// formatLocation shows rack and unit first, since that is what finds the machine, with the wider place after it. Matches the CLI
export function formatLocation(loc?: Location | null) {
  if (!loc) return "";
  const head = [
    loc.rack,
    loc.rack_unit != null ? `U${loc.rack_unit}` : undefined,
    loc.row ? `row ${loc.row}` : undefined,
  ].filter(Boolean);
  const place = [loc.datacenter, loc.room].filter(Boolean);
  const out = head.join(" ");
  if (place.length === 0) return out;
  if (!out) return place.join("/");
  return `${out} (${place.join("/")})`;
}

// parseCsv splits CSV text into rows, handling quoted fields with commas, doubled quotes and newlines
export function parseCsv(text: string): string[][] {
  const rows: string[][] = [];
  let row: string[] = [];
  let field = "";
  let quoted = false;
  text = text.replace(/^\ufeff/, "");

  for (let i = 0; i < text.length; i++) {
    const c = text[i];
    if (quoted) {
      if (c === '"' && text[i + 1] === '"') {
        field += '"';
        i++;
      } else if (c === '"') {
        quoted = false;
      } else {
        field += c;
      }
      continue;
    }
    if (c === '"' && field.trim() === "") {
      quoted = true;
      field = "";
    } else if (c === ",") {
      row.push(field);
      field = "";
    } else if (c === "\n" || c === "\r") {
      if (c === "\r" && text[i + 1] === "\n") i++;
      row.push(field);
      rows.push(row);
      row = [];
      field = "";
    } else {
      field += c;
    }
  }
  if (field !== "" || row.length > 0) {
    row.push(field);
    rows.push(row);
  }
  return rows;
}

const serialHeaders = ["serial", "serialnumber", "servicetag", "sn"];
const tagHeaders = ["assettag", "tag", "asset"];

export const normalizeHeader = (h: string) =>
  h
    .trim()
    .toLowerCase()
    .replace(/[\s_\-\ufeff]/g, "");

export type AssetCsv = {
  // header is the first row when it named the columns, so the columns can be picked by name
  header?: string[];
  serialIndex: number;
  tagIndex: number;
  bySerial: Record<string, string>;
  error?: string;
};

// guessColumns finds the serial and asset tag columns by header, the same names the CLI accepts. Without a header row the file is read as serial then asset tag
export function guessColumns(rows: string[][]) {
  const first = (rows[0] ?? []).map(normalizeHeader);
  const serialIndex = first.findIndex((h) => serialHeaders.includes(h));
  const tagIndex = first.findIndex((h) => tagHeaders.includes(h));
  const header = serialIndex >= 0 || tagIndex >= 0;
  if (!header) return { header: false, serialIndex: 0, tagIndex: 1 };
  return { header: true, serialIndex, tagIndex };
}

// assetCsv reads the serial to asset tag pairs from parsed rows. A serial listed twice is an error even with the same tag, as in the CLI
export function assetCsv(
  rows: string[][],
  hasHeader: boolean,
  serialIndex: number,
  tagIndex: number,
): AssetCsv {
  const header = hasHeader ? rows[0] : undefined;
  const out: AssetCsv = { header, serialIndex, tagIndex, bySerial: {} };
  if (rows.length === 0) return { ...out, error: "The file is empty." };
  if (serialIndex < 0 || tagIndex < 0) {
    return { ...out, error: "Pick the serial and asset tag columns." };
  }
  if (serialIndex === tagIndex) {
    return { ...out, error: "The serial and asset tag columns must differ." };
  }

  const start = hasHeader ? 1 : 0;
  for (let i = start; i < rows.length; i++) {
    const row = rows[i];
    if (row.length === 1 && row[0].trim() === "") continue;
    const line = i + 1;
    if (row.length <= Math.max(serialIndex, tagIndex)) {
      return { ...out, error: `Line ${line} has ${row.length} column(s).` };
    }
    const serial = row[serialIndex].trim();
    const tag = row[tagIndex].trim();
    if (!serial) continue;
    if (serial in out.bySerial) {
      return { ...out, error: `Line ${line}: serial ${serial} is listed twice.` };
    }
    out.bySerial[serial] = tag;
  }
  if (Object.keys(out.bySerial).length === 0) {
    return { ...out, error: "No rows have a serial number." };
  }
  return out;
}
