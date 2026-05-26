import { useEffect, useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "@/hooks/use-auth";
import { beginOAuthFlow, listOAuthProviders, type OAuthProviderInfo } from "@/api/auth";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
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
    <div className="flex min-h-[100dvh] items-center justify-center bg-background px-4">
      <div className="absolute right-4 top-4">
        <ThemeControls />
      </div>
      <Card className="w-full max-w-sm">
        <CardHeader className="text-center">
          <BrandMark className="mx-auto mb-3" size={44} />
          <CardTitle className="text-2xl font-semibold tracking-tight">Butter</CardTitle>
          <CardDescription>Sign in to continue</CardDescription>
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
