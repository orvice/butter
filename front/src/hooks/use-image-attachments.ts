import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import type { ChangeEvent, ClipboardEvent, DragEvent } from "react";
import {
  acceptImageFiles,
  IMAGE_FILE_ACCEPT,
  validateImageFiles,
} from "@/lib/image-attachments";
import { toast } from "sonner";

// A picked image paired with its preview object URL. The URL is created
// once when the file is attached (event handler, not render — creating it
// in a useMemo leaks under StrictMode double-invocation and can revoke
// still-displayed URLs in discarded concurrent renders) and revoked when
// the attachment is removed, cleared, or dropped on session switch. Each
// file keeps its URL for its whole lifetime, so thumbnails never flicker
// on add/remove.
interface AttachmentItem {
  file: File;
  previewUrl: string;
}

function revokeItems(items: AttachmentItem[]) {
  items.forEach((item) => URL.revokeObjectURL(item.previewUrl));
}

export function useImageAttachments(sessionId: string) {
  const [items, setItems] = useState<AttachmentItem[]>([]);
  const [isDragOver, setIsDragOver] = useState(false);
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const itemsRef = useRef(items);
  const dragCountRef = useRef(0);

  // Keep a ref in sync for event handlers and the unmount cleanup; updated
  // in an effect (not render) per the react-hooks/refs rule. Event handlers
  // always run after effects, so they never observe a stale value.
  useEffect(() => {
    itemsRef.current = items;
  }, [items]);

  const attachments = useMemo(() => items.map((item) => item.file), [items]);
  const previewUrls = useMemo(() => items.map((item) => item.previewUrl), [items]);

  useEffect(() => {
    return () => {
      revokeItems(itemsRef.current);
      itemsRef.current = [];
    };
  }, []);

  // Drop attachments when the user switches sessions (state adjustment
  // during render, per React's "you might not need an effect" guidance).
  // revokeObjectURL is idempotent, so a StrictMode double render is
  // harmless.
  const [attachmentsSessionId, setAttachmentsSessionId] = useState(sessionId);
  if (attachmentsSessionId !== sessionId) {
    setAttachmentsSessionId(sessionId);
    revokeItems(items);
    setItems([]);
  }

  const addFiles = useCallback((files: File[]) => {
    if (files.length === 0) return;
    const { accepted, errors } = acceptImageFiles(
      itemsRef.current.map((item) => item.file),
      files,
    );
    errors.forEach((msg) => toast.error(msg));
    if (accepted.length > 0) {
      const added = accepted.map((file) => ({ file, previewUrl: URL.createObjectURL(file) }));
      setItems((prev) => [...prev, ...added]);
    }
  }, []);

  const removeAttachment = useCallback((index: number) => {
    const target = itemsRef.current[index];
    if (target) URL.revokeObjectURL(target.previewUrl);
    setItems((prev) => prev.filter((_, i) => i !== index));
  }, []);

  const clearAttachments = useCallback(() => {
    revokeItems(itemsRef.current);
    setItems([]);
  }, []);

  const validateForSend = useCallback((): string[] => {
    return validateImageFiles(itemsRef.current.map((item) => item.file));
  }, []);

  const handleDragOver = useCallback((e: DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
  }, []);

  const handleDragEnter = useCallback((e: DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    dragCountRef.current += 1;
    if (dragCountRef.current === 1) setIsDragOver(true);
  }, []);

  const handleDragLeave = useCallback((e: DragEvent) => {
    e.preventDefault();
    e.stopPropagation();
    dragCountRef.current -= 1;
    if (dragCountRef.current === 0) setIsDragOver(false);
  }, []);

  const handleDrop = useCallback(
    (e: DragEvent) => {
      e.preventDefault();
      e.stopPropagation();
      dragCountRef.current = 0;
      setIsDragOver(false);
      addFiles(Array.from(e.dataTransfer.files));
    },
    [addFiles],
  );

  const handlePaste = useCallback(
    (e: ClipboardEvent) => {
      const files = Array.from(e.clipboardData.items)
        .filter((item) => item.kind === "file" && item.type.startsWith("image/"))
        .map((item) => item.getAsFile())
        .filter((f): f is File => f !== null);
      if (files.length > 0) {
        e.preventDefault();
        addFiles(files);
      }
    },
    [addFiles],
  );

  const openFilePicker = useCallback(() => {
    fileInputRef.current?.click();
  }, []);

  const handleFileInputChange = useCallback(
    (e: ChangeEvent<HTMLInputElement>) => {
      addFiles(Array.from(e.target.files ?? []));
      e.target.value = "";
    },
    [addFiles],
  );

  return {
    attachments,
    previewUrls,
    isDragOver,
    addFiles,
    removeAttachment,
    clearAttachments,
    validateForSend,
    handleDragOver,
    handleDragEnter,
    handleDragLeave,
    handleDrop,
    handlePaste,
    fileInputRef,
    openFilePicker,
    handleFileInputChange,
    fileAccept: IMAGE_FILE_ACCEPT,
  };
}
