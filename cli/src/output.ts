import type { GlobalOptions } from "./types.js";

export function output(data: unknown, options: GlobalOptions, human?: string): void {
  if (options.json) {
    process.stdout.write(`${JSON.stringify(data, null, 2)}\n`);
    return;
  }
  if (human !== undefined) {
    process.stdout.write(`${human}\n`);
    return;
  }
  if (typeof data === "string") {
    process.stdout.write(`${data}\n`);
    return;
  }
  process.stdout.write(`${JSON.stringify(data, null, 2)}\n`);
}

export function jsonLine(data: unknown): void {
  process.stdout.write(`${JSON.stringify(data)}\n`);
}

export function table(rows: Array<Record<string, unknown>>, columns: string[], options: GlobalOptions): void {
  if (options.json) {
    output(rows, options);
    return;
  }
  if (rows.length === 0) {
    process.stdout.write("No records.\n");
    return;
  }
  const widths = columns.map((column) => {
    const values = rows.map((row) => stringify(row[column]));
    return Math.max(column.length, ...values.map((value) => value.length));
  });
  process.stdout.write(`${columns.map((column, index) => column.padEnd(widths[index])).join("  ")}\n`);
  process.stdout.write(`${widths.map((width) => "-".repeat(width)).join("  ")}\n`);
  for (const row of rows) {
    process.stdout.write(`${columns.map((column, index) => stringify(row[column]).padEnd(widths[index])).join("  ")}\n`);
  }
}

function stringify(value: unknown): string {
  if (value === null || value === undefined) {
    return "";
  }
  if (typeof value === "boolean") {
    return value ? "yes" : "no";
  }
  return String(value);
}
