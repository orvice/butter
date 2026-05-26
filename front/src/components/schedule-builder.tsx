import { useEffect, useMemo, useRef, useState } from "react";
import { Input } from "@/components/ui/input";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";

/**
 * ScheduleBuilder — friendly UI for building a 5-field cron expression.
 *
 * Two modes:
 *  - "simple": pick a frequency (minute / hour / day / week / month) and a time,
 *    component composes the cron expression.
 *  - "advanced": free-form cron expression (or @-shortcuts) — for power users.
 *
 * The component is fully controlled via `value` / `onChange` so it works inside
 * react-hook-form without ref forwarding.
 */

export interface ScheduleBuilderProps {
  value: string;
  onChange: (next: string) => void;
}

type Frequency = "minutes" | "hourly" | "daily" | "weekly" | "monthly";

interface SimpleState {
  frequency: Frequency;
  interval: number;   // for minutes / hourly
  hour: number;       // 0-23 for daily / weekly / monthly
  minute: number;     // 0-59
  weekdays: number[]; // 0-6 (Sun-Sat), weekly
  dayOfMonth: number; // 1-31, monthly
}

const DEFAULT_SIMPLE: SimpleState = {
  frequency: "daily",
  interval: 5,
  hour: 9,
  minute: 0,
  weekdays: [1, 2, 3, 4, 5],
  dayOfMonth: 1,
};

const WEEKDAY_LABELS = ["Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"];

function pad2(n: number) {
  return n.toString().padStart(2, "0");
}

function buildCron(s: SimpleState): string {
  switch (s.frequency) {
    case "minutes":
      return `*/${Math.max(1, Math.min(59, s.interval))} * * * *`;
    case "hourly":
      return `${s.minute} */${Math.max(1, Math.min(23, s.interval))} * * *`;
    case "daily":
      return `${s.minute} ${s.hour} * * *`;
    case "weekly": {
      // Require at least one weekday — otherwise emit an empty string so the
      // form's `min(1)` validation rejects submission instead of silently
      // falling back to a daily schedule.
      if (s.weekdays.length === 0) return "";
      const days = [...s.weekdays].sort((a, b) => a - b).join(",");
      return `${s.minute} ${s.hour} * * ${days}`;
    }
    case "monthly":
      return `${s.minute} ${s.hour} ${Math.max(1, Math.min(31, s.dayOfMonth))} * *`;
  }
}

/**
 * Try to parse a cron expression back into SimpleState. Returns null when the
 * expression cannot be represented in simple mode (so we fall back to advanced).
 */
function parseSimple(expr: string): SimpleState | null {
  const trimmed = expr.trim();
  if (!trimmed || trimmed.startsWith("@")) return null;
  const parts = trimmed.split(/\s+/);
  if (parts.length !== 5) return null;
  const [minF, hourF, domF, monF, dowF] = parts;

  // Only month=* supported by simple mode.
  if (monF !== "*") return null;

  // Every N minutes: */N * * * *
  const minStep = /^\*\/(\d+)$/.exec(minF);
  if (minStep && hourF === "*" && domF === "*" && dowF === "*") {
    return { ...DEFAULT_SIMPLE, frequency: "minutes", interval: parseInt(minStep[1], 10) };
  }

  // Hourly: M */N * * * (or M * * * *)
  const hourStep = /^\*\/(\d+)$/.exec(hourF);
  const minNum = /^(\d+)$/.exec(minF);
  if (minNum && (hourStep || hourF === "*") && domF === "*" && dowF === "*") {
    return {
      ...DEFAULT_SIMPLE,
      frequency: "hourly",
      interval: hourStep ? parseInt(hourStep[1], 10) : 1,
      minute: parseInt(minNum[1], 10),
    };
  }

  const hourNum = /^(\d+)$/.exec(hourF);
  if (!minNum || !hourNum) return null;

  // Daily: M H * * *
  if (domF === "*" && dowF === "*") {
    return {
      ...DEFAULT_SIMPLE,
      frequency: "daily",
      minute: parseInt(minNum[1], 10),
      hour: parseInt(hourNum[1], 10),
    };
  }

  // Weekly: M H * * D[,D...]
  if (domF === "*" && dowF !== "*") {
    const days = dowF.split(",").map((d) => parseInt(d, 10));
    if (days.every((d) => Number.isInteger(d) && d >= 0 && d <= 6)) {
      return {
        ...DEFAULT_SIMPLE,
        frequency: "weekly",
        minute: parseInt(minNum[1], 10),
        hour: parseInt(hourNum[1], 10),
        weekdays: days,
      };
    }
  }

  // Monthly: M H D * *
  if (dowF === "*") {
    const dom = parseInt(domF, 10);
    if (Number.isInteger(dom) && dom >= 1 && dom <= 31) {
      return {
        ...DEFAULT_SIMPLE,
        frequency: "monthly",
        minute: parseInt(minNum[1], 10),
        hour: parseInt(hourNum[1], 10),
        dayOfMonth: dom,
      };
    }
  }

  return null;
}

function describe(s: SimpleState): string {
  const time = `${pad2(s.hour)}:${pad2(s.minute)}`;
  switch (s.frequency) {
    case "minutes":
      return s.interval === 1 ? "Every minute" : `Every ${s.interval} minutes`;
    case "hourly":
      return s.interval === 1
        ? `Every hour at :${pad2(s.minute)}`
        : `Every ${s.interval} hours at :${pad2(s.minute)}`;
    case "daily":
      return `Every day at ${time}`;
    case "weekly": {
      if (s.weekdays.length === 0) return "Pick at least one day";
      const names = [...s.weekdays].sort((a, b) => a - b).map((d) => WEEKDAY_LABELS[d]).join(", ");
      return `Every week on ${names} at ${time}`;
    }
    case "monthly":
      return `Every month on day ${s.dayOfMonth} at ${time}`;
  }
}

export function ScheduleBuilder({ value, onChange }: ScheduleBuilderProps) {
  // Initial mode decided once from incoming value.
  const initialSimple = useMemo(() => parseSimple(value) ?? null, []); // eslint-disable-line react-hooks/exhaustive-deps
  const [mode, setMode] = useState<"simple" | "advanced">(
    value && initialSimple === null ? "advanced" : "simple",
  );
  const [simple, setSimple] = useState<SimpleState>(initialSimple ?? DEFAULT_SIMPLE);

  // When the incoming `value` changes externally (e.g. edit page hydrating),
  // re-sync once if we haven't user-edited yet.
  const hydrated = useRef(false);
  useEffect(() => {
    if (hydrated.current) return;
    if (!value) return;
    const parsed = parseSimple(value);
    if (parsed) {
      setSimple(parsed);
      setMode("simple");
    } else {
      setMode("advanced");
    }
    hydrated.current = true;
  }, [value]);

  // Push simple-built cron upward whenever it changes (only while in simple mode).
  useEffect(() => {
    if (mode !== "simple") return;
    const next = buildCron(simple);
    if (next !== value) onChange(next);
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [simple, mode]);

  function updateSimple<K extends keyof SimpleState>(key: K, v: SimpleState[K]) {
    setSimple((prev) => ({ ...prev, [key]: v }));
  }

  function toggleWeekday(d: number) {
    setSimple((prev) => {
      const has = prev.weekdays.includes(d);
      const next = has ? prev.weekdays.filter((x) => x !== d) : [...prev.weekdays, d];
      return { ...prev, weekdays: next };
    });
  }

  const preview = buildCron(simple);
  const invalid = mode === "simple" && simple.frequency === "weekly" && simple.weekdays.length === 0;

  return (
    <div className="space-y-3">
      <Tabs
        value={mode}
        onValueChange={(v) => setMode(v as "simple" | "advanced")}
      >
        <TabsList>
          <TabsTrigger value="simple">Simple</TabsTrigger>
          <TabsTrigger value="advanced">Advanced (cron)</TabsTrigger>
        </TabsList>

        <TabsContent value="simple" className="mt-3 space-y-3">
          <div className="grid gap-3 sm:grid-cols-2">
            <div className="space-y-1.5">
              <label className="text-sm font-medium">Frequency</label>
              <Select
                value={simple.frequency}
                onValueChange={(v) => updateSimple("frequency", v as Frequency)}
              >
                <SelectTrigger><SelectValue /></SelectTrigger>
                <SelectContent>
                  <SelectItem value="minutes">Every N minutes</SelectItem>
                  <SelectItem value="hourly">Hourly</SelectItem>
                  <SelectItem value="daily">Daily</SelectItem>
                  <SelectItem value="weekly">Weekly</SelectItem>
                  <SelectItem value="monthly">Monthly</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {simple.frequency === "minutes" && (
              <div className="space-y-1.5">
                <label className="text-sm font-medium">Interval (minutes)</label>
                <Input
                  type="number"
                  min={1}
                  max={59}
                  value={simple.interval}
                  onChange={(e) => updateSimple("interval", Math.max(1, Math.min(59, parseInt(e.target.value, 10) || 1)))}
                />
              </div>
            )}

            {simple.frequency === "hourly" && (
              <>
                <div className="space-y-1.5">
                  <label className="text-sm font-medium">Interval (hours)</label>
                  <Input
                    type="number"
                    min={1}
                    max={23}
                    value={simple.interval}
                    onChange={(e) => updateSimple("interval", Math.max(1, Math.min(23, parseInt(e.target.value, 10) || 1)))}
                  />
                </div>
                <div className="space-y-1.5">
                  <label className="text-sm font-medium">At minute</label>
                  <Input
                    type="number"
                    min={0}
                    max={59}
                    value={simple.minute}
                    onChange={(e) => updateSimple("minute", Math.max(0, Math.min(59, parseInt(e.target.value, 10) || 0)))}
                  />
                </div>
              </>
            )}

            {(simple.frequency === "daily" ||
              simple.frequency === "weekly" ||
              simple.frequency === "monthly") && (
              <div className="space-y-1.5">
                <label className="text-sm font-medium">Time</label>
                <Input
                  type="time"
                  value={`${pad2(simple.hour)}:${pad2(simple.minute)}`}
                  onChange={(e) => {
                    const [hh, mm] = e.target.value.split(":").map((x) => parseInt(x, 10) || 0);
                    setSimple((prev) => ({ ...prev, hour: hh, minute: mm }));
                  }}
                />
              </div>
            )}

            {simple.frequency === "monthly" && (
              <div className="space-y-1.5">
                <label className="text-sm font-medium">Day of month</label>
                <Input
                  type="number"
                  min={1}
                  max={31}
                  value={simple.dayOfMonth}
                  onChange={(e) => updateSimple("dayOfMonth", Math.max(1, Math.min(31, parseInt(e.target.value, 10) || 1)))}
                />
              </div>
            )}
          </div>

          {simple.frequency === "weekly" && (
            <div className="space-y-1.5">
              <label className="text-sm font-medium">Days of week</label>
              <div className="flex flex-wrap gap-1.5">
                {WEEKDAY_LABELS.map((label, idx) => {
                  const active = simple.weekdays.includes(idx);
                  return (
                    <Button
                      key={label}
                      type="button"
                      variant={active ? "default" : "outline"}
                      size="sm"
                      className={cn("min-w-[3.25rem]", active && "shadow-sm")}
                      onClick={() => toggleWeekday(idx)}
                    >
                      {label}
                    </Button>
                  );
                })}
              </div>
            </div>
          )}

          <div
            className={cn(
              "rounded-md border bg-muted/40 px-3 py-2 text-sm",
              invalid && "border-destructive/50 bg-destructive/5",
            )}
          >
            <div className={cn("text-muted-foreground", invalid && "text-destructive")}>
              {describe(simple)}
            </div>
            <div className="mt-1 font-mono text-xs">
              {invalid ? <span className="text-destructive">invalid</span> : preview}
            </div>
          </div>
        </TabsContent>

        <TabsContent value="advanced" className="mt-3 space-y-2">
          <Input
            placeholder="0 9 * * *  (or @daily, @every 1h)"
            value={value}
            onChange={(e) => onChange(e.target.value)}
            className="font-mono"
          />
          <p className="text-xs text-muted-foreground">
            Standard 5-field cron (minute hour day-of-month month day-of-week), or
            shortcuts like <code className="font-mono">@hourly</code>,{" "}
            <code className="font-mono">@daily</code>,{" "}
            <code className="font-mono">@weekly</code>,{" "}
            <code className="font-mono">@every 30m</code>.
          </p>
        </TabsContent>
      </Tabs>
    </div>
  );
}
