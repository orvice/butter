/* eslint-disable react-refresh/only-export-components */

import { createContext, useContext, useEffect, useMemo, useState, type ReactNode } from "react";
import { COLOR_THEME_KEY } from "@/lib/constants";

export type ColorThemeId =
  | "material-blue"
  | "material-green"
  | "material-red"
  | "material-orange"
  | "material-cyan"
  | "material-indigo"
  | "material-teal"
  | "butter"
  | "mint"
  | "blueberry"
  | "rose"
  | "slate";

export type ColorTheme = {
  id: ColorThemeId;
  name: string;
  description: string;
  swatches: [string, string, string];
};

const COLOR_THEMES: ColorTheme[] = [
  {
    id: "material-blue",
    name: "Material Blue",
    description: "Material Admin signature azure",
    swatches: ["#1e91ff", "#10b981", "#e0f2fe"],
  },
  {
    id: "material-green",
    name: "Material Green",
    description: "Material Admin emerald accent",
    swatches: ["#10b981", "#1e91ff", "#d1fae5"],
  },
  {
    id: "material-red",
    name: "Material Red",
    description: "Material Admin coral accent",
    swatches: ["#ff6b68", "#1e91ff", "#fee2e2"],
  },
  {
    id: "material-orange",
    name: "Material Orange",
    description: "Material Admin amber accent",
    swatches: ["#fea84c", "#1e91ff", "#ffedd5"],
  },
  {
    id: "material-cyan",
    name: "Material Cyan",
    description: "Material Admin cyan accent",
    swatches: ["#00bcd4", "#1e91ff", "#cffafe"],
  },
  {
    id: "material-indigo",
    name: "Material Indigo",
    description: "Material Admin indigo accent",
    swatches: ["#5c6bc0", "#1e91ff", "#e0e7ff"],
  },
  {
    id: "material-teal",
    name: "Material Teal",
    description: "Material Admin teal accent",
    swatches: ["#39bbb0", "#1e91ff", "#ccfbf1"],
  },
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
];

const DEFAULT_COLOR_THEME: ColorThemeId = "material-blue";
const LEGACY_DEFAULT_THEME: ColorThemeId = "butter";

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
  if (typeof window === "undefined") return DEFAULT_COLOR_THEME;
  try {
    const storedTheme = window.localStorage.getItem(COLOR_THEME_KEY);
    if (storedTheme === LEGACY_DEFAULT_THEME) {
      return DEFAULT_COLOR_THEME;
    }
    return isColorThemeId(storedTheme) ? storedTheme : DEFAULT_COLOR_THEME;
  } catch {
    return DEFAULT_COLOR_THEME;
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
