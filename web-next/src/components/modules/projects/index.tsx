"use client";

import { useDeferredValue, useEffect, useRef, useState } from "react";
import {
  closestCenter,
  DndContext,
  DragOverlay,
  KeyboardSensor,
  PointerSensor,
  useSensor,
  useSensors,
  type DragEndEvent,
  type DragStartEvent,
} from "@dnd-kit/core";
import { restrictToParentElement } from "@dnd-kit/modifiers";
import {
  arrayMove,
  rectSortingStrategy,
  SortableContext,
  sortableKeyboardCoordinates,
  useSortable,
} from "@dnd-kit/sortable";
import { CSS } from "@dnd-kit/utilities";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Textarea } from "@/components/ui/textarea";
import { Card, CardContent } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { projectApi, hostApi, notifyApi } from "@/api/client";
import type {
  Project,
  Host,
  NotifyChannel,
  DeployConfig,
  ProjectDetail,
  ImageSearchItem,
} from "@/types";
import { toast } from "sonner";
import { Copy, CopyPlus, GripVertical, LoaderCircle, Pencil, Play, Search, Trash2, X } from "lucide-react";
import { motion, AnimatePresence } from "motion/react";
import { normalizeBuildImageInput } from "@/lib/image-input";
import { cn, formatMultilineValue, parseMultilineInput } from "@/lib/utils";
import { useNavStore } from "@/stores";
import { useToolbarSearchStore } from "@/components/modules/toolbar/search-store";
import { useToolbarViewOptionsStore } from "@/components/modules/toolbar/view-options-store";

type GitAuthType = Project["git_auth_type"];

type DeployConfigFormState = {
  host_id: string;
  build_image: string;
  build_commands: string;
  artifact_filter_mode: "include" | "exclude";
  artifact_rules: string;
  remote_save_dir: string;
  remote_deploy_dir: string;
  pre_deploy_commands: string;
  post_deploy_commands: string;
  version_count: number;
  timeout_minutes: number;
  notification_channel_id: string;
};

type ProjectFormState = {
  name: string;
  branch: string;
  repo_url: string;
  description: string;
  git_auth_type: GitAuthType;
  git_username: string;
  git_password: string;
  git_ssh_key: string;
  has_existing_git_auth: boolean;
  has_existing_git_password: boolean;
  has_existing_git_ssh_key: boolean;
  original_git_auth_type: GitAuthType;
  deploy_config: DeployConfigFormState;
};

// 全局缓存
let projectsCache: Project[] | null = null;
let hostsCache: Host[] | null = null;
let channelsCache: NotifyChannel[] | null = null;

const DEFAULT_NOTIFICATION_CHANNEL = "__default__";

const defaultDeployConfig: DeployConfigFormState = {
  host_id: "",
  build_image: "node:20",
  build_commands: "",
  artifact_filter_mode: "include",
  artifact_rules: "",
  remote_save_dir: "/data/jimuqu/projects",
  remote_deploy_dir: "",
  pre_deploy_commands: "",
  post_deploy_commands: "",
  version_count: 5,
  timeout_minutes: 30,
  notification_channel_id: DEFAULT_NOTIFICATION_CHANNEL,
};

function createDefaultFormData(): ProjectFormState {
  return {
    name: "",
    branch: "main",
    repo_url: "",
    description: "",
    git_auth_type: "none",
    git_username: "",
    git_password: "",
    git_ssh_key: "",
    has_existing_git_auth: false,
    has_existing_git_password: false,
    has_existing_git_ssh_key: false,
    original_git_auth_type: "none",
    deploy_config: { ...defaultDeployConfig },
  };
}

function mapDeployConfigToForm(config?: DeployConfig | null): DeployConfigFormState {
  if (!config) {
    return { ...defaultDeployConfig };
  }

  return {
    host_id: config.host_id ? String(config.host_id) : "",
    build_image: config.build_image || defaultDeployConfig.build_image,
    build_commands: formatMultilineValue(config.build_commands),
    artifact_filter_mode:
      config.artifact_filter_mode === "exclude" ? "exclude" : "include",
    artifact_rules: formatMultilineValue(config.artifact_rules),
    remote_save_dir: config.remote_save_dir || defaultDeployConfig.remote_save_dir,
    remote_deploy_dir: config.remote_deploy_dir || "",
    pre_deploy_commands: formatMultilineValue(config.pre_deploy_commands),
    post_deploy_commands: formatMultilineValue(config.post_deploy_commands),
    version_count: Math.max(1, config.version_count || defaultDeployConfig.version_count),
    timeout_minutes: Math.max(1, Math.floor((config.timeout_seconds || 1800) / 60)),
    notification_channel_id:
      config.notification_channel_id == null
        ? DEFAULT_NOTIFICATION_CHANNEL
        : String(config.notification_channel_id),
  };
}

function buildProjectPayload(formData: ProjectFormState, isEditing: boolean) {
  const projectPayload: Record<string, unknown> = {
    name: formData.name.trim(),
    branch: formData.branch.trim(),
    repo_url: formData.repo_url.trim(),
    description: formData.description.trim(),
    git_auth_type: formData.git_auth_type,
  };

  const canReuseUsernamePassword =
    isEditing &&
    formData.has_existing_git_auth &&
    (formData.original_git_auth_type === "username" || formData.original_git_auth_type === "token") &&
    (formData.git_auth_type === "username" || formData.git_auth_type === "token");

  const canReuseSSHKey =
    isEditing &&
    formData.has_existing_git_auth &&
    formData.original_git_auth_type === "ssh" &&
    formData.git_auth_type === "ssh";

  if (formData.git_auth_type === "username" || formData.git_auth_type === "token") {
    const gitUsername = formData.git_username.trim();
    const gitPassword = formData.git_password.trim();
    if (gitUsername !== "" || !canReuseUsernamePassword) {
      projectPayload.git_username = gitUsername;
    }
    if (gitPassword !== "" || !canReuseUsernamePassword) {
      projectPayload.git_password = gitPassword;
    }
  }

  if (formData.git_auth_type === "ssh") {
    const gitSSHKey = formData.git_ssh_key.trim();
    if (gitSSHKey !== "" || !canReuseSSHKey) {
      projectPayload.git_ssh_key = gitSSHKey;
    }
  }

  return {
    ...projectPayload,
    deploy_config: buildDeployConfigPayload(formData.deploy_config),
  };
}

function getProjectValidationError(formData: ProjectFormState, isEditing: boolean): string | null {
  if (!formData.name.trim()) {
    return "项目名称不能为空";
  }
  if (!formData.branch.trim()) {
    return "分支不能为空";
  }
  if (!formData.repo_url.trim()) {
    return "仓库地址不能为空";
  }

  const canReuseUsernamePassword =
    isEditing &&
    formData.has_existing_git_auth &&
    (formData.original_git_auth_type === "username" || formData.original_git_auth_type === "token") &&
    (formData.git_auth_type === "username" || formData.git_auth_type === "token");

  const canReuseSSHKey =
    isEditing &&
    formData.has_existing_git_auth &&
    formData.original_git_auth_type === "ssh" &&
    formData.git_auth_type === "ssh";

  if (formData.git_auth_type === "username" || formData.git_auth_type === "token") {
    if (!canReuseUsernamePassword && !formData.git_username.trim()) {
      return "Git 用户名不能为空";
    }
    if (!canReuseUsernamePassword && !formData.git_password.trim()) {
      return formData.git_auth_type === "token" ? "Git Token 不能为空" : "Git 密码不能为空";
    }
  }

  if (formData.git_auth_type === "ssh" && !canReuseSSHKey && !formData.git_ssh_key.trim()) {
    return "SSH 私钥不能为空";
  }

  return null;
}

function buildDeployConfigPayload(formData: DeployConfigFormState) {
  const artifactRules = parseMultilineInput(formData.artifact_rules);

  return {
    host_id: Number(formData.host_id),
    build_image: normalizeBuildImageInput(formData.build_image),
    build_commands: parseMultilineInput(formData.build_commands),
    artifact_filter_mode: artifactRules.length ? formData.artifact_filter_mode : "none",
    artifact_rules: artifactRules,
    remote_save_dir: formData.remote_save_dir.trim(),
    remote_deploy_dir: formData.remote_deploy_dir.trim(),
    pre_deploy_commands: parseMultilineInput(formData.pre_deploy_commands),
    post_deploy_commands: parseMultilineInput(formData.post_deploy_commands),
    version_count: Math.max(1, formData.version_count),
    timeout_seconds: Math.max(1, formData.timeout_minutes) * 60,
    notification_channel_id:
      formData.notification_channel_id === DEFAULT_NOTIFICATION_CHANNEL
        ? null
        : Number(formData.notification_channel_id),
  };
}

function getDeployConfigValidationError(formData: DeployConfigFormState): string | null {
  if (!formData.host_id) {
    return "请选择目标主机";
  }
  if (!normalizeBuildImageInput(formData.build_image)) {
    return "编译镜像不能为空";
  }
  if (parseMultilineInput(formData.build_commands).length === 0) {
    return "请至少填写一条编译命令";
  }
  if (!formData.remote_save_dir.trim()) {
    return "远程保存目录不能为空";
  }
  if (!formData.remote_deploy_dir.trim()) {
    return "远程部署目录不能为空";
  }
  if (formData.version_count <= 0) {
    return "版本数量必须大于 0";
  }
  if (formData.timeout_minutes <= 0) {
    return "部署超时必须大于 0 分钟";
  }
  return null;
}

type ProjectCardViewProps = {
  project: Project;
  dragHandle?: React.ReactNode;
  isDeleting: boolean;
  isDragging?: boolean;
  isOverlay?: boolean;
  onTrigger: () => void;
  onCopyWebhook: () => void;
  onClone: () => void;
  onEdit: () => void;
  onDeleteRequest: () => void;
  onDeleteCancel: () => void;
  onDeleteConfirm: () => void;
};

function ProjectCardView({
  project,
  dragHandle,
  isDeleting,
  isDragging = false,
  isOverlay = false,
  onTrigger,
  onCopyWebhook,
  onClone,
  onEdit,
  onDeleteRequest,
  onDeleteCancel,
  onDeleteConfirm,
}: ProjectCardViewProps) {
  return (
    <Card
      className={cn(
        "h-full overflow-hidden border-border/70 transition-[opacity,box-shadow,transform] duration-200 will-change-transform",
        isDragging && "opacity-35",
        isOverlay && "rotate-[1deg] scale-[1.02] opacity-100 shadow-2xl ring-2 ring-primary/30"
      )}
    >
      <CardContent className="p-4">
        <header className="relative mb-3 flex items-start justify-between">
          <div className="mr-2 min-w-0 flex-1">
            <Tooltip>
              <TooltipTrigger asChild>
                <h3 className="truncate font-semibold text-foreground">{project.name}</h3>
              </TooltipTrigger>
              <TooltipContent>{project.name}</TooltipContent>
            </Tooltip>
            <p className="text-sm text-muted-foreground">{project.branch}</p>
          </div>
          <div className="flex shrink-0 items-center gap-1">
            {dragHandle}
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-8 w-8"
                  aria-label="触发部署"
                  onClick={onTrigger}
                >
                  <Play className="h-4 w-4" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>触发部署</TooltipContent>
            </Tooltip>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-8 w-8"
                  aria-label="复制 Webhook"
                  onClick={onCopyWebhook}
                >
                  <Copy className="h-4 w-4" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>复制 Webhook</TooltipContent>
            </Tooltip>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-8 w-8"
                  aria-label="复制项目"
                  onClick={onClone}
                >
                  <CopyPlus className="h-4 w-4" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>复制项目</TooltipContent>
            </Tooltip>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-8 w-8"
                  aria-label="编辑"
                  onClick={onEdit}
                >
                  <Pencil className="h-4 w-4" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>编辑</TooltipContent>
            </Tooltip>
            {!isDeleting ? (
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button
                    variant="ghost"
                    size="icon"
                    className="h-8 w-8 text-muted-foreground hover:text-destructive"
                    aria-label="删除"
                    onClick={onDeleteRequest}
                  >
                    <Trash2 className="h-4 w-4" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>删除</TooltipContent>
              </Tooltip>
            ) : null}
          </div>

          <AnimatePresence>
            {isDeleting ? (
              <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                exit={{ opacity: 0 }}
                className="absolute inset-0 flex items-center justify-center gap-2 rounded-lg bg-destructive p-2"
              >
                <button
                  type="button"
                  onClick={onDeleteCancel}
                  className="flex h-7 w-7 items-center justify-center rounded-lg bg-white/20 text-white transition-all hover:bg-white/30 active:scale-95"
                >
                  <X className="h-4 w-4" />
                </button>
                <button
                  type="button"
                  onClick={onDeleteConfirm}
                  className="flex h-7 flex-1 items-center justify-center gap-2 rounded-lg bg-white text-sm font-semibold text-destructive transition-all hover:bg-white/90 active:scale-[0.98]"
                >
                  <Trash2 className="h-3.5 w-3.5" />
                  确认删除
                </button>
              </motion.div>
            ) : null}
          </AnimatePresence>
        </header>

        <p className="truncate text-sm text-muted-foreground">{project.repo_url}</p>
      </CardContent>
    </Card>
  );
}

type SortableProjectCardProps = {
  project: Project;
  isDeleting: boolean;
  onTrigger: () => void;
  onCopyWebhook: () => void;
  onClone: () => void;
  onEdit: () => void;
  onDeleteRequest: () => void;
  onDeleteCancel: () => void;
  onDeleteConfirm: () => void;
};

function SortableProjectCard(props: SortableProjectCardProps) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: props.project.id,
  });

  return (
    <div
      ref={setNodeRef}
      style={{
        transform: CSS.Transform.toString(transform),
        transition,
        zIndex: isDragging ? 20 : undefined,
      }}
      className="h-full will-change-transform"
    >
      <ProjectCardView
        {...props}
        isDragging={isDragging}
        dragHandle={
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                type="button"
                variant="ghost"
                size="icon"
                className={cn(
                  "h-8 w-8 cursor-grab touch-none active:cursor-grabbing",
                  isDragging && "cursor-grabbing"
                )}
                aria-label="拖拽排序"
                {...attributes}
                {...listeners}
              >
                <GripVertical className="h-4 w-4" />
              </Button>
            </TooltipTrigger>
            <TooltipContent>拖拽排序</TooltipContent>
          </Tooltip>
        }
      />
    </div>
  );
}

export function Projects() {
  const { setActiveView, setPendingRunId } = useNavStore();
  const [projects, setProjects] = useState<Project[]>(projectsCache || []);
  const [hosts, setHosts] = useState<Host[]>(hostsCache || []);
  const [channels, setChannels] = useState<NotifyChannel[]>(channelsCache || []);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingProject, setEditingProject] = useState<Project | null>(null);
  const [deletingProject, setDeletingProject] = useState<Project | null>(null);
  const [activeProjectId, setActiveProjectId] = useState<number | null>(null);
  const [activeTab, setActiveTab] = useState("basic");
  const [imageSearchResults, setImageSearchResults] = useState<ImageSearchItem[]>([]);
  const [imageSearchLoading, setImageSearchLoading] = useState(false);
  const [imageSearchError, setImageSearchError] = useState<string | null>(null);
  const [showImageSearchResults, setShowImageSearchResults] = useState(false);
  const loadingRef = useRef(false);
  const imageSearchRequestRef = useRef(0);
  const [formData, setFormData] = useState<ProjectFormState>(createDefaultFormData());
  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: { distance: 2 },
    }),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  );
  const searchTerm = useToolbarSearchStore((state) => state.searchTerms.projects || "");
  const projectFilter = useToolbarViewOptionsStore((state) => state.projectFilter);
  const sortOrder = useToolbarViewOptionsStore((state) => state.getSortOrder("projects"));
  const deferredSearchTerm = useDeferredValue(searchTerm);
  const deferredBuildImage = useDeferredValue(
    normalizeBuildImageInput(formData.deploy_config.build_image)
  );

  const loadData = async (forceRefresh = false) => {
    if (!forceRefresh && projectsCache && hostsCache && channelsCache) {
      setProjects(projectsCache);
      setHosts(hostsCache);
      setChannels(channelsCache);
      return;
    }
    if (loadingRef.current) return;
    loadingRef.current = true;
    try {
      const [projectList, hostList, channelList] = await Promise.all([
        !forceRefresh && projectsCache ? Promise.resolve(projectsCache) : projectApi.list(),
        !forceRefresh && hostsCache ? Promise.resolve(hostsCache) : hostApi.list(),
        !forceRefresh && channelsCache ? Promise.resolve(channelsCache) : notifyApi.list(),
      ]);

      const nextProjects = Array.isArray(projectList) ? projectList : [];
      const nextHosts = Array.isArray(hostList) ? hostList : [];
      const nextChannels = Array.isArray(channelList) ? channelList : [];

      projectsCache = nextProjects;
      hostsCache = nextHosts;
      channelsCache = nextChannels;
      setProjects(nextProjects);
      setHosts(nextHosts);
      setChannels(nextChannels);
    } catch (error) {
      console.error(error);
    } finally {
      loadingRef.current = false;
    }
  };

  useEffect(() => {
    loadData();

    const handleOpenDialog = (e: CustomEvent) => {
      if (e.detail.mode === "create") {
        setEditingProject(null);
        setFormData(createDefaultFormData());
        setActiveTab("basic");
        setDialogOpen(true);
      }
    };
    window.addEventListener("open-project-dialog", handleOpenDialog as EventListener);
    return () => window.removeEventListener("open-project-dialog", handleOpenDialog as EventListener);
  }, []);

  useEffect(() => {
    if (!dialogOpen || activeTab !== "build" || deferredBuildImage.length < 2) {
      setImageSearchResults([]);
      setImageSearchLoading(false);
      setImageSearchError(null);
      return;
    }

    const requestId = imageSearchRequestRef.current + 1;
    imageSearchRequestRef.current = requestId;
    setImageSearchLoading(true);
    setImageSearchError(null);

    const timer = window.setTimeout(() => {
      void projectApi
        .searchImages(deferredBuildImage, 8)
        .then((response) => {
          if (imageSearchRequestRef.current !== requestId) {
            return;
          }
          setImageSearchResults(response.items);
        })
        .catch((error) => {
          if (imageSearchRequestRef.current !== requestId) {
            return;
          }
          setImageSearchResults([]);
          setImageSearchError(error instanceof Error ? error.message : "官方镜像搜索失败");
        })
        .finally(() => {
          if (imageSearchRequestRef.current === requestId) {
            setImageSearchLoading(false);
          }
        });
    }, 300);

    return () => {
      window.clearTimeout(timer);
    };
  }, [activeTab, deferredBuildImage, dialogOpen]);

  const applyProjectDetailToForm = (detail: ProjectDetail) => {
    setEditingProject(detail.project);
    setFormData({
      name: detail.project.name,
      branch: detail.project.branch,
      repo_url: detail.project.repo_url,
      description: detail.project.description || "",
      git_auth_type: detail.project.git_auth_type || "none",
      git_username: detail.project.git_username || "",
      git_password: "",
      git_ssh_key: "",
      has_existing_git_auth: detail.project.has_git_auth,
      has_existing_git_password: detail.project.has_git_password,
      has_existing_git_ssh_key: detail.project.has_git_ssh_key,
      original_git_auth_type: detail.project.git_auth_type || "none",
      deploy_config: mapDeployConfigToForm(detail.deploy_config),
    });
    setActiveTab("basic");
    setDialogOpen(true);
  };

  const openEditDialog = async (project: Project) => {
    try {
      const detail: ProjectDetail = await projectApi.get(project.id);
      const deployConfig =
        detail.deploy_config ??
        (await projectApi.getDeployConfig(project.id).catch(() => null));
      applyProjectDetailToForm({
        ...detail,
        deploy_config: deployConfig,
      });
    } catch (error: any) {
      toast.error(error.message || "加载项目详情失败");
    }
  };

  const handleClone = async (project: Project) => {
    const cloneName = `${project.name}-副本`;
    const cloneBranch = `${project.branch}-copy`;

    try {
      const detail = await projectApi.clone(project.id, {
        name: cloneName,
        branch: cloneBranch,
      });
      projectsCache = null;
      await loadData(true);
      applyProjectDetailToForm(detail);
      toast.success("项目已复制，请按需调整分支或仓库");
    } catch (error: any) {
      toast.error(error.message || "复制项目失败");
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    const projectError = getProjectValidationError(formData, !!editingProject);
    if (projectError) {
      setActiveTab("basic");
      toast.error(projectError);
      return;
    }

    const deployConfigError = getDeployConfigValidationError(formData.deploy_config);
    if (deployConfigError) {
      setActiveTab("deploy");
      toast.error(deployConfigError);
      return;
    }

    try {
      const projectPayload = buildProjectPayload(formData, !!editingProject);

      if (editingProject) {
        await projectApi.update(editingProject.id, projectPayload);
      } else {
        await projectApi.create(projectPayload);
      }

      projectsCache = null;
      setDialogOpen(false);
      await loadData(true);
      toast.success(editingProject ? "项目更新成功" : "项目创建成功");
    } catch (error: any) {
      toast.error(error.message || "操作失败");
    }
  };

  const handleDelete = async (project: Project) => {
    try {
      await projectApi.delete(project.id);
      projectsCache = null;
      toast.success("项目删除成功");
      setDeletingProject(null);
      await loadData(true);
    } catch (error: any) {
      toast.error(error.message || "删除失败");
    }
  };

  const handleTrigger = async (id: number) => {
    try {
      const run = await projectApi.trigger(id);
      setPendingRunId(run.id);
      setActiveView("logs");
      toast.success("构建已触发");
    } catch (error: any) {
      toast.error(error.message || "触发失败");
    }
  };

  const copyWebhook = (token: string) => {
    const origin =
      typeof window !== "undefined" ? window.location.origin : "";
    const url = `${origin}/api/v1/webhooks/${token}`;
    navigator.clipboard.writeText(url);
    toast.success("Webhook URL 已复制");
  };

  const handleDragStart = (event: DragStartEvent) => {
    setActiveProjectId(Number(event.active.id));
  };

  const handleDragCancel = () => {
    setActiveProjectId(null);
  };

  const activeProject =
    activeProjectId == null ? null : projects.find((item) => item.id === activeProjectId) ?? null;

  const normalizedSearchTerm = deferredSearchTerm.trim().toLowerCase();
  const visibleProjects = projects
    .filter((project) => {
      if (projectFilter === "all") {
        return true;
      }
      return project.git_auth_type === projectFilter;
    })
    .filter((project) => {
      if (!normalizedSearchTerm) {
        return true;
      }

      return [project.name, project.branch, project.repo_url, project.description]
        .join(" ")
        .toLowerCase()
        .includes(normalizedSearchTerm);
    });

  if (sortOrder === "name-asc") {
    visibleProjects.sort((left, right) => left.name.localeCompare(right.name, "zh-CN"));
  } else if (sortOrder === "name-desc") {
    visibleProjects.sort((left, right) => right.name.localeCompare(left.name, "zh-CN"));
  }

  const reorderEnabled =
    sortOrder === "manual" &&
    projectFilter === "all" &&
    normalizedSearchTerm.length === 0;

  const handleDragEnd = async (event: DragEndEvent) => {
    const { active, over } = event;
    setActiveProjectId(null);

    if (!over || active.id === over.id) {
      return;
    }

    const oldIndex = projects.findIndex((item) => item.id === Number(active.id));
    const newIndex = projects.findIndex((item) => item.id === Number(over.id));
    if (oldIndex === -1 || newIndex === -1 || oldIndex === newIndex) {
      return;
    }

    const nextProjects = arrayMove(projects, oldIndex, newIndex);
    const previousProjects = projects;
    projectsCache = nextProjects;
    setProjects(nextProjects);

    try {
      await projectApi.reorder(nextProjects.map((item) => item.id));
    } catch (error: any) {
      projectsCache = previousProjects;
      setProjects(previousProjects);
      toast.error(error.message || "项目排序保存失败");
    }
  };

  return (
    <div className="space-y-4">
      {projects.length === 0 ? (
        <Card>
          <CardContent className="p-12 text-center text-muted-foreground">
            暂无项目，点击上方按钮新增项目。
          </CardContent>
        </Card>
      ) : visibleProjects.length === 0 ? (
        <Card>
          <CardContent className="p-12 text-center text-muted-foreground">
            没有找到符合当前查询条件的项目。
          </CardContent>
        </Card>
      ) : reorderEnabled ? (
        <DndContext
          sensors={sensors}
          collisionDetection={closestCenter}
          modifiers={[restrictToParentElement]}
          onDragStart={handleDragStart}
          onDragCancel={handleDragCancel}
          onDragEnd={(event) => void handleDragEnd(event)}
        >
          <SortableContext
            items={visibleProjects.map((project) => project.id)}
            strategy={rectSortingStrategy}
          >
            <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
              {visibleProjects.map((project) => (
                <SortableProjectCard
                  key={project.id}
                  project={project}
                  isDeleting={deletingProject?.id === project.id}
                  onTrigger={() => void handleTrigger(project.id)}
                  onCopyWebhook={() => copyWebhook(project.webhook_token)}
                  onClone={() => void handleClone(project)}
                  onEdit={() => void openEditDialog(project)}
                  onDeleteRequest={() => setDeletingProject(project)}
                  onDeleteCancel={() => setDeletingProject(null)}
                  onDeleteConfirm={() => void handleDelete(project)}
                />
              ))}
            </div>
          </SortableContext>
          <DragOverlay>
            {activeProject ? (
              <div className="w-full max-w-[min(100%,32rem)]">
                <ProjectCardView
                  project={activeProject}
                  isDeleting={false}
                  isDragging
                  isOverlay
                  onTrigger={() => {}}
                  onCopyWebhook={() => {}}
                  onClone={() => {}}
                  onEdit={() => {}}
                  onDeleteRequest={() => {}}
                  onDeleteCancel={() => {}}
                  onDeleteConfirm={() => {}}
                />
              </div>
            ) : null}
          </DragOverlay>
        </DndContext>
      ) : (
        <div className="grid grid-cols-1 gap-4 lg:grid-cols-2">
          {visibleProjects.map((project) => (
            <ProjectCardView
              key={project.id}
              project={project}
              isDeleting={deletingProject?.id === project.id}
              onTrigger={() => void handleTrigger(project.id)}
              onCopyWebhook={() => copyWebhook(project.webhook_token)}
              onClone={() => void handleClone(project)}
              onEdit={() => void openEditDialog(project)}
              onDeleteRequest={() => setDeletingProject(project)}
              onDeleteCancel={() => setDeletingProject(null)}
              onDeleteConfirm={() => void handleDelete(project)}
            />
          ))}
        </div>
      )}

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="max-h-[90vh] overflow-y-auto sm:max-w-[600px]">
          <DialogHeader>
            <DialogTitle>{editingProject ? "编辑项目" : "新增项目"}</DialogTitle>
            <DialogDescription className="sr-only">
              配置项目基础信息、编译参数和部署参数。
            </DialogDescription>
          </DialogHeader>
          <form onSubmit={handleSubmit} autoComplete="off">
            <input
              type="text"
              name="username"
              autoComplete="username"
              tabIndex={-1}
              aria-hidden="true"
              className="hidden"
            />
            <input
              type="password"
              name="password"
              autoComplete="current-password"
              tabIndex={-1}
              aria-hidden="true"
              className="hidden"
            />
            <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
              <TabsList className="grid w-full grid-cols-3">
                <TabsTrigger value="basic">基础信息</TabsTrigger>
                <TabsTrigger value="build">编译配置</TabsTrigger>
                <TabsTrigger value="deploy">部署配置</TabsTrigger>
              </TabsList>

              <TabsContent value="basic" className="mt-4 space-y-4">
                <div className="grid gap-2">
                  <Label>项目名称</Label>
                  <Input
                    name="project_name"
                    autoComplete="off"
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    placeholder="portal-prod"
                    required
                  />
                </div>
                <div className="grid gap-2">
                  <Label>分支</Label>
                  <Input
                    name="project_branch"
                    autoComplete="off"
                    value={formData.branch}
                    onChange={(e) => setFormData({ ...formData, branch: e.target.value })}
                    placeholder="main"
                    required
                  />
                </div>
                <div className="grid gap-2">
                  <Label>仓库地址</Label>
                  <Input
                    name="project_repo_url"
                    autoComplete="off"
                    value={formData.repo_url}
                    onChange={(e) => setFormData({ ...formData, repo_url: e.target.value })}
                    placeholder="https://github.com/example/app.git"
                    required
                  />
                </div>
                <div className="grid gap-2">
                  <Label>仓库认证方式</Label>
                  <Select
                    value={formData.git_auth_type}
                    onValueChange={(value) =>
                      setFormData({
                        ...formData,
                        git_auth_type: value as GitAuthType,
                      })
                    }
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="none">公开仓库（无认证）</SelectItem>
                      <SelectItem value="username">用户名 / 密码</SelectItem>
                      <SelectItem value="token">用户名 / Token</SelectItem>
                      <SelectItem value="ssh">SSH 私钥</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                {formData.git_auth_type === "username" || formData.git_auth_type === "token" ? (
                  <>
                    <div className="grid gap-2">
                      <Label>Git 用户名</Label>
                      <Input
                        name="git_credential_user"
                        autoComplete="off"
                        data-1p-ignore="true"
                        data-lpignore="true"
                        spellCheck={false}
                        value={formData.git_username}
                        onChange={(e) =>
                          setFormData({ ...formData, git_username: e.target.value })
                        }
                        placeholder={formData.git_auth_type === "token" ? "gitee 用户名" : "git 用户名"}
                      />
                    </div>
                    <div className="grid gap-2">
                      <Label>{formData.git_auth_type === "token" ? "Git Token" : "Git 密码"}</Label>
                      <Input
                        type="password"
                        name="git_credential_secret"
                        autoComplete="new-password"
                        data-1p-ignore="true"
                        data-lpignore="true"
                        value={formData.git_password}
                        onChange={(e) =>
                          setFormData({ ...formData, git_password: e.target.value })
                        }
                        placeholder={
                          editingProject && formData.has_existing_git_password
                            ? formData.git_auth_type === "token"
                              ? "已保存 Token，留空则不修改"
                              : "已保存密码，留空则不修改"
                            : formData.git_auth_type === "token"
                              ? "输入访问 Token"
                              : "输入仓库密码"
                        }
                      />
                      {editingProject &&
                      formData.has_existing_git_password &&
                      (formData.original_git_auth_type === "username" ||
                        formData.original_git_auth_type === "token") ? (
                        <p className="text-xs text-muted-foreground">
                          用户名会直接回显；密码 / Token 已保存，留空则继续沿用。
                        </p>
                      ) : null}
                    </div>
                  </>
                ) : null}
                {formData.git_auth_type === "ssh" ? (
                  <div className="grid gap-2">
                    <Label>SSH 私钥</Label>
                    <Textarea
                      name="git_credential_ssh_key"
                      autoComplete="off"
                      data-1p-ignore="true"
                      data-lpignore="true"
                      spellCheck={false}
                      value={formData.git_ssh_key}
                      onChange={(e) =>
                        setFormData({ ...formData, git_ssh_key: e.target.value })
                      }
                      placeholder={"请输入私钥内容，例如\n-----BEGIN OPENSSH PRIVATE KEY-----"}
                      rows={6}
                    />
                    {editingProject &&
                    formData.has_existing_git_ssh_key &&
                    formData.original_git_auth_type === "ssh" ? (
                      <p className="text-xs text-muted-foreground">
                        留空则保留当前已保存的 SSH 私钥。
                      </p>
                    ) : null}
                  </div>
                ) : null}
                <div className="flex justify-end gap-2">
                  <Button type="button" variant="outline" onClick={() => setDialogOpen(false)}>
                    取消
                  </Button>
                  <Button type="button" onClick={() => setActiveTab("build")}>
                    下一步
                  </Button>
                </div>
              </TabsContent>

              <TabsContent value="build" className="mt-4 space-y-4">
                <div className="grid gap-2">
                  <Label>编译镜像</Label>
                  <div className="relative">
                    <Input
                      value={formData.deploy_config.build_image}
                      onChange={(e) =>
                        setFormData({
                          ...formData,
                          deploy_config: { ...formData.deploy_config, build_image: e.target.value },
                        })
                      }
                      onFocus={() => setShowImageSearchResults(true)}
                      onBlur={() => {
                        window.setTimeout(() => setShowImageSearchResults(false), 150);
                      }}
                      placeholder="node:20"
                      required
                    />
                    <div className="pointer-events-none absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground">
                      {imageSearchLoading ? (
                        <LoaderCircle className="h-4 w-4 animate-spin" />
                      ) : (
                        <Search className="h-4 w-4" />
                      )}
                    </div>
                    {showImageSearchResults &&
                    normalizeBuildImageInput(formData.deploy_config.build_image).length >= 2 ? (
                      <div className="absolute z-20 mt-2 w-full overflow-hidden rounded-xl border border-border bg-background shadow-xl">
                        {imageSearchError ? (
                          <div className="px-3 py-2 text-sm text-destructive">{imageSearchError}</div>
                        ) : imageSearchResults.length > 0 ? (
                          <div className="max-h-64 overflow-y-auto py-1">
                            {imageSearchResults.map((item) => (
                              <button
                                key={`${item.name}-${item.star_count}`}
                                type="button"
                                className="flex w-full items-start justify-between gap-3 px-3 py-2 text-left transition-colors hover:bg-muted"
                                onMouseDown={(event) => {
                                  event.preventDefault();
                                  setFormData({
                                    ...formData,
                                    deploy_config: {
                                      ...formData.deploy_config,
                                      build_image: item.name,
                                    },
                                  });
                                  setShowImageSearchResults(false);
                                }}
                              >
                                <div className="min-w-0">
                                  <div className="font-medium text-foreground">{item.display_name}</div>
                                  {item.description ? (
                                    <div className="truncate text-xs text-muted-foreground">
                                      {item.description}
                                    </div>
                                  ) : null}
                                </div>
                                <div className="shrink-0 text-xs text-muted-foreground">
                                  ★ {item.star_count}
                                </div>
                              </button>
                            ))}
                          </div>
                        ) : (
                          <div className="px-3 py-2 text-sm text-muted-foreground">
                            没找到官方镜像，仍可直接输入完整镜像名。
                          </div>
                        )}
                      </div>
                    ) : null}
                  </div>
                  <p className="text-xs text-muted-foreground">
                    仅搜索 Docker Hub 官方镜像，选择后仍可自行补充 tag，例如 `node:20`。
                  </p>
                </div>
                <div className="grid gap-2">
                  <Label>编译命令</Label>
                  <Textarea
                    value={formData.deploy_config.build_commands}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        deploy_config: { ...formData.deploy_config, build_commands: e.target.value },
                      })
                    }
                    placeholder={"每行一个命令，例如\npnpm install\npnpm run build"}
                    rows={5}
                    required
                  />
                </div>
                <div className="grid gap-2">
                  <Label>制品过滤模式</Label>
                  <Select
                    value={formData.deploy_config.artifact_filter_mode}
                    onValueChange={(value) =>
                      setFormData({
                        ...formData,
                        deploy_config: {
                          ...formData.deploy_config,
                          artifact_filter_mode: value as DeployConfigFormState["artifact_filter_mode"],
                        },
                      })
                    }
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="include">包含</SelectItem>
                      <SelectItem value="exclude">排除</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div className="grid gap-2">
                  <Label>制品过滤规则</Label>
                  <Textarea
                    value={formData.deploy_config.artifact_rules}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        deploy_config: { ...formData.deploy_config, artifact_rules: e.target.value },
                      })
                    }
                    placeholder="每行一个目录或文件名"
                    rows={4}
                  />
                </div>
                <div className="flex justify-end gap-2">
                  <Button type="button" variant="outline" onClick={() => setActiveTab("basic")}>
                    上一步
                  </Button>
                  <Button type="button" onClick={() => setActiveTab("deploy")}>
                    下一步
                  </Button>
                </div>
              </TabsContent>

              <TabsContent value="deploy" className="mt-4 space-y-4">
                <div className="grid gap-2">
                  <Label>目标主机</Label>
                  <Select
                    value={formData.deploy_config.host_id}
                    onValueChange={(value) =>
                      setFormData({
                        ...formData,
                        deploy_config: { ...formData.deploy_config, host_id: value },
                      })
                    }
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="请选择主机" />
                    </SelectTrigger>
                    <SelectContent>
                      {hosts.map((host) => (
                        <SelectItem key={host.id} value={String(host.id)}>
                          {host.name} ({host.address})
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="grid grid-cols-2 gap-4">
                  <div className="grid gap-2">
                    <Label>远程保存目录</Label>
                    <Input
                      value={formData.deploy_config.remote_save_dir}
                      onChange={(e) =>
                        setFormData({
                          ...formData,
                          deploy_config: { ...formData.deploy_config, remote_save_dir: e.target.value },
                        })
                      }
                      placeholder="/data/jimuqu/projects"
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label>远程部署目录</Label>
                    <Input
                      value={formData.deploy_config.remote_deploy_dir}
                      onChange={(e) =>
                        setFormData({
                          ...formData,
                          deploy_config: { ...formData.deploy_config, remote_deploy_dir: e.target.value },
                        })
                      }
                      placeholder="/data/apps/portal"
                    />
                  </div>
                </div>
                <div className="grid grid-cols-1 gap-4 md:grid-cols-2">
                  <div className="grid gap-2">
                    <Label>保留版本数量</Label>
                    <Input
                      type="number"
                      min={1}
                      value={formData.deploy_config.version_count}
                      onChange={(e) =>
                        setFormData({
                          ...formData,
                          deploy_config: {
                            ...formData.deploy_config,
                            version_count: Number(e.target.value) || 0,
                          },
                        })
                      }
                    />
                    <p className="text-xs text-muted-foreground">
                      远程保存目录默认保留最近 5 个历史版本，超出后会自动清理旧版本。
                    </p>
                  </div>
                  <div className="grid gap-2">
                    <Label>部署超时(分钟)</Label>
                    <Input
                      type="number"
                      min={1}
                      value={formData.deploy_config.timeout_minutes}
                      onChange={(e) =>
                        setFormData({
                          ...formData,
                          deploy_config: {
                            ...formData.deploy_config,
                            timeout_minutes: Number(e.target.value) || 0,
                          },
                        })
                      }
                    />
                  </div>
                </div>
                <div className="grid gap-2">
                  <Label>部署前命令</Label>
                  <Textarea
                    value={formData.deploy_config.pre_deploy_commands}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        deploy_config: { ...formData.deploy_config, pre_deploy_commands: e.target.value },
                      })
                    }
                    placeholder="每行一个远程命令"
                    rows={3}
                  />
                </div>
                <div className="grid gap-2">
                  <Label>部署后命令</Label>
                  <Textarea
                    value={formData.deploy_config.post_deploy_commands}
                    onChange={(e) =>
                      setFormData({
                        ...formData,
                        deploy_config: { ...formData.deploy_config, post_deploy_commands: e.target.value },
                      })
                    }
                    placeholder={"例如\ndocker restart app"}
                    rows={3}
                  />
                </div>
                <div className="grid gap-2">
                  <Label>通知渠道</Label>
                  <Select
                    value={formData.deploy_config.notification_channel_id}
                    onValueChange={(value) =>
                      setFormData({
                        ...formData,
                        deploy_config: {
                          ...formData.deploy_config,
                          notification_channel_id: value,
                        },
                      })
                    }
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="使用默认渠道" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value={DEFAULT_NOTIFICATION_CHANNEL}>使用默认渠道</SelectItem>
                      {channels.map((channel) => (
                        <SelectItem key={channel.id} value={String(channel.id)}>
                          {channel.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <DialogFooter className="mt-4">
                  <Button type="button" variant="outline" onClick={() => setActiveTab("build")}>
                    上一步
                  </Button>
                  <Button type="submit">保存</Button>
                </DialogFooter>
              </TabsContent>
            </Tabs>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
