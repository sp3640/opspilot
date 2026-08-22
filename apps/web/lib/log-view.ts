export type LogTimeRangeOption = {
  label: string;
  /** undefined means "all available" - no sinceSeconds bound sent to the API. */
  sinceSeconds: number | undefined;
};

export const LOG_TIME_RANGE_OPTIONS: readonly LogTimeRangeOption[] = [
  { label: "Last 5 minutes", sinceSeconds: 5 * 60 },
  { label: "Last 15 minutes", sinceSeconds: 15 * 60 },
  { label: "Last hour", sinceSeconds: 60 * 60 },
  { label: "Last 6 hours", sinceSeconds: 6 * 60 * 60 },
  { label: "Last 24 hours", sinceSeconds: 24 * 60 * 60 },
  { label: "All available", sinceSeconds: undefined },
];

export const LOG_TAIL_LINE_OPTIONS: readonly number[] = [100, 500, 1000, 5000];

const TIMESTAMP_PREFIX_PATTERN = /^(\d{4}-\d{2}-\d{2}T[0-9:.]+Z)\s(.*)$/;

/**
 * Splits a raw log blob into lines, filtering out the trailing empty line a
 * trailing newline produces so line counts and search results are exact.
 */
export function splitLogLines(logText: string): string[] {
  if (!logText) return [];
  const lines = logText.split("\n");
  if (lines.length > 0 && lines[lines.length - 1] === "") {
    lines.pop();
  }
  return lines;
}

/** Case-insensitive substring filter. An empty query returns every line. */
export function filterLogLines(lines: string[], query: string): string[] {
  const trimmed = query.trim().toLowerCase();
  if (!trimmed) return lines;
  return lines.filter((line) => line.toLowerCase().includes(trimmed));
}

/**
 * Kubernetes' `timestamps=true` option prefixes each line with an RFC3339
 * timestamp. This splits that prefix from the message so the UI can toggle
 * displaying it without re-fetching, and so callers can sort/merge lines
 * from multiple pods chronologically.
 */
export function parseLogLine(line: string): { timestamp: string | null; message: string } {
  const match = TIMESTAMP_PREFIX_PATTERN.exec(line);
  if (!match) return { timestamp: null, message: line };
  return { timestamp: match[1] ?? null, message: match[2] ?? "" };
}

export function formatLogLine(line: string, showTimestamps: boolean): string {
  const { timestamp, message } = parseLogLine(line);
  if (!timestamp) return line;
  return showTimestamps ? line : message;
}
