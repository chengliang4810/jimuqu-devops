const normalizeLineEndings = (input: string): string =>
  input.replace(/\r\n/g, "\n").replace(/\r/g, "\n");

export function normalizeBuildImageInput(value?: string | null): string {
  if (!value) {
    return "";
  }

  const cleaned = normalizeLineEndings(value);
  const trimmed = cleaned.trim();

  return trimmed;
}
