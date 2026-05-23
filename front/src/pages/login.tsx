import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "@/hooks/use-auth";
import { beginOAuthFlow, listOAuthProviders, type OAuthProviderInfo } from "@/api/auth";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { ThemeControls } from "@/components/theme-controls";
import { OAuthProviderIcon } from "@/components/oauth-provider-icon";
import { Sparkles } from "lucide-react";

export default function LoginPage() {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const [oauthProviders, setOauthProviders] = useState<OAuthProviderInfo[]>([]);
  const [oauthLoading, setOauthLoading] = useState<string | null>(null);
  const { login } = useAuth();
  const navigate = useNavigate();

  useEffect(() => {
    let cancelled = false;
    listOAuthProviders()
      .then((res) => {
        if (cancelled) return;
        setOauthProviders(res.providers ?? []);
      })
      .catch(() => {
        if (!cancelled) setOauthProviders([]);
      });
    return () => {
      cancelled = true;
    };
  }, []);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      const ok = await login(username.trim(), password);
      if (ok) {
        navigate("/", { replace: true });
      } else {
        setError("Invalid username or password. Please check and try again.");
      }
    } catch {
      setError("Connection failed. Is the server running?");
    } finally {
      setLoading(false);
    }
  }

  async function handleOAuth(providerName: string) {
    setError("");
    setOauthLoading(providerName);
    try {
      const redirectUri = `${window.location.origin}/auth/oauth/callback/${providerName}`;
      const res = await beginOAuthFlow(providerName, redirectUri);
      const url = res.authorize_url ?? res.authorizeUrl;
      if (!url) {
        setError("Provider did not return an authorize URL.");
        setOauthLoading(null);
        return;
      }
      window.location.assign(url);
    } catch (e: unknown) {
      setError(e instanceof Error ? e.message : "Failed to start OAuth flow.");
      setOauthLoading(null);
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-background px-4">
      <div className="absolute right-4 top-4">
        <ThemeControls />
      </div>
      <div className="pointer-events-none absolute inset-0 -z-10 bg-[radial-gradient(circle_at_top,color-mix(in_srgb,var(--primary)_36%,transparent)_0,color-mix(in_srgb,var(--accent)_60%,transparent)_34%,transparent_62%)] dark:bg-[radial-gradient(circle_at_top,color-mix(in_srgb,var(--primary)_26%,transparent)_0,var(--background)_48%,transparent_72%)]" />
      <Card className="w-full max-w-sm border-primary/20 shadow-[0_18px_50px_color-mix(in_srgb,var(--primary)_20%,transparent)]">
        <CardHeader className="text-center">
          <div className="mx-auto mb-2 flex h-12 w-12 items-center justify-center rounded-md border border-primary/40 bg-primary text-primary-foreground shadow-[inset_0_1px_0_rgba(255,255,255,0.65)]">
            <Sparkles className="h-5 w-5" />
          </div>
          <CardTitle className="text-2xl font-black">Butter</CardTitle>
          <CardDescription>Sign in to your smooth agent workspace</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <Input
              type="text"
              autoComplete="username"
              placeholder="Username"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              disabled={loading}
            />
            <Input
              type="password"
              autoComplete="current-password"
              placeholder="Password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              disabled={loading}
            />
            {error && <p className="text-sm text-destructive">{error}</p>}
            <Button type="submit" className="w-full" disabled={loading || !username.trim() || !password}>
              {loading ? "Signing in..." : "Sign in"}
            </Button>
          </form>
          {oauthProviders.length > 0 && (
            <div className="mt-6 space-y-3">
              <div className="relative">
                <div className="absolute inset-0 flex items-center">
                  <span className="w-full border-t border-border" />
                </div>
                <div className="relative flex justify-center text-xs uppercase">
                  <span className="bg-card px-2 text-muted-foreground">Or continue with</span>
                </div>
              </div>
              <div className="grid gap-2">
                {oauthProviders.map((p) => {
                  const label = p.display_name ?? p.displayName ?? p.name;
                  return (
                    <Button
                      key={p.name}
                      type="button"
                      variant="outline"
                      className="w-full gap-2"
                      onClick={() => handleOAuth(p.name)}
                      disabled={!!oauthLoading || loading}
                    >
                      <OAuthProviderIcon provider={p.name} className="h-4 w-4 shrink-0" />
                      {oauthLoading === p.name ? `Redirecting to ${label}…` : `Sign in with ${label}`}
                    </Button>
                  );
                })}
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
