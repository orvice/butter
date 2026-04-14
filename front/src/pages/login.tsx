import { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useAuth } from "@/hooks/use-auth";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";

export default function LoginPage() {
  const [tokenInput, setTokenInput] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);
  const { login } = useAuth();
  const navigate = useNavigate();

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();
    setError("");
    setLoading(true);
    try {
      const ok = await login(tokenInput.trim());
      if (ok) {
        navigate("/");
      } else {
        setError("Invalid token. Please check and try again.");
      }
    } catch {
      setError("Connection failed. Is the server running?");
    } finally {
      setLoading(false);
    }
  }

  return (
    <div className="flex min-h-screen items-center justify-center bg-background">
      <Card className="w-full max-w-sm">
        <CardHeader className="text-center">
          <CardTitle className="text-2xl">Butter</CardTitle>
          <CardDescription>Enter your API token to connect</CardDescription>
        </CardHeader>
        <CardContent>
          <form onSubmit={handleSubmit} className="space-y-4">
            <Input
              type="password"
              placeholder="Paste your API token"
              value={tokenInput}
              onChange={(e) => setTokenInput(e.target.value)}
              disabled={loading}
            />
            {error && <p className="text-sm text-destructive">{error}</p>}
            <Button type="submit" className="w-full" disabled={loading || !tokenInput.trim()}>
              {loading ? "Connecting..." : "Connect"}
            </Button>
          </form>
        </CardContent>
      </Card>
    </div>
  );
}
