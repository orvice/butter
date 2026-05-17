import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { useCreateModelProvider } from "@/api/model-providers";
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbSeparator } from "@/components/ui/breadcrumb";
import ModelProviderForm from "./form";
import type { ModelProvider } from "@/types/api";

export default function ModelProviderCreatePage() {
  const navigate = useNavigate();
  const createMutation = useCreateModelProvider();

  function onSubmit(provider: ModelProvider) {
    createMutation.mutate(provider, {
      onSuccess: () => {
        toast.success("Model provider created");
        navigate("/model-providers");
      },
      onError: (err) => toast.error(err.message),
    });
  }

  return (
    <>
      <Breadcrumb className="mb-4">
        <BreadcrumbList>
          <BreadcrumbItem><BreadcrumbLink href="/model-providers">Model Providers</BreadcrumbLink></BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>Create</BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <h2 className="mb-6 text-2xl font-bold">Create Model Provider</h2>
      <ModelProviderForm
        mode="create"
        submitLabel="Create"
        loading={createMutation.isPending}
        onCancel={() => navigate("/model-providers")}
        onSubmit={onSubmit}
      />
    </>
  );
}
