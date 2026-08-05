import { useNavigate } from "react-router-dom";
import { toast } from "sonner";
import { useCreateModelProvider } from "@/api/model-providers";
import { Page, PageHeader, PageScroll } from "@/components/butter/page-parts";
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
    <Page>
      <PageHeader
        className="max-w-3xl"
        title="Create Model Provider"
        subtitle="Register an LLM provider and the models it exposes to agents and channels."
      />
      <PageScroll className="max-w-3xl">
        <ModelProviderForm
          mode="create"
          submitLabel="Create"
          loading={createMutation.isPending}
          onCancel={() => navigate("/model-providers")}
          onSubmit={onSubmit}
        />
      </PageScroll>
    </Page>
  );
}
