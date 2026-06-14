import { useTheme } from "next-themes";
import { Check, Maximize2, Minimize2, Moon, Palette, Sun } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { useColorTheme, type ColorThemeId } from "@/hooks/use-color-theme";
import { useLayoutDensity } from "@/hooks/use-layout-density";
import { cn } from "@/lib/utils";

type ThemeControlsProps = {
  className?: string;
  mode?: "inline" | "menu";
  triggerClassName?: string;
};

export function ThemeControls({ className = "", mode = "inline", triggerClassName }: ThemeControlsProps) {
  const { resolvedTheme, theme, setTheme } = useTheme();
  const { colorTheme, setColorTheme, themes } = useColorTheme();
  const { isCompact, toggleDensity } = useLayoutDensity();
  const isDark = (theme === "system" ? resolvedTheme : theme) === "dark";
  const selectedTheme = themes.find((item) => item.id === colorTheme) ?? themes[0];

  if (mode === "menu") {
    return (
      <div className={cn("group/theme-menu relative", className)}>
        <button
          type="button"
          aria-label="Appearance settings"
          aria-haspopup="menu"
          className={cn(
            "inline-flex size-9 shrink-0 items-center justify-center rounded-md border border-transparent text-sm font-medium transition-all outline-none focus-visible:border-ring focus-visible:ring-3 focus-visible:ring-ring/20 active:translate-y-px",
            triggerClassName,
          )}
        >
          <Palette className="h-4 w-4" />
        </button>
        <div
          role="menu"
          className="invisible absolute right-0 top-full z-50 mt-2 w-64 rounded-md border border-border bg-popover p-2 text-popover-foreground opacity-0 shadow-dropdown transition duration-150 group-hover/theme-menu:visible group-hover/theme-menu:opacity-100 group-focus-within/theme-menu:visible group-focus-within/theme-menu:opacity-100"
        >
          <div className="px-1 pb-1 text-xs font-medium text-muted-foreground">Appearance</div>
          <div className="space-y-0.5">
            {themes.map((item) => (
              <button
                key={item.id}
                type="button"
                role="menuitemradio"
                aria-checked={item.id === colorTheme}
                className={cn(
                  "flex w-full items-center gap-2 rounded-md px-1.5 py-1.5 text-left text-sm outline-none hover:bg-accent hover:text-accent-foreground focus-visible:bg-accent focus-visible:text-accent-foreground",
                  item.id === colorTheme && "bg-accent/60 text-accent-foreground",
                )}
                onClick={() => setColorTheme(item.id)}
              >
                <span className="flex min-w-0 items-center gap-2">
                  <span className="flex shrink-0 overflow-hidden rounded-full border">
                    {item.swatches.map((swatch) => (
                      <span key={swatch} className="h-3.5 w-3.5" style={{ backgroundColor: swatch }} />
                    ))}
                  </span>
                  <span className="flex min-w-0 flex-col leading-tight">
                    <span>{item.name}</span>
                    <span className="truncate text-[10px] text-muted-foreground">{item.description}</span>
                  </span>
                </span>
                {item.id === colorTheme ? <Check className="ml-auto h-3.5 w-3.5 shrink-0" /> : null}
              </button>
            ))}
          </div>
          <div className="my-2 h-px bg-border" />
          <div className="grid grid-cols-2 gap-1">
            <Button
              variant="ghost"
              size="sm"
              className="justify-start"
              onClick={() => {
                toggleDensity();
              }}
            >
              {isCompact ? <Maximize2 className="h-3.5 w-3.5" /> : <Minimize2 className="h-3.5 w-3.5" />}
              {isCompact ? "Comfort" : "Compact"}
            </Button>
            <Button
              variant="ghost"
              size="sm"
              className="justify-start"
              onClick={() => {
                setTheme(isDark ? "light" : "dark");
              }}
            >
              {isDark ? <Sun className="h-3.5 w-3.5" /> : <Moon className="h-3.5 w-3.5" />}
              {isDark ? "Light" : "Dark"}
            </Button>
          </div>
        </div>
      </div>
    );
  }

  return (
    <div className={cn("flex items-center gap-2", className)}>
      <Select value={colorTheme} onValueChange={(value) => setColorTheme(value as ColorThemeId)}>
        <SelectTrigger size="sm" className="w-32" aria-label="Color theme">
          <SelectValue>
            <span className="flex items-center gap-1.5">
              <Palette className="h-3.5 w-3.5 text-muted-foreground" />
              {selectedTheme.name}
            </span>
          </SelectValue>
        </SelectTrigger>
        <SelectContent align="end" className="w-56">
          {themes.map((item) => (
            <SelectItem key={item.id} value={item.id}>
              <span className="flex min-w-0 items-center gap-2">
                <span className="flex shrink-0 overflow-hidden rounded-full border">
                  {item.swatches.map((swatch) => (
                    <span key={swatch} className="h-4 w-4" style={{ backgroundColor: swatch }} />
                  ))}
                </span>
                <span className="flex min-w-0 flex-col leading-tight">
                  <span>{item.name}</span>
                  <span className="truncate text-[10px] text-muted-foreground">{item.description}</span>
                </span>
              </span>
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      <Button
        variant="ghost"
        size="icon"
        onClick={toggleDensity}
        aria-label={isCompact ? "Use comfortable layout" : "Use compact layout"}
        title={isCompact ? "Comfortable layout" : "Compact layout"}
      >
        {isCompact ? <Maximize2 className="h-4 w-4" /> : <Minimize2 className="h-4 w-4" />}
      </Button>
      <Button
        variant="ghost"
        size="icon"
        onClick={() => setTheme(isDark ? "light" : "dark")}
        aria-label={isDark ? "Switch to light mode" : "Switch to dark mode"}
      >
        {isDark ? <Sun className="h-4 w-4" /> : <Moon className="h-4 w-4" />}
      </Button>
    </div>
  );
}
