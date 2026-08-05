import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "@/hooks/use-auth";
import { beginOAuthFlow, listOAuthProviders, type OAuthProviderInfo } from "@/api/auth";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent } from "@/components/ui/card";
import { ThemeControls } from "@/components/theme-controls";
import { OAuthProviderIcon } from "@/components/oauth-provider-icon";
import { BrandMark } from "@/components/brand-mark";

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
    <div className="relative flex min-h-[100dvh] items-center justify-center overflow-hidden bg-background p-5">
      <div className="absolute inset-0 bg-[url('/md-theme/body.jpg')] bg-cover bg-center opacity-70 dark:hidden" />
      <div className="absolute inset-0 hidden bg-[radial-gradient(circle_at_20%_10%,rgba(30,145,255,0.24),transparent_30%),radial-gradient(circle_at_80%_20%,rgba(57,187,176,0.15),transparent_32%),var(--background)] dark:block" />
      <div className="absolute right-4 top-4 z-10">
        <ThemeControls className="rounded-md border border-border bg-card/80 p-1 shadow-card backdrop-blur" />
      </div>
      <Card className="relative z-10 w-full max-w-[400px] p-4 shadow-dropdown">
        <CardContent className="p-4">
          <div className="mb-6 flex items-center gap-3">
            <BrandMark size={42} />
            <div>
              <h1 className="text-xl font-semibold leading-tight text-card-foreground">Welcome to Butter</h1>
              <p className="text-sm text-muted-foreground">Sign in to continue to the dashboard</p>
            </div>
          </div>
          <form onSubmit={handleSubmit} className="mb-5 space-y-4">
            <div className="space-y-1.5">
              <label htmlFor="login-username" className="text-sm font-medium text-card-foreground">Username</label>
              <Input
                id="login-username"
                type="text"
                autoComplete="username"
                placeholder="Username"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                disabled={loading}
              />
            </div>
            <div className="space-y-1.5">
              <label htmlFor="login-password" className="text-sm font-medium text-card-foreground">Password</label>
              <Input
                id="login-password"
                type="password"
                autoComplete="current-password"
                placeholder="Password"
                value={password}
                onChange={(e) => setPassword(e.target.value)}
                disabled={loading}
              />
            </div>
            {error && <p className="text-sm text-destructive">{error}</p>}
            <Button type="submit" className="w-full" size="lg" disabled={loading || !username.trim() || !password}>
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
