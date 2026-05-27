/* eslint-disable react-refresh/only-export-components */

import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from "react";
import { LAYOUT_DENSITY_KEY } from "@/lib/constants";

export type LayoutDensity = "comfortable" | "compact";

type LayoutDensityContextValue = {
  density: LayoutDensity;
  isCompact: boolean;
  setDensity: (density: LayoutDensity) => void;
  toggleDensity: () => void;
};

const LayoutDensityContext = createContext<LayoutDensityContextValue | null>(null);

function isLayoutDensity(value: string | null): value is LayoutDensity {
  return value === "comfortable" || value === "compact";
}

function getInitialDensity(): LayoutDensity {
  if (typeof window === "undefined") return "comfortable";
  try {
    const storedDensity = window.localStorage.getItem(LAYOUT_DENSITY_KEY);
    return isLayoutDensity(storedDensity) ? storedDensity : "comfortable";
  } catch {
    return "comfortable";
  }
}

export function LayoutDensityProvider({ children }: { children: ReactNode }) {
  const [density, setDensityState] = useState<LayoutDensity>(getInitialDensity);

  useEffect(() => {
    document.documentElement.dataset.density = density;
    try {
      window.localStorage.setItem(LAYOUT_DENSITY_KEY, density);
    } catch {
      // Keep the live density preference even when persistence is unavailable.
    }
  }, [density]);

  const value = useMemo<LayoutDensityContextValue>(
    () => ({
      density,
      isCompact: density === "compact",
      setDensity: setDensityState,
      toggleDensity: () => setDensityState((current) => (current === "compact" ? "comfortable" : "compact")),
    }),
    [density],
  );

  return <LayoutDensityContext.Provider value={value}>{children}</LayoutDensityContext.Provider>;
}

export function useLayoutDensity() {
  const context = useContext(LayoutDensityContext);
  if (!context) {
    throw new Error("useLayoutDensity must be used within LayoutDensityProvider");
  }
  return context;
}
