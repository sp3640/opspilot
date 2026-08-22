import { describe, expect, it } from "vitest";

import { filterLogLines, formatLogLine, parseLogLine, splitLogLines } from "./log-view";

describe("splitLogLines", () => {
  it("returns an empty array for empty input", () => {
    expect(splitLogLines("")).toEqual([]);
  });

  it("splits on newlines and drops a trailing empty line", () => {
    expect(splitLogLines("line one\nline two\n")).toEqual(["line one", "line two"]);
  });

  it("keeps a final line that has no trailing newline", () => {
    expect(splitLogLines("line one\nline two")).toEqual(["line one", "line two"]);
  });
});

describe("filterLogLines", () => {
  const lines = ["INFO starting server", "ERROR connection refused", "INFO ready"];

  it("returns all lines when the query is empty", () => {
    expect(filterLogLines(lines, "")).toEqual(lines);
    expect(filterLogLines(lines, "   ")).toEqual(lines);
  });

  it("filters case-insensitively", () => {
    expect(filterLogLines(lines, "error")).toEqual(["ERROR connection refused"]);
  });

  it("returns an empty array when nothing matches", () => {
    expect(filterLogLines(lines, "does-not-exist")).toEqual([]);
  });
});

describe("parseLogLine", () => {
  it("splits an RFC3339 timestamp prefix from the message", () => {
    const result = parseLogLine("2026-08-22T10:00:00.123456789Z starting up");
    expect(result.timestamp).toBe("2026-08-22T10:00:00.123456789Z");
    expect(result.message).toBe("starting up");
  });

  it("returns null timestamp when the line has no timestamp prefix", () => {
    const result = parseLogLine("starting up");
    expect(result.timestamp).toBeNull();
    expect(result.message).toBe("starting up");
  });
});

describe("formatLogLine", () => {
  const line = "2026-08-22T10:00:00Z starting up";

  it("strips the timestamp when showTimestamps is false", () => {
    expect(formatLogLine(line, false)).toBe("starting up");
  });

  it("keeps the full line when showTimestamps is true", () => {
    expect(formatLogLine(line, true)).toBe(line);
  });

  it("leaves untimestamped lines untouched either way", () => {
    expect(formatLogLine("no timestamp here", false)).toBe("no timestamp here");
    expect(formatLogLine("no timestamp here", true)).toBe("no timestamp here");
  });
});
