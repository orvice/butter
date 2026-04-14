import { useNavigate } from "react-router-dom";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { z } from "zod";
import { toast } from "sonner";
import { useCreateRemoteAgent } from "@/api/remote-agents";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "@/components/ui/form";
import { Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList, BreadcrumbSeparator } from "@/components/ui/breadcrumb";

const schema = z.object({
  id: z.string().min(1, "ID is required"),
  name: z.string().min(1, "Name is required"),
  url: z.string().min(1, "URL is required").url("Must be a valid URL"),
});

type FormValues = z.infer<typeof schema>;

export default function RemoteAgentCreatePage() {
  const navigate = useNavigate();
  const createMutation = useCreateRemoteAgent();
  const form = useForm<FormValues>({ resolver: zodResolver(schema), defaultValues: { id: "", name: "", url: "" } });

  function onSubmit(values: FormValues) {
    createMutation.mutate(
      { id: values.id, name: values.name, url: values.url, protocol: "REMOTE_AGENT_PROTOCOL_A2A" },
      {
        onSuccess: () => { toast.success("Remote agent created"); navigate("/remote-agents"); },
        onError: (err) => toast.error(err.message),
      },
    );
  }

  return (
    <>
      <Breadcrumb className="mb-4">
        <BreadcrumbList>
          <BreadcrumbItem><BreadcrumbLink href="/remote-agents">Remote Agents</BreadcrumbLink></BreadcrumbItem>
          <BreadcrumbSeparator />
          <BreadcrumbItem>Create</BreadcrumbItem>
        </BreadcrumbList>
      </Breadcrumb>
      <h2 className="mb-6 text-2xl font-bold">Create Remote Agent</h2>
      <Form {...form}>
        <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-6">
          <Card>
            <CardHeader><CardTitle>Remote Agent</CardTitle></CardHeader>
            <CardContent className="space-y-4">
              <FormField control={form.control} name="id" render={({ field }) => (
                <FormItem><FormLabel>ID</FormLabel><FormControl><Input placeholder="remote-assistant" {...field} /></FormControl><FormMessage /></FormItem>
              )} />
              <FormField control={form.control} name="name" render={({ field }) => (
                <FormItem><FormLabel>Name</FormLabel><FormControl><Input placeholder="Remote Assistant" {...field} /></FormControl><FormMessage /></FormItem>
              )} />
              <FormField control={form.control} name="url" render={({ field }) => (
                <FormItem><FormLabel>URL</FormLabel><FormControl><Input placeholder="https://agent.example.com/a2a" {...field} /></FormControl><FormMessage /></FormItem>
              )} />
            </CardContent>
          </Card>
          <div className="flex gap-3">
            <Button type="button" variant="outline" onClick={() => navigate("/remote-agents")}>Cancel</Button>
            <Button type="submit" disabled={createMutation.isPending}>{createMutation.isPending ? "Creating..." : "Create"}</Button>
          </div>
        </form>
      </Form>
    </>
  );
}
