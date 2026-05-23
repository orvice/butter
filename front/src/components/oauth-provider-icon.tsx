import type { SVGProps } from "react";

interface Props extends SVGProps<SVGSVGElement> {
  provider: string;
}

// OAuthProviderIcon renders a brand mark for a given provider key. lucide-react
// doesn't ship branded logos, so we inline minimal SVGs. Unknown providers
// fall back to a generic "shield" mark.
export function OAuthProviderIcon({ provider, className = "h-4 w-4", ...rest }: Props) {
  switch (provider) {
    case "github":
      return (
        <svg viewBox="0 0 24 24" aria-hidden="true" className={className} {...rest}>
          <path
            fill="currentColor"
            d="M12 .5C5.65.5.5 5.65.5 12c0 5.08 3.29 9.39 7.86 10.91.57.11.78-.25.78-.55v-2.07c-3.2.7-3.87-1.37-3.87-1.37-.52-1.32-1.27-1.67-1.27-1.67-1.04-.71.08-.7.08-.7 1.15.08 1.76 1.18 1.76 1.18 1.02 1.75 2.69 1.25 3.34.96.1-.74.4-1.25.72-1.54-2.55-.29-5.23-1.28-5.23-5.7 0-1.26.45-2.29 1.18-3.1-.12-.29-.51-1.47.11-3.06 0 0 .97-.31 3.18 1.18a11.05 11.05 0 0 1 5.79 0c2.2-1.49 3.17-1.18 3.17-1.18.62 1.59.23 2.77.11 3.06.74.81 1.18 1.84 1.18 3.1 0 4.43-2.69 5.41-5.25 5.69.41.35.78 1.05.78 2.12v3.14c0 .31.21.67.79.55A11.51 11.51 0 0 0 23.5 12C23.5 5.65 18.35.5 12 .5Z"
          />
        </svg>
      );
    case "google":
      // Official Google "G" mark, four-color, simplified to inline SVG.
      return (
        <svg viewBox="0 0 24 24" aria-hidden="true" className={className} {...rest}>
          <path
            fill="#4285F4"
            d="M23.49 12.27c0-.79-.07-1.54-.19-2.27H12v4.51h6.44c-.28 1.4-1.07 2.59-2.27 3.39v2.77h3.66c2.14-1.97 3.66-4.96 3.66-8.4Z"
          />
          <path
            fill="#34A853"
            d="M12 24c3.24 0 5.95-1.08 7.93-2.91l-3.66-2.77c-1.02.69-2.34 1.09-4.27 1.09-3.27 0-6.04-2.2-7.04-5.18H1.18v3.26C3.16 21.3 7.27 24 12 24Z"
          />
          <path
            fill="#FBBC05"
            d="M4.96 14.23A7.18 7.18 0 0 1 4.59 12c0-.78.13-1.53.36-2.23V6.51H1.18A12 12 0 0 0 0 12c0 1.94.46 3.78 1.18 5.49l3.78-3.26Z"
          />
          <path
            fill="#EA4335"
            d="M12 4.75c1.77 0 3.36.61 4.61 1.8l3.25-3.25C17.95 1.19 15.24 0 12 0 7.27 0 3.16 2.7 1.18 6.51l3.78 3.26c1-2.98 3.77-5.02 7.04-5.02Z"
          />
        </svg>
      );
    default:
      return (
        <svg viewBox="0 0 24 24" aria-hidden="true" className={className} {...rest}>
          <path
            fill="currentColor"
            d="M12 2 4 5v6c0 4.97 3.32 9.4 8 11 4.68-1.6 8-6.03 8-11V5l-8-3Zm0 2.18 6 2.25V11c0 3.93-2.5 7.6-6 9-3.5-1.4-6-5.07-6-9V6.43l6-2.25Z"
          />
        </svg>
      );
  }
}
