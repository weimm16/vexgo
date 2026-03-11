// hooks/useSSOProviders.ts
import { useEffect, useState } from "react";

export type SSOProvider = "github" | "google" | "oidc";

interface SSOProvidersResponse {
  providers: SSOProvider[];
  allow_local_login: boolean;
}

interface UseSSOProvidersResult {
  providers: SSOProvider[];
  allowLocalLogin: boolean;
  loading: boolean;
}

export function useSSOProviders(): UseSSOProvidersResult {
  const [providers, setProviders] = useState<SSOProvider[]>([]);
  const [allowLocalLogin, setAllowLocalLogin] = useState(true);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch("/api/sso/providers")
      .then((r) => r.json())
      .then((data: SSOProvidersResponse) => {
        setProviders(data.providers ?? []);
        setAllowLocalLogin(data.allow_local_login ?? true);
      })
      .catch(() => {
        // On error: show nothing, keep local login
        setProviders([]);
        setAllowLocalLogin(true);
      })
      .finally(() => setLoading(false));
  }, []);

  return { providers, allowLocalLogin, loading };
}
