"use client";

import { useEffect, useRef, useState } from "react";
import { AnimatePresence, motion } from "motion/react";
import { useNavStore } from "@/stores";
import { buttonVariants } from "@/components/ui/button";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { cn } from "@/lib/utils";
import {
  ArrowDownAZ,
  ArrowUpAZ,
  Plus,
  RefreshCw,
  Search,
  SlidersHorizontal,
  X,
} from "lucide-react";
import { useToolbarSearchStore } from "./search-store";
import {
  type HostFilter,
  type NotificationFilter,
  type ProjectFilter,
  type ToolbarPage,
  type ToolbarSortOrder,
  useToolbarViewOptionsStore,
} from "./view-options-store";

type FilterOption<T extends string> = {
  value: T;
  label: string;
};

type SortOption = FilterOption<ToolbarSortOrder>;

const SORT_OPTIONS: SortOption[] = [
  { value: "manual", label: "手动排序" },
  { value: "name-asc", label: "名称 A-Z" },
  { value: "name-desc", label: "名称 Z-A" },
];

const HOST_FILTER_OPTIONS: FilterOption<HostFilter>[] = [
  { value: "all", label: "全部主机" },
  { value: "port-22", label: "22 端口" },
  { value: "custom-port", label: "自定义端口" },
];

const PROJECT_FILTER_OPTIONS: FilterOption<ProjectFilter>[] = [
  { value: "all", label: "全部项目" },
  { value: "none", label: "免鉴权" },
  { value: "username", label: "用户名密码" },
  { value: "token", label: "Token" },
  { value: "ssh", label: "SSH 私钥" },
];

const NOTIFICATION_FILTER_OPTIONS: FilterOption<NotificationFilter>[] = [
  { value: "all", label: "全部渠道" },
  { value: "default", label: "默认渠道" },
  { value: "webhook", label: "Webhook" },
  { value: "dingtalk", label: "钉钉" },
  { value: "wechat", label: "企业微信" },
  { value: "feishu", label: "飞书" },
];

function getPageMeta(page: ToolbarPage) {
  switch (page) {
    case "hosts":
      return {
        createLabel: "新增主机",
        searchPlaceholder: "主机/IP",
      };
    case "projects":
      return {
        createLabel: "新增项目",
        searchPlaceholder: "项目/分支/仓库",
      };
    case "notifications":
      return {
        createLabel: "新增渠道",
        searchPlaceholder: "渠道名称",
      };
    case "logs":
      return {
        createLabel: "",
        searchPlaceholder: "项目/分支",
      };
  }
}

function IconActionButton({
  label,
  onClick,
  children,
  active = false,
  className,
  disabled = false,
}: {
  label: string;
  onClick?: () => void;
  children: React.ReactNode;
  active?: boolean;
  className?: string;
  disabled?: boolean;
}) {
  return (
    <Tooltip>
      <TooltipTrigger asChild>
        <button
          type="button"
          aria-label={label}
          disabled={disabled}
          onClick={onClick}
          className={buttonVariants({
            variant: "ghost",
            size: "icon",
            className: cn(
              "h-9 w-9 rounded-xl transition-none hover:bg-transparent",
              active ? "text-foreground" : "text-muted-foreground hover:text-foreground",
              className
            ),
          })}
        >
          {children}
        </button>
      </TooltipTrigger>
      <TooltipContent>{label}</TooltipContent>
    </Tooltip>
  );
}

function QueryPanel({
  page,
  onClose,
}: {
  page: ToolbarPage;
  onClose: () => void;
}) {
  const sortOrder = useToolbarViewOptionsStore((state) => state.getSortOrder(page));
  const setSortOrder = useToolbarViewOptionsStore((state) => state.setSortOrder);
  const hostFilter = useToolbarViewOptionsStore((state) => state.hostFilter);
  const projectFilter = useToolbarViewOptionsStore((state) => state.projectFilter);
  const notificationFilter = useToolbarViewOptionsStore((state) => state.notificationFilter);
  const setHostFilter = useToolbarViewOptionsStore((state) => state.setHostFilter);
  const setProjectFilter = useToolbarViewOptionsStore((state) => state.setProjectFilter);
  const setNotificationFilter = useToolbarViewOptionsStore((state) => state.setNotificationFilter);

  const filterOptions =
    page === "hosts"
      ? HOST_FILTER_OPTIONS
      : page === "projects"
        ? PROJECT_FILTER_OPTIONS
        : NOTIFICATION_FILTER_OPTIONS;

  const activeFilter =
    page === "hosts"
      ? hostFilter
      : page === "projects"
        ? projectFilter
        : notificationFilter;

  const handleFilterChange = (value: string) => {
    if (page === "hosts") {
      setHostFilter(value as HostFilter);
      return;
    }
    if (page === "projects") {
      setProjectFilter(value as ProjectFilter);
      return;
    }
    setNotificationFilter(value as NotificationFilter);
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: -8, scale: 0.96 }}
      animate={{ opacity: 1, y: 0, scale: 1 }}
      exit={{ opacity: 0, y: -8, scale: 0.96 }}
      transition={{ duration: 0.16 }}
      className="absolute right-0 top-12 z-30 w-72 rounded-2xl border border-border/70 bg-background/95 p-3 shadow-2xl backdrop-blur"
    >
      <div className="mb-3 flex items-center justify-between">
        <div>
          <p className="text-sm font-semibold text-foreground">查询条件</p>
          <p className="text-xs text-muted-foreground">切换筛选或查看顺序</p>
        </div>
        <button
          type="button"
          onClick={onClose}
          className="rounded-lg p-1 text-muted-foreground transition-colors hover:text-foreground"
          aria-label="关闭查询面板"
        >
          <X className="h-4 w-4" />
        </button>
      </div>

      <div className="space-y-3">
        <div className="space-y-2">
          <p className="text-xs font-medium text-muted-foreground">显示顺序</p>
          <div className="grid grid-cols-2 gap-2">
            {SORT_OPTIONS.map((option) => (
              <button
                key={option.value}
                type="button"
                onClick={() => setSortOrder(page, option.value)}
                className={cn(
                  "inline-flex h-8 items-center justify-center gap-1.5 rounded-lg border px-2 text-xs font-medium transition-colors",
                  sortOrder === option.value
                    ? "border-primary/40 bg-primary text-primary-foreground"
                    : "border-border bg-muted/20 text-foreground hover:bg-muted/40"
                )}
              >
                {option.value === "name-desc" ? (
                  <ArrowDownAZ className="h-3.5 w-3.5" />
                ) : (
                  <ArrowUpAZ className="h-3.5 w-3.5" />
                )}
                {option.label}
              </button>
            ))}
          </div>
        </div>

        <div className="space-y-2">
          <p className="text-xs font-medium text-muted-foreground">筛选范围</p>
          <div className="grid gap-2">
            {filterOptions.map((option) => (
              <button
                key={option.value}
                type="button"
                onClick={() => handleFilterChange(option.value)}
                className={cn(
                  "inline-flex h-8 items-center rounded-lg border px-3 text-left text-xs font-medium transition-colors",
                  activeFilter === option.value
                    ? "border-primary/40 bg-primary text-primary-foreground"
                    : "border-border bg-muted/20 text-foreground hover:bg-muted/40"
                )}
              >
                {option.label}
              </button>
            ))}
          </div>
        </div>
      </div>
    </motion.div>
  );
}

function HostsToolbar() {
  const handleAdd = () => {
    window.dispatchEvent(new CustomEvent("open-host-dialog", { detail: { mode: "create" } }));
  };

  return <ToolbarActions page="hosts" onCreate={handleAdd} />;
}

function ProjectsToolbar() {
  const handleAdd = () => {
    window.dispatchEvent(new CustomEvent("open-project-dialog", { detail: { mode: "create" } }));
  };

  return <ToolbarActions page="projects" onCreate={handleAdd} />;
}

function LogsToolbar() {
  const [loading, setLoading] = useState(false);
  const meta = getPageMeta("logs");
  const searchTerm = useToolbarSearchStore((state) => state.searchTerms.logs || "");
  const setSearchTerm = useToolbarSearchStore((state) => state.setSearchTerm);
  const clearSearchTerm = useToolbarSearchStore((state) => state.clearSearchTerm);
  const [expandedSearch, setExpandedSearch] = useState(false);

  const handleRefresh = () => {
    setLoading(true);
    window.dispatchEvent(new CustomEvent("refresh-logs"));
    setTimeout(() => setLoading(false), 1000);
  };

  return (
    <div className="flex items-center gap-2">
      <div className="relative h-9">
        {!expandedSearch ? (
          <IconActionButton
            label="搜索部署记录"
            onClick={() => setExpandedSearch(true)}
            active={searchTerm.trim().length > 0}
          >
            <Search className="h-4 w-4" />
          </IconActionButton>
        ) : (
          <motion.div
            className="flex h-9 items-center gap-2 rounded-xl border border-border/70 bg-background/90 px-3 shadow-sm backdrop-blur"
            transition={{ type: "spring", stiffness: 420, damping: 30 }}
          >
            <Search className="h-4 w-4 shrink-0 text-muted-foreground" />
            <input
              type="text"
              value={searchTerm}
              autoFocus
              onChange={(event) => setSearchTerm("logs", event.target.value)}
              placeholder={meta.searchPlaceholder}
              className="w-28 bg-transparent text-sm outline-none placeholder:text-muted-foreground md:w-40"
            />
            <button
              type="button"
              aria-label="清空搜索"
              onClick={() => {
                clearSearchTerm("logs");
                setExpandedSearch(false);
              }}
              className="rounded-md p-0.5 text-muted-foreground transition-colors hover:text-foreground"
            >
              <X className="h-3.5 w-3.5" />
            </button>
          </motion.div>
        )}
      </div>

      <IconActionButton
        label="刷新记录"
        onClick={handleRefresh}
        disabled={loading}
        className="border border-border/70 bg-background/80 text-foreground shadow-sm hover:bg-background"
      >
        <RefreshCw className={cn("h-4 w-4", loading && "animate-spin")} />
      </IconActionButton>
    </div>
  );
}

function NotificationsToolbar() {
  const handleAdd = () => {
    window.dispatchEvent(new CustomEvent("open-notify-dialog", { detail: { mode: "create" } }));
  };

  return <ToolbarActions page="notifications" onCreate={handleAdd} />;
}

function ToolbarActions({
  page,
  onCreate,
}: {
  page: ToolbarPage;
  onCreate: () => void;
}) {
  const panelRef = useRef<HTMLDivElement | null>(null);
  const meta = getPageMeta(page);
  const searchTerm = useToolbarSearchStore((state) => state.searchTerms[page] || "");
  const setSearchTerm = useToolbarSearchStore((state) => state.setSearchTerm);
  const clearSearchTerm = useToolbarSearchStore((state) => state.clearSearchTerm);
  const [expandedSearchItem, setExpandedSearchItem] = useState<ToolbarPage | null>(null);
  const [openPanelItem, setOpenPanelItem] = useState<ToolbarPage | null>(null);
  const searchExpanded = expandedSearchItem === page;
  const panelOpen = openPanelItem === page;

  useEffect(() => {
    if (!panelOpen) {
      return;
    }

    const handlePointerDown = (event: MouseEvent) => {
      if (!panelRef.current?.contains(event.target as Node)) {
        setOpenPanelItem(null);
      }
    };

    document.addEventListener("mousedown", handlePointerDown);
    return () => document.removeEventListener("mousedown", handlePointerDown);
  }, [panelOpen]);

  return (
    <div ref={panelRef} className="relative flex items-center gap-2">
      <div className="relative h-9">
        {!searchExpanded ? (
          <IconActionButton
            label={`搜索${meta.createLabel.replace("新增", "")}`}
            onClick={() => setExpandedSearchItem(page)}
            active={searchTerm.trim().length > 0}
          >
            <Search className="h-4 w-4" />
          </IconActionButton>
        ) : (
          <motion.div
            layoutId={`toolbar-search-${page}`}
            className="flex h-9 items-center gap-2 rounded-xl border border-border/70 bg-background/90 px-3 shadow-sm backdrop-blur"
            transition={{ type: "spring", stiffness: 420, damping: 30 }}
          >
            <Search className="h-4 w-4 shrink-0 text-muted-foreground" />
            <input
              type="text"
              value={searchTerm}
              autoFocus
              onChange={(event) => setSearchTerm(page, event.target.value)}
              placeholder={meta.searchPlaceholder}
              className="w-28 bg-transparent text-sm outline-none placeholder:text-muted-foreground md:w-40"
            />
            <button
              type="button"
              aria-label="清空搜索"
              onClick={() => {
                clearSearchTerm(page);
                setExpandedSearchItem(null);
              }}
              className="rounded-md p-0.5 text-muted-foreground transition-colors hover:text-foreground"
            >
              <X className="h-3.5 w-3.5" />
            </button>
          </motion.div>
        )}
      </div>

      <IconActionButton
        label="查询条件"
        onClick={() => setOpenPanelItem((current) => (current === page ? null : page))}
        active={panelOpen}
      >
        <SlidersHorizontal className="h-4 w-4" />
      </IconActionButton>

      <IconActionButton label={meta.createLabel} onClick={onCreate}>
        <Plus className="h-4 w-4" />
      </IconActionButton>

      <AnimatePresence>
        {panelOpen ? <QueryPanel page={page} onClose={() => setOpenPanelItem(null)} /> : null}
      </AnimatePresence>
    </div>
  );
}

export function Toolbar() {
  const { activeView } = useNavStore();

  return (
    <AnimatePresence mode="wait">
      <motion.div
        key={activeView}
        initial={{ opacity: 0, scale: 0.9 }}
        animate={{ opacity: 1, scale: 1 }}
        exit={{ opacity: 0, scale: 0.9 }}
        transition={{ duration: 0.2 }}
      >
        {activeView === "hosts" && <HostsToolbar />}
        {activeView === "projects" && <ProjectsToolbar />}
        {activeView === "logs" && <LogsToolbar />}
        {activeView === "notifications" && <NotificationsToolbar />}
        {activeView === "home" && null}
      </motion.div>
    </AnimatePresence>
  );
}
