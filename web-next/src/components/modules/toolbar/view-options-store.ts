"use client";

import { create } from "zustand";
import { persist } from "zustand/middleware";

export const TOOLBAR_PAGES = ["hosts", "projects", "notifications", "logs"] as const;

export type ToolbarPage = (typeof TOOLBAR_PAGES)[number];
export type ToolbarSortOrder = "manual" | "name-asc" | "name-desc";
export type HostFilter = "all" | "port-22" | "custom-port";
export type ProjectFilter = "all" | "none" | "username" | "token" | "ssh";
export type NotificationFilter =
  | "all"
  | "default"
  | "webhook"
  | "dingtalk"
  | "wechat"
  | "feishu";

interface ToolbarViewOptionsState {
  sortOrders: Partial<Record<ToolbarPage, ToolbarSortOrder>>;
  hostFilter: HostFilter;
  projectFilter: ProjectFilter;
  notificationFilter: NotificationFilter;
  getSortOrder: (page: ToolbarPage) => ToolbarSortOrder;
  setSortOrder: (page: ToolbarPage, value: ToolbarSortOrder) => void;
  setHostFilter: (value: HostFilter) => void;
  setProjectFilter: (value: ProjectFilter) => void;
  setNotificationFilter: (value: NotificationFilter) => void;
}

export const useToolbarViewOptionsStore = create<ToolbarViewOptionsState>()(
  persist(
    (set, get) => ({
      sortOrders: {},
      hostFilter: "all",
      projectFilter: "all",
      notificationFilter: "all",

      getSortOrder: (page) => get().sortOrders[page] || "manual",
      setSortOrder: (page, value) =>
        set((state) => ({
          sortOrders: {
            ...state.sortOrders,
            [page]: value,
          },
        })),
      setHostFilter: (value) => set({ hostFilter: value }),
      setProjectFilter: (value) => set({ projectFilter: value }),
      setNotificationFilter: (value) => set({ notificationFilter: value }),
    }),
    {
      name: "toolbar-view-options-storage",
      partialize: (state) => ({
        sortOrders: state.sortOrders,
        hostFilter: state.hostFilter,
        projectFilter: state.projectFilter,
        notificationFilter: state.notificationFilter,
      }),
    }
  )
);
