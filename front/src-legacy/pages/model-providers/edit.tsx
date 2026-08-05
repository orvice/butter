import { useParams, useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { useModelProvider, useUpdateModelProvider } from "@/api/model-providers";
import { Page, PageHeader, PageScroll } from "@/components/butter/page-parts";
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

  if (isLoading) {
    return (
      <div className="p-6">
        <Skeleton className="h-96 w-full" />
      </div>
    );
  }

  return (
    <Page>
      <PageHeader className="max-w-3xl" title="Edit Model Provider" subtitle={name} />
      <PageScroll className="max-w-3xl">
        <ModelProviderForm
          mode="edit"
          initialValue={data?.model_provider}
          submitLabel="Save"
          loading={updateMutation.isPending}
          onCancel={() => navigate("/model-providers")}
          onSubmit={onSubmit}
        />
      </PageScroll>
    </Page>
  );
}
