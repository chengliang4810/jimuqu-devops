"use client";

import { create } from "zustand";
import type { ToolbarPage } from "./view-options-store";

interface ToolbarSearchState {
  searchTerms: Partial<Record<ToolbarPage, string>>;
  setSearchTerm: (page: ToolbarPage, term: string) => void;
  clearSearchTerm: (page: ToolbarPage) => void;
}

export const useToolbarSearchStore = create<ToolbarSearchState>((set) => ({
  searchTerms: {},
  setSearchTerm: (page, term) =>
    set((state) => ({
      searchTerms: {
        ...state.searchTerms,
        [page]: term,
      },
    })),
  clearSearchTerm: (page) =>
    set((state) => ({
      searchTerms: {
        ...state.searchTerms,
        [page]: "",
      },
    })),
}));
