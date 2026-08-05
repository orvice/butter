import { useEffect, useRef, useState } from "react";
import { toast } from "sonner";
import { cn } from "@/lib/utils";
import {
  MAX_TITLE_CODE_POINTS,
  normalizeTitleInput,
  titleCodePointCount,
} from "@/lib/session-title";

interface InlineTitleInputProps {
  /** Current effective title used to seed the draft. */
  initial: string;
  /** Persists the normalized title; must throw/reject on failure. */
  onSave: (title: string) => Promise<void>;
  /** Leaves edit mode (after a successful save, or on cancel). */
  onClose: () => void;
  className?: string;
}

/**
 * Inline session-title editor: Enter and blur save a valid edit, Escape
 * cancels, and a failed save keeps the draft visible so it can be retried.
 */
export function InlineTitleInput({ initial, onSave, onClose, className }: InlineTitleInputProps) {
  const [draft, setDraft] = useState(initial);
  const [saving, setSaving] = useState(false);
  const inputRef = useRef<HTMLInputElement | null>(null);
  // Set once Escape/Enter resolved the edit, so the input's unmount blur
  // cannot fire a second commit.
  const doneRef = useRef(false);

  useEffect(() => {
    inputRef.current?.focus();
    inputRef.current?.select();
  }, []);

  function cancel() {
    if (doneRef.current) return;
    doneRef.current = true;
    onClose();
  }

  async function commit() {
    if (doneRef.current || saving) return;
    const value = normalizeTitleInput(draft);
    if (!value || value === normalizeTitleInput(initial)) {
      cancel();
      return;
    }
    if (titleCodePointCount(value) > MAX_TITLE_CODE_POINTS) {
      toast.error(`Title must be at most ${MAX_TITLE_CODE_POINTS} characters`);
      return;
    }
    setSaving(true);
    try {
      await onSave(value);
      doneRef.current = true;
      onClose();
    } catch (err) {
      // Keep the draft and edit mode so the rename can be retried as-is.
      toast.error(err instanceof Error ? err.message : "Failed to rename chat");
      setSaving(false);
    }
  }

  return (
    <input
      ref={inputRef}
      value={draft}
      disabled={saving}
      onChange={(e) => setDraft(e.target.value)}
      onKeyDown={(e) => {
        if (e.key === "Enter" && !e.nativeEvent.isComposing) {
          e.preventDefault();
          void commit();
        } else if (e.key === "Escape") {
          e.preventDefault();
          cancel();
        }
      }}
      onBlur={() => void commit()}
      aria-label="Chat title"
      className={cn(
        "w-full min-w-0 rounded border border-ring bg-background px-1.5 py-0.5 text-sm outline-none disabled:opacity-60",
        className,
      )}
    />
  );
}
