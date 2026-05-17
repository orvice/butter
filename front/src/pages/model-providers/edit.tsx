import { useParams, useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { useModelProvider, useUpdateModelProvider } from "@/api/model-providers";
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbSeparator } from "@/components/ui/breadcrumb";
import { Skeleton } from "@/components/ui/skeleton";
import ModelProviderForm from "./form";
import type { ModelProvider } from "@/types/api";

export default function ModelProviderEditPage() {
  const { name } = useParams<{ name: string }>();
  const navigate = useNavigate();
  const { data, isLoading } = useModelProvider(name ?? "");
  const updateMutation = useUpdateModelProvider();

  function onSubmit(provider: ModelProvider) {
    updateMutation.mutate(provider, {
      onSuccess: () => {
        toast.success("Model provider updated");
        navigate("/model-providers");
      },
      onError: (err) => toast.error(err.message),
    });
  }

  if (isLoading) return <Skeleton className="h-96 w-full" />;

  return (
    <>
      <Breadcrumb className="mb-4">
        <BreadcrumbList>
          <BreadcrumbItem><BreadcrumbLink href="/model-providers">Model Providers</BreadcrumbLink></BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>{name}</BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <h2 className="mb-6 text-2xl font-bold">Edit Model Provider</h2>
      <ModelProviderForm
        mode="edit"
        initialValue={data?.model_provider}
        submitLabel="Save"
        loading={updateMutation.isPending}
        onCancel={() => navigate("/model-providers")}
        onSubmit={onSubmit}
      />
    </>
  );
}
