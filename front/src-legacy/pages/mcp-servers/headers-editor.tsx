/* eslint-disable react-refresh/only-export-components */

import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Plus, Trash2 } from "lucide-react";

export interface HeaderEntry {
  key: string;
  value: string;
}

interface HeadersEditorProps {
  value: HeaderEntry[];
  onChange: (next: HeaderEntry[]) => void;
}

export function HeadersEditor({ value, onChange }: HeadersEditorProps) {
  function update(index: number, patch: Partial<HeaderEntry>) {
    onChange(value.map((row, i) => (i === index ? { ...row, ...patch } : row)));
  }
  function remove(index: number) {
    onChange(value.filter((_, i) => i !== index));
  }
  function add() {
    onChange([...value, { key: "", value: "" }]);
  }

  return (
    <div className="space-y-2">
      {value.length === 0 ? (
        <p className="text-xs text-muted-foreground">No headers configured.</p>
      ) : (
        value.map((row, i) => (
          <div key={i} className="flex items-center gap-2">
            <Input
              placeholder="Header name (e.g. Authorization)"
              value={row.key}
              onChange={(e) => update(i, { key: e.target.value })}
              className="flex-1"
            />
            <Input
              placeholder="Value"
              value={row.value}
              onChange={(e) => update(i, { value: e.target.value })}
              className="flex-1"
            />
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              onClick={() => remove(i)}
              aria-label="Remove header"
            >
              <Trash2 className="h-3 w-3" />
            </Button>
          </div>
        ))
      )}
      <Button type="button" variant="outline" size="sm" onClick={add}>
        <Plus className="mr-1 h-3 w-3" /> Add header
      </Button>
    </div>
  );
}

export function entriesToRecord(entries: HeaderEntry[]): Record<string, string> {
  const out: Record<string, string> = {};
  for (const { key, value } of entries) {
    const k = key.trim();
    if (!k) continue;
    out[k] = value;
  }
  return out;
}

export function recordToEntries(rec: Record<string, string> | undefined): HeaderEntry[] {
  if (!rec) return [];
  return Object.entries(rec).map(([key, value]) => ({ key, value }));
}
