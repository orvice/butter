export const MAX_TITLE_CODE_POINTS = 100;

/** Trims and collapses a manual title to a single line, mirroring the backend. */
export function normalizeTitleInput(raw: string): string {
  return raw.replace(/\r\n|\r|\n/g, " ").trim();
}

export function titleCodePointCount(s: string): number {
  return [...s].length;
}
