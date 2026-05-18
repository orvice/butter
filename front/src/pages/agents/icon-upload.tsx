import { useEffect, useMemo, useState } from "react";
import { Bot, Upload } from "lucide-react";
import { toast } from "sonner";
import { useUploadAvatar } from "@/api/uploads";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";

interface AgentIconUploadProps {
  agentName: string;
  value?: string;
  onChange: (url: string) => void;
}

export function AgentIconUpload({ agentName, value, onChange }: AgentIconUploadProps) {
  const uploadAvatar = useUploadAvatar();
  const [file, setFile] = useState<File | null>(null);

  const previewUrl = useMemo(() => {
    if (!file) return "";
    return URL.createObjectURL(file);
  }, [file]);

  useEffect(() => {
    return () => {
      if (previewUrl) URL.revokeObjectURL(previewUrl);
    };
  }, [previewUrl]);

  function handleFileChange(nextFile: File | undefined) {
    if (!nextFile) {
      setFile(null);
      return;
    }
    if (!nextFile.type.startsWith("image/")) {
      toast.error("Choose an image file");
      return;
    }
    setFile(nextFile);
  }

  function handleUpload() {
    const ownerId = agentName.trim();
    if (!ownerId) {
      toast.error("Enter the agent name first");
      return;
    }
    if (!file) {
      toast.error("Choose an icon image");
      return;
    }

    uploadAvatar.mutate(
      { file, ownerKind: "agent", ownerId },
      {
        onSuccess: (res) => {
          onChange(res.url);
          toast.success("Agent icon uploaded");
        },
        onError: (err) => toast.error(err.message),
      },
    );
  }

  const displayUrl = previewUrl || value || "";

  return (
    <div className="space-y-4">
      <div className="flex flex-col gap-4 sm:flex-row sm:items-center">
        <div className="flex h-16 w-16 shrink-0 items-center justify-center overflow-hidden rounded-md border bg-muted">
          {displayUrl ? (
            <img src={displayUrl} alt="" className="h-full w-full object-cover" />
          ) : (
            <Bot className="h-7 w-7 text-muted-foreground" />
          )}
        </div>
        <div className="min-w-0 flex-1 space-y-2">
          <Label htmlFor="agent-icon-file">Icon image</Label>
          <Input
            id="agent-icon-file"
            type="file"
            accept="image/png,image/jpeg,image/gif,image/webp"
            onChange={(e) => handleFileChange(e.target.files?.[0])}
          />
          <div className="text-xs text-muted-foreground">PNG, JPEG, GIF, and WebP images are supported.</div>
        </div>
      </div>
      <div className="flex flex-wrap gap-2">
        <Button type="button" onClick={handleUpload} disabled={!file || uploadAvatar.isPending}>
          <Upload className="mr-2 h-4 w-4" />
          {uploadAvatar.isPending ? "Uploading..." : "Upload icon"}
        </Button>
        {value ? (
          <Button type="button" variant="outline" onClick={() => onChange("")}>
            Clear icon
          </Button>
        ) : null}
      </div>
      {value ? (
        <div className="space-y-1">
          <Label htmlFor="agent-icon-url">Icon URL</Label>
          <Input id="agent-icon-url" readOnly value={value} />
        </div>
      ) : null}
    </div>
  );
}
