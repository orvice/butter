import { useTheme } from "next-themes";
import { Maximize2, Minimize2, Moon, Palette, Sun } from "lucide-react";
import { Button } from "@/components/ui/button";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "@/components/ui/dropdown-menu";
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
      <DropdownMenu>
        <DropdownMenuTrigger asChild>
          <Button
            type="button"
            variant="ghost"
            size="icon"
            aria-label="Appearance settings"
            className={cn("shrink-0", className, triggerClassName)}
          >
            <Palette className="h-4 w-4" />
          </Button>
        </DropdownMenuTrigger>
        <DropdownMenuContent align="end" sideOffset={8} className="w-64 p-2">
          <DropdownMenuLabel className="px-1 pb-1">Appearance</DropdownMenuLabel>
          <DropdownMenuRadioGroup
            value={colorTheme}
            onValueChange={(value) => setColorTheme(value as ColorThemeId)}
          >
            {themes.map((item) => (
              <DropdownMenuRadioItem
                key={item.id}
                value={item.id}
                className="gap-2 py-1.5 pr-8"
              >
                <span className="flex shrink-0 overflow-hidden rounded-full border">
                  {item.swatches.map((swatch) => (
                    <span key={swatch} className="h-3.5 w-3.5" style={{ backgroundColor: swatch }} />
                  ))}
                </span>
                <span className="flex min-w-0 flex-col leading-tight">
                  <span>{item.name}</span>
                  <span className="truncate text-[10px] text-muted-foreground">{item.description}</span>
                </span>
              </DropdownMenuRadioItem>
            ))}
          </DropdownMenuRadioGroup>
          <DropdownMenuSeparator className="my-2" />
          <div className="grid grid-cols-2 gap-1">
            <DropdownMenuItem onClick={toggleDensity} className="justify-start">
              {isCompact ? <Maximize2 className="h-3.5 w-3.5" /> : <Minimize2 className="h-3.5 w-3.5" />}
              {isCompact ? "Comfort" : "Compact"}
            </DropdownMenuItem>
            <DropdownMenuItem onClick={() => setTheme(isDark ? "light" : "dark")} className="justify-start">
              {isDark ? <Sun className="h-3.5 w-3.5" /> : <Moon className="h-3.5 w-3.5" />}
              {isDark ? "Light" : "Dark"}
            </DropdownMenuItem>
          </div>
        </DropdownMenuContent>
      </DropdownMenu>
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
