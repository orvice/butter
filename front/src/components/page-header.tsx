import { useNavigate } from "react-router-dom";
import { Button } from "@/components/ui/button";
import { Plus } from "lucide-react";

interface PageHeaderProps {
  title: string;
  description?: string;
  createLabel?: string;
  createTo?: string;
}

export function PageHeader({ title, description, createLabel, createTo }: PageHeaderProps) {
  const navigate = useNavigate();
  return (
    <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <div>
        <h2 className="text-2xl font-semibold tracking-tight text-foreground">{title}</h2>
        {description ? <p className="mt-1 text-sm text-muted-foreground">{description}</p> : null}
      </div>
      {createLabel && createTo && (
        <Button className="w-full sm:w-auto" onClick={() => navigate(createTo)}>
          <Plus className="mr-2 h-4 w-4" />
          {createLabel}
        </Button>
      )}
    </div>
  );
}
