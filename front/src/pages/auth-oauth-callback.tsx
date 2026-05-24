import { useEffect, useRef, useState } from "react";
import { useNavigate, useParams, useSearchParams } from "react-router-dom";
import { useAuth } from "@/hooks/use-auth";
import { completeOAuthFlow } from "@/api/auth";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { Button } from "@/components/ui/button";

export default function OAuthCallbackPage() {
  const { provider = "" } = useParams<{ provider: string }>();
  const [params] = useSearchParams();
  const navigate = useNavigate();
  const { applyLoginResponse } = useAuth();
  const errParam = params.get("error");
  const code = params.get("code") ?? "";
  const state = params.get("state") ?? "";
  const callbackError = errParam
    ? `${errParam}: ${params.get("error_description") ?? "authorization rejected"}`
    : !provider || !code || !state
      ? "Missing provider, code, or state in callback URL."
      : "";
  const [error, setError] = useState<string>(callbackError);
  const consumed = useRef(false);

  useEffect(() => {
    if (callbackError) return;
    if (consumed.current) return;
    consumed.current = true;

    completeOAuthFlow(provider, code, state)
      .then((res) => {
        if (!res.token) {
          setError("Login response missing token.");
          return;
        }
        applyLoginResponse(res);
        navigate("/", { replace: true });
      })
      .catch((e: unknown) => {
        setError(e instanceof Error ? e.message : "OAuth login failed.");
      });
  }, [provider, code, state, callbackError, applyLoginResponse, navigate]);

  return (
    <div className="flex min-h-screen items-center justify-center bg-background px-4">
      <Card className="w-full max-w-sm">
        <CardHeader className="text-center">
          <CardTitle>Signing you in…</CardTitle>
          <CardDescription>Completing {provider || "OAuth"} login.</CardDescription>
        </CardHeader>
        <CardContent>
          {error ? (
            <div className="space-y-3 text-sm">
              <p className="text-destructive">{error}</p>
              <Button variant="outline" className="w-full" onClick={() => navigate("/login", { replace: true })}>
                Back to sign in
              </Button>
            </div>
          ) : (
            <p className="text-sm text-muted-foreground">Hang tight — you’ll be redirected shortly.</p>
          )}
        </CardContent>
      </Card>
    </div>
  );
}
