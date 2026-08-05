import { Link, type LinkProps } from "@tanstack/react-router";
import {
  Bell,
  Bot,
  BrainCircuit,
  Building2,
  Cable,
  CalendarClock,
  ChevronRight,
  Cpu,
  FileText,
  FolderOpen,
  KeyRound,
  LifeBuoy,
  MessagesSquare,
  Server,
  ShieldCheck,
  UserCircle,
  Users,
  type LucideIcon,
} from "lucide-react";
import { useAuth } from "@/hooks/use-auth";
import { Page, PageHeader, PageScroll } from "@/components/butter/page-parts";
import { cn } from "@/lib/utils";

type ManageItem = {
  label: string;
  description: string;
  to: string;
  icon: LucideIcon;
  external?: boolean;
};

type ManageGroup = {
  title: string;
  description: string;
  items: ManageItem[];
};

const GROUPS: ManageGroup[] = [
  {
    title: "Agents",
    description: "Build agents and configure the resources they use.",
    items: [
      { label: "Agents", description: "Create and manage agent definitions.", to: "/agents", icon: Bot },
      { label: "Agent Files", description: "Manage files mounted into agents.", to: "/agent-files", icon: FolderOpen },
      { label: "Model Providers", description: "Configure models and provider credentials.", to: "/model-providers", icon: BrainCircuit },
    ],
  },
  {
    title: "Connections",
    description: "Connect Butter to tools, runtimes, and messaging platforms.",
    items: [
      { label: "Integrations", description: "Manage MCP servers and remote agents.", to: "/integrations", icon: Server },
      { label: "Channels", description: "Configure Telegram, Discord, and other channels.", to: "/channels", icon: Cable },
      { label: "Execution", description: "Manage daemon execution environments.", to: "/daemons", icon: Cpu },
    ],
  },
  {
    title: "Automation & Activity",
    description: "Schedule work and inspect background activity.",
    items: [
      { label: "Automations", description: "Schedule recurring agent runs and deliveries.", to: "/automations", icon: CalendarClock },
      { label: "Sessions", description: "Inspect saved agent sessions.", to: "/sessions", icon: FileText },
      { label: "Notify Groups", description: "Configure notification destinations.", to: "/notify-groups", icon: Bell },
    ],
  },
  {
    title: "Workspace & Access",
    description: "Manage workspace scope, credentials, and personal preferences.",
    items: [
      { label: "Workspaces", description: "Create and switch workspace environments.", to: "/workspaces", icon: Building2 },
      { label: "API Tokens", description: "Issue and revoke API access tokens.", to: "/api-tokens", icon: KeyRound },
      { label: "Profile", description: "Update your profile and appearance.", to: "/profile", icon: UserCircle },
    ],
  },
  {
    title: "Community & Help",
    description: "Open collaboration and support resources.",
    items: [
      { label: "Forum", description: "Browse workspace discussions.", to: "/forum", icon: MessagesSquare },
      {
        label: "Documentation",
        description: "Read Butter documentation on GitHub.",
        to: "https://github.com/orvice/butter",
        icon: FileText,
        external: true,
      },
      {
        label: "Support",
        description: "Report issues and request help.",
        to: "https://github.com/orvice/butter/issues",
        icon: LifeBuoy,
        external: true,
      },
    ],
  },
];

const ADMIN_GROUP: ManageGroup = {
  title: "Administration",
  description: "Manage global users and shared infrastructure.",
  items: [
    { label: "Users", description: "Manage user accounts and access.", to: "/admin/users", icon: Users },
    { label: "Global MCP", description: "Manage globally available MCP servers.", to: "/admin/global-mcp-servers", icon: ShieldCheck },
  ],
};

function ManageLink({ item }: { item: ManageItem }) {
  const content = (
    <>
      <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-md bg-primary/10 text-primary">
        <item.icon className="h-4 w-4" />
      </span>
      <span className="min-w-0 flex-1">
        <span className="block font-medium text-foreground">{item.label}</span>
        <span className="mt-0.5 block text-xs leading-5 text-muted-foreground">{item.description}</span>
      </span>
      <ChevronRight className="h-4 w-4 shrink-0 text-muted-foreground transition-transform group-hover:translate-x-0.5" />
    </>
  );

  const className = cn(
    "group flex items-center gap-3 rounded-md border border-transparent px-3 py-2.5",
    "transition-colors hover:border-border hover:bg-muted/60",
  );

  if (item.external) {
    return (
      <a href={item.to} target="_blank" rel="noreferrer" className={className}>
        {content}
      </a>
    );
  }

  return (
    <Link to={item.to as LinkProps['to']} className={className}>
      {content}
    </Link>
  );
}

function ManageSection({ group }: { group: ManageGroup }) {
  return (
    <section className="overflow-hidden rounded-lg border border-border bg-card">
      <div className="border-b border-border px-4 py-3">
        <h2 className="text-sm font-semibold">{group.title}</h2>
        <p className="mt-0.5 text-xs text-muted-foreground">{group.description}</p>
      </div>
      <div className="grid gap-1 p-2 sm:grid-cols-2">
        {group.items.map((item) => (
          <ManageLink key={item.label} item={item} />
        ))}
      </div>
    </section>
  );
}

export function ManagePage() {
  const { isAdmin } = useAuth();
  const groups = isAdmin ? [...GROUPS, ADMIN_GROUP] : GROUPS;

  return (
    <Page>
      <PageHeader
        title="Manage"
        subtitle="Configure agents, connections, automations, and workspace access."
      />
      <PageScroll>
        <div className="grid gap-4 lg:grid-cols-2">
          {groups.map((group) => (
            <ManageSection key={group.title} group={group} />
          ))}
        </div>
      </PageScroll>
    </Page>
  );
}
