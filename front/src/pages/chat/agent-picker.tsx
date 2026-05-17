import { useState } from "react";
import { useAgents } from "@/api/agents";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Label } from "@/components/ui/label";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Bot } from "lucide-react";

interface AgentPickerProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onConfirm: (agentName: string) => void;
  busy?: boolean;
}

export function AgentPicker({ open, onOpenChange, onConfirm, busy }: AgentPickerProps) {
  const { data, isLoading } = useAgents({ page_size: 200 });
  const [selected, setSelected] = useState<string>("");

  const agents = data?.agents ?? [];

  function handleConfirm() {
    if (selected) onConfirm(selected);
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle className="flex items-center gap-2">
            <Bot className="h-4 w-4" /> Start a new chat
          </DialogTitle>
          <DialogDescription>Pick an agent to chat with.</DialogDescription>
        </DialogHeader>

        <div className="space-y-2">
          <Label htmlFor="agent-select">Agent</Label>
          <Select
            value={selected}
            onValueChange={(v) => setSelected(typeof v === "string" ? v : "")}
            disabled={isLoading}
          >
            <SelectTrigger id="agent-select" className="w-full">
              <SelectValue placeholder={isLoading ? "Loading agents..." : "Select an agent"} />
            </SelectTrigger>
            <SelectContent>
              {agents.map((a) => (
                <SelectItem key={a.name} value={a.name}>
                  <span className="flex flex-col items-start leading-tight">
                    <span>{a.name}</span>
                    {a.description ? (
                      <span className="text-[10px] text-muted-foreground">{a.description}</span>
                    ) : null}
                  </span>
                </SelectItem>
              ))}
              {!isLoading && agents.length === 0 ? (
                <div className="px-3 py-2 text-xs text-muted-foreground">No agents configured.</div>
              ) : null}
            </SelectContent>
          </Select>
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={() => onOpenChange(false)} disabled={busy}>
            Cancel
          </Button>
          <Button onClick={handleConfirm} disabled={!selected || busy}>
            {busy ? "Creating..." : "Start chat"}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  );
}
