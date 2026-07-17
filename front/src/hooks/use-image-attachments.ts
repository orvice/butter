import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import type { ChangeEvent, ClipboardEvent, DragEvent } from "react";
import {
  acceptImageFiles,
  IMAGE_FILE_ACCEPT,
  validateImageFiles,
} from "@/lib/image-attachments";
import { toast } from "sonner";

export function useImageAttachments(sessionId: string) {
  const [attachments, setAttachments] = useState<File[]>([]);
  const [isDragOver, setIsDragOver] = useState(false);
  const fileInputRef = useRef<HTMLInputElement | null>(null);
  const attachmentsRef = useRef(attachments);
  attachmentsRef.current = attachments;
  const dragCountRef = useRef(0);

  // Revoke old object URLs synchronously via ref so removed thumbnails
  // don't flicker during rapid add/remove.
  const previewUrlsRef = useRef<string[]>([]);
  const previewUrls = useMemo(() => {
    previewUrlsRef.current.forEach((url) => URL.revokeObjectURL(url));
    const urls = attachments.map((file) => URL.createObjectURL(file));
    previewUrlsRef.current = urls;
    return urls;
  }, [attachments]);

  useEffect(() => {
    return () => {
      previewUrlsRef.current.forEach((url) => URL.revokeObjectURL(url));
      previewUrlsRef.current = [];
    };
  }, []);

  // Drop attachments when the user switches sessions.
  const prevSessionRef = useRef(sessionId);
  if (prevSessionRef.current !== sessionId) {
    prevSessionRef.current = sessionId;
    setAttachments([]);
  }

  const addFiles = useCallback((files: File[]) => {
    if (files.length === 0) return;
    const { accepted, errors } = acceptImageFiles(attachmentsRef.current, files);
    errors.forEach((msg) => toast.error(msg));
    if (accepted.length > 0) setAttachments((prev) => [...prev, ...accepted]);
  }, []);

  const removeAttachment = useCallback((index: number) => {
    setAttachments((prev) => prev.filter((_, i) => i !== index));
  }, []);

  const clearAttachments = useCallback(() => setAttachments([]), []);
  const restoreAttachments = useCallback((files: File[]) => setAttachments(files), []);

  const validateForSend = useCallback((): string[] => {
    return validateImageFiles(attachmentsRef.current);
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
    restoreAttachments,
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
