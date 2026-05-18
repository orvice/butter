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
    <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <h2 className="text-xl font-bold tracking-tight sm:text-2xl">{title}</h2>
      {createLabel && createTo && (
        <Button className="w-full sm:w-auto" onClick={() => navigate(createTo)}>
          <Plus className="mr-2 h-4 w-4" />
          {createLabel}
        </Button>
      )}
    </div>
  );
}
