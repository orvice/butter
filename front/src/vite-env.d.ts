/// <reference types="vite/client" />

declare module "*.css";

declare module "next-themes" {
  import type { ComponentType, ReactNode } from "react";

  export interface ThemeProviderProps {
    attribute?: string;
    children?: ReactNode;
    defaultTheme?: string;
    enableSystem?: boolean;
  }

  export interface UseThemeResult {
    theme?: string;
    setTheme: (theme: string) => void;
  }

  export const ThemeProvider: ComponentType<ThemeProviderProps>;
  export function useTheme(): UseThemeResult;
}

declare module "class-variance-authority" {
  export function cva(base?: string, config?: unknown): (props?: Record<string, unknown>) => string;
  export type VariantProps<T extends (...args: never[]) => unknown> = T extends (props?: infer P) => unknown ? P : never;
}
