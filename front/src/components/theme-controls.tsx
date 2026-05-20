import { useTheme } from "next-themes";
import { Moon, Palette, Sun } from "lucide-react";
import { Button } from "@/components/ui/button";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "@/components/ui/select";
import { useColorTheme, type ColorThemeId } from "@/hooks/use-color-theme";

export function ThemeControls({ className = "" }: { className?: string }) {
  const { resolvedTheme, theme, setTheme } = useTheme();
  const { colorTheme, setColorTheme, themes } = useColorTheme();
  const isDark = (theme === "system" ? resolvedTheme : theme) === "dark";
  const selectedTheme = themes.find((item) => item.id === colorTheme) ?? themes[0];

  return (
    <div className={`flex items-center gap-2 ${className}`}>
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
        onClick={() => setTheme(isDark ? "light" : "dark")}
        aria-label={isDark ? "Switch to light mode" : "Switch to dark mode"}
      >
        {isDark ? <Sun className="h-4 w-4" /> : <Moon className="h-4 w-4" />}
      </Button>
    </div>
  );
}
