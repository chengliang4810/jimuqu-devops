import { describe, expect, it } from "vitest";

import { normalizeBuildImageInput } from "./image-input";

describe("normalizeBuildImageInput", () => {
  it("trims spaces and outer line breaks", () => {
    expect(normalizeBuildImageInput(" \nnode:20\r\n ")).toBe("node:20");
  });

  it("returns empty string for whitespace-only input", () => {
    expect(normalizeBuildImageInput(" \r\n ")).toBe("");
  });

  it("returns empty string for undefined or null input", () => {
    expect(normalizeBuildImageInput(undefined)).toBe("");
    expect(normalizeBuildImageInput(null)).toBe("");
  });

  it("normalizes stray carriage returns", () => {
    expect(normalizeBuildImageInput("node\r")).toBe("node");
  });
});
