/* eslint-disable react-refresh/only-export-components */

import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from "react";
import { COLOR_THEME_KEY } from "@/lib/constants";

export type ColorThemeId = "butter" | "mint" | "blueberry" | "rose" | "slate" | "day" | "dark";

export type ColorTheme = {
  id: ColorThemeId;
  name: string;
  description: string;
  swatches: [string, string, string];
};

const COLOR_THEMES: ColorTheme[] = [
  {
    id: "butter",
    name: "Butter",
    description: "Warm amber and herb green",
    swatches: ["#f6c343", "#2f6f5e", "#fff2c2"],
  },
  {
    id: "mint",
    name: "Mint",
    description: "Fresh teal with clean green accents",
    swatches: ["#14b8a6", "#2563eb", "#ccfbf1"],
  },
  {
    id: "blueberry",
    name: "Blueberry",
    description: "Crisp blue with violet depth",
    swatches: ["#3b82f6", "#7c3aed", "#dbeafe"],
  },
  {
    id: "rose",
    name: "Rose",
    description: "Soft berry with warm coral notes",
    swatches: ["#e11d48", "#f97316", "#ffe4e6"],
  },
  {
    id: "slate",
    name: "Slate",
    description: "Neutral graphite with cyan signals",
    swatches: ["#475569", "#0891b2", "#e2e8f0"],
  },
  {
    id: "day",
    name: "Day",
    description: "Bright sky blue with warm sunlight",
    swatches: ["#0ea5e9", "#f59e0b", "#e0f2fe"],
  },
  {
    id: "dark",
    name: "Dark",
    description: "Deep charcoal with neutral greys",
    swatches: ["#27272a", "#52525b", "#e4e4e7"],
  },
];

type ColorThemeContextValue = {
  colorTheme: ColorThemeId;
  setColorTheme: (theme: ColorThemeId) => void;
  themes: ColorTheme[];
};

const ColorThemeContext = createContext<ColorThemeContextValue | null>(null);

function isColorThemeId(value: string | null): value is ColorThemeId {
  return COLOR_THEMES.some((theme) => theme.id === value);
}

function getInitialColorTheme(): ColorThemeId {
  if (typeof window === "undefined") return "butter";
  try {
    const storedTheme = window.localStorage.getItem(COLOR_THEME_KEY);
    return isColorThemeId(storedTheme) ? storedTheme : "butter";
  } catch {
    return "butter";
  }
}

export function ColorThemeProvider({ children }: { children: ReactNode }) {
  const [colorTheme, setColorThemeState] = useState<ColorThemeId>(getInitialColorTheme);

  useEffect(() => {
    document.documentElement.dataset.colorTheme = colorTheme;
    try {
      window.localStorage.setItem(COLOR_THEME_KEY, colorTheme);
    } catch {
      // Keep the live theme even when persistence is unavailable.
    }
  }, [colorTheme]);

  const value = useMemo<ColorThemeContextValue>(
    () => ({
      colorTheme,
      setColorTheme: setColorThemeState,
      themes: COLOR_THEMES,
    }),
    [colorTheme],
  );

  return <ColorThemeContext.Provider value={value}>{children}</ColorThemeContext.Provider>;
}

export function useColorTheme() {
  const context = useContext(ColorThemeContext);
  if (!context) {
    throw new Error("useColorTheme must be used within ColorThemeProvider");
  }
  return context;
}
