import { useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Plus } from "lucide-react";

interface PageHeaderProps {
  title: string;
  createLabel?: string;
  createTo?: string;
}

export function PageHeader({ title, createLabel, createTo }: PageHeaderProps) {
  const navigate = useNavigate();
  return (
    <div className="mb-6 flex items-center justify-between">
      <h2 className="text-2xl font-bold tracking-tight">{title}</h2>
      {createLabel && createTo && (
        <Button onClick={() => navigate(createTo)}>
          <Plus className="mr-2 h-4 w-4" />
          {createLabel}
        </Button>
      )}
    </div>
  );
}
