"use client";

import { Search, X } from "lucide-react";

import { Button } from "@/components/ui/button";

type IncidentSearchProps = {
  value: string;
  onChange: (value: string) => void;
};

export function IncidentSearch({ value, onChange }: IncidentSearchProps) {
  return (
    <div className="relative flex-1">
      <Search
        aria-hidden="true"
        className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2"
        style={{ color: "var(--muted-foreground)" }}
      />
      <input
        type="text"
        placeholder="Search incidents by title or description..."
        value={value}
        onChange={(e) => onChange(e.target.value)}
        className="w-full rounded-xl border bg-transparent py-2 pl-10 pr-10 text-sm placeholder-shown:text-transparent focus:outline-none"
        style={{ borderColor: "var(--border)", color: "var(--foreground)" }}
      />
      {value && (
        <Button
          type="button"
          variant="ghost"
          onClick={() => onChange("")}
          className="absolute right-1 top-1/2 h-6 w-6 -translate-y-1/2 rounded-lg p-0"
          aria-label="Clear search"
        >
          <X aria-hidden="true" className="h-4 w-4" />
        </Button>
      )}
    </div>
  );
}
