"use client";

import { useEffect, useState, useRef } from "react";
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
  DialogFooter,
} from "@/components/ui/dialog";
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { Tabs, TabsList, TabsTrigger, TabsContent } from "@/components/ui/tabs";
import { projectApi, hostApi, notifyApi } from "@/api/client";
import type { Project, Host, NotifyChannel, DeployConfig } from "@/types";
import { toast } from "sonner";
import { Plus, Pencil, Trash2, Play, Copy, X } from "lucide-react";
import { motion, AnimatePresence } from "motion/react";

// 全局缓存
let projectsCache: Project[] | null = null;
let hostsCache: Host[] | null = null;
let channelsCache: NotifyChannel[] | null = null;

const defaultDeployConfig: Partial<DeployConfig> = {
  build_image: "node:20",
  build_commands: "",
  artifact_filter_mode: "exclude",
  artifact_rules: "",
  remote_save_dir: "/data/jimuqu/projects",
  remote_deploy_dir: "",
  pre_deploy_commands: "",
  post_deploy_commands: "",
  version_count: 5,
  notification_channel_id: null,
};

export function Projects() {
  const [projects, setProjects] = useState<Project[]>(projectsCache || []);
  const [hosts, setHosts] = useState<Host[]>(hostsCache || []);
  const [channels, setChannels] = useState<NotifyChannel[]>(channelsCache || []);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingProject, setEditingProject] = useState<Project | null>(null);
  const [deletingProject, setDeletingProject] = useState<Project | null>(null);
  const [activeTab, setActiveTab] = useState("basic");
  const loadingRef = useRef(false);
  const [formData, setFormData] = useState({
    name: "",
    branch: "main",
    repo_url: "",
    description: "",
    timeout_minutes: 30,
    deploy_config: { ...defaultDeployConfig } as Partial<DeployConfig>,
  });

  const loadData = async () => {
    // 如果有缓存且不是强制刷新，使用缓存
    if (projectsCache && hostsCache && channelsCache) {
      setProjects(projectsCache);
      setHosts(hostsCache);
      setChannels(channelsCache);
      return;
    }
    if (loadingRef.current) return;
    loadingRef.current = true;
    try {
      const [projRes, hostRes, chanRes] = await Promise.all([
        projectsCache ? Promise.resolve(projectsCache) : projectApi.list(),
        hostsCache ? Promise.resolve(hostsCache) : hostApi.list(),
        channelsCache ? Promise.resolve(channelsCache) : notifyApi.list(),
      ]);
      const projects = Array.isArray(projRes) ? projRes : [];
      const hosts = Array.isArray(hostRes) ? hostRes : [];
      const channels = Array.isArray(chanRes) ? chanRes : [];
      projectsCache = projects;
      hostsCache = hosts;
      channelsCache = channels;
      setProjects(projects);
      setHosts(hosts);
      setChannels(channels);
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
        openCreateDialog();
      }
    };
    window.addEventListener("open-project-dialog", handleOpenDialog as EventListener);
    return () => window.removeEventListener("open-project-dialog", handleOpenDialog as EventListener);
  }, []);

  const resetForm = () => {
    setFormData({
      name: "",
      branch: "main",
      repo_url: "",
      description: "",
      timeout_minutes: 30,
      deploy_config: { ...defaultDeployConfig },
    });
    setActiveTab("basic");
  };

  const openCreateDialog = () => {
    setEditingProject(null);
    resetForm();
    setDialogOpen(true);
  };

  const openEditDialog = (project: Project) => {
    setEditingProject(project);
    setFormData({
      name: project.name,
      branch: project.branch,
      repo_url: project.repo_url,
      description: project.description,
      timeout_minutes: project.timeout_minutes,
      deploy_config: project.deploy_config || { ...defaultDeployConfig },
    });
    setActiveTab("basic");
    setDialogOpen(true);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      if (editingProject) {
        await projectApi.update(editingProject.id, formData);
        toast.success("项目更新成功");
      } else {
        await projectApi.create(formData);
        toast.success("项目创建成功");
      }
      setDialogOpen(false);
      loadData();
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
      loadData();
    } catch (error: any) {
      toast.error(error.message || "删除失败");
    }
  };

  const handleTrigger = async (id: number) => {
    try {
      await projectApi.trigger(id);
      toast.success("构建已触发");
    } catch (error: any) {
      toast.error(error.message || "触发失败");
    }
  };

  const copyWebhook = (token: string) => {
    const url = `${window.location.origin}/api/v1/webhooks/${token}`;
    navigator.clipboard.writeText(url);
    toast.success("Webhook URL 已复制");
  };

  return (
    <div className="space-y-4">
      {projects.length === 0 ? (
        <Card>
          <CardContent className="p-12 text-center text-muted-foreground">
            暂无项目，点击上方按钮新增项目。
          </CardContent>
        </Card>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
          {projects.map((project) => (
            <Card key={project.id} className="overflow-hidden">
              <CardContent className="p-4">
                <header className="flex items-start justify-between mb-3 relative">
                  <div className="flex-1 mr-2 min-w-0">
                    <h3 className="font-semibold text-foreground truncate">{project.name}</h3>
                    <p className="text-sm text-muted-foreground">{project.branch}</p>
                  </div>
                  <div className="flex items-center gap-1 shrink-0">
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-8 w-8"
                      onClick={() => handleTrigger(project.id)}
                    >
                      <Play className="h-4 w-4" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-8 w-8"
                      onClick={() => copyWebhook(project.webhook_token)}
                    >
                      <Copy className="h-4 w-4" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-8 w-8"
                      onClick={() => openEditDialog(project)}
                    >
                      <Pencil className="h-4 w-4" />
                    </Button>
                    {deletingProject?.id !== project.id ? (
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8 text-muted-foreground hover:text-destructive"
                        onClick={() => setDeletingProject(project)}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    ) : null}
                  </div>

                  <AnimatePresence>
                    {deletingProject?.id === project.id && (
                      <motion.div
                        initial={{ opacity: 0 }}
                        animate={{ opacity: 1 }}
                        exit={{ opacity: 0 }}
                        className="absolute inset-0 flex items-center justify-center gap-2 bg-destructive p-2 rounded-lg"
                      >
                        <button
                          type="button"
                          onClick={() => setDeletingProject(null)}
                          className="flex h-7 w-7 items-center justify-center rounded-lg bg-white/20 text-white transition-all hover:bg-white/30 active:scale-95"
                        >
                          <X className="h-4 w-4" />
                        </button>
                        <button
                          type="button"
                          onClick={() => handleDelete(project)}
                          className="flex-1 h-7 flex items-center justify-center gap-2 rounded-lg bg-white text-destructive text-sm font-semibold transition-all hover:bg-white/90 active:scale-[0.98]"
                        >
                          <Trash2 className="h-3.5 w-3.5" />
                          确认删除
                        </button>
                      </motion.div>
                    )}
                  </AnimatePresence>
                </header>

                <p className="text-sm text-muted-foreground truncate">
                  {project.repo_url}
                </p>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="sm:max-w-[600px] max-h-[90vh] overflow-y-auto">
          <DialogHeader>
            <DialogTitle>{editingProject ? "编辑项目" : "新增项目"}</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmit}>
            <Tabs value={activeTab} onValueChange={setActiveTab} className="w-full">
              <TabsList className="grid w-full grid-cols-3">
                <TabsTrigger value="basic">基础信息</TabsTrigger>
                <TabsTrigger value="build">编译配置</TabsTrigger>
                <TabsTrigger value="deploy">部署配置</TabsTrigger>
              </TabsList>

              <TabsContent value="basic" className="space-y-4 mt-4">
                <div className="grid gap-2">
                  <Label>项目名称</Label>
                  <Input
                    value={formData.name}
                    onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                    placeholder="portal-prod"
                    required
                  />
                </div>
                <div className="grid gap-2">
                  <Label>分支</Label>
                  <Input
                    value={formData.branch}
                    onChange={(e) => setFormData({ ...formData, branch: e.target.value })}
                    placeholder="main"
                    required
                  />
                </div>
                <div className="grid gap-2">
                  <Label>仓库地址</Label>
                  <Input
                    value={formData.repo_url}
                    onChange={(e) => setFormData({ ...formData, repo_url: e.target.value })}
                    placeholder="https://github.com/example/app.git"
                    required
                  />
                </div>
                <div className="grid gap-2">
                  <Label>部署超时(分钟)</Label>
                  <Input
                    type="number"
                    value={formData.timeout_minutes}
                    onChange={(e) => setFormData({ ...formData, timeout_minutes: parseInt(e.target.value) })}
                    required
                  />
                </div>
                <div className="grid gap-2">
                  <Label>描述</Label>
                  <Textarea
                    value={formData.description}
                    onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                    placeholder="可填写环境、用途或备注"
                    rows={3}
                  />
                </div>
                <div className="flex justify-end gap-2">
                  <Button type="button" variant="outline" onClick={() => setDialogOpen(false)}>
                    取消
                  </Button>
                  <Button type="button" onClick={() => setActiveTab("build")}>
                    下一步
                  </Button>
                </div>
              </TabsContent>

              <TabsContent value="build" className="space-y-4 mt-4">
                <div className="grid gap-2">
                  <Label>编译镜像</Label>
                  <Input
                    value={formData.deploy_config.build_image}
                    onChange={(e) => setFormData({
                      ...formData,
                      deploy_config: { ...formData.deploy_config, build_image: e.target.value }
                    })}
                    placeholder="node:20"
                    required
                  />
                </div>
                <div className="grid gap-2">
                  <Label>编译命令</Label>
                  <Textarea
                    value={formData.deploy_config.build_commands}
                    onChange={(e) => setFormData({
                      ...formData,
                      deploy_config: { ...formData.deploy_config, build_commands: e.target.value }
                    })}
                    placeholder={"每行一个命令，例如\npnpm install\npnpm run build"}
                    rows={5}
                  />
                </div>
                <div className="grid gap-2">
                  <Label>制品过滤模式</Label>
                  <Select
                    value={formData.deploy_config.artifact_filter_mode}
                    onValueChange={(v) => setFormData({
                      ...formData,
                      deploy_config: { ...formData.deploy_config, artifact_filter_mode: v as any }
                    })}
                  >
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="none">无</SelectItem>
                      <SelectItem value="include">包含</SelectItem>
                      <SelectItem value="exclude">排除</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
                <div className="grid gap-2">
                  <Label>制品过滤规则</Label>
                  <Textarea
                    value={formData.deploy_config.artifact_rules}
                    onChange={(e) => setFormData({
                      ...formData,
                      deploy_config: { ...formData.deploy_config, artifact_rules: e.target.value }
                    })}
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

              <TabsContent value="deploy" className="space-y-4 mt-4">
                <div className="grid gap-2">
                  <Label>目标主机</Label>
                  <Select
                    value={formData.deploy_config.host_id?.toString() || ""}
                    onValueChange={(v) => setFormData({
                      ...formData,
                      deploy_config: { ...formData.deploy_config, host_id: parseInt(v) }
                    })}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="请选择主机" />
                    </SelectTrigger>
                    <SelectContent>
                      {hosts.map((host) => (
                        <SelectItem key={host.id} value={host.id.toString()}>
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
                      onChange={(e) => setFormData({
                        ...formData,
                        deploy_config: { ...formData.deploy_config, remote_save_dir: e.target.value }
                      })}
                      placeholder="/data/jimuqu/projects"
                    />
                  </div>
                  <div className="grid gap-2">
                    <Label>远程部署目录</Label>
                    <Input
                      value={formData.deploy_config.remote_deploy_dir}
                      onChange={(e) => setFormData({
                        ...formData,
                        deploy_config: { ...formData.deploy_config, remote_deploy_dir: e.target.value }
                      })}
                      placeholder="/data/apps/portal"
                    />
                  </div>
                </div>
                <div className="grid gap-2">
                  <Label>保留版本数量</Label>
                  <Input
                    type="number"
                    value={formData.deploy_config.version_count}
                    onChange={(e) => setFormData({
                      ...formData,
                      deploy_config: { ...formData.deploy_config, version_count: parseInt(e.target.value) }
                    })}
                  />
                </div>
                <div className="grid gap-2">
                  <Label>部署前命令</Label>
                  <Textarea
                    value={formData.deploy_config.pre_deploy_commands}
                    onChange={(e) => setFormData({
                      ...formData,
                      deploy_config: { ...formData.deploy_config, pre_deploy_commands: e.target.value }
                    })}
                    placeholder="每行一个远程命令"
                    rows={3}
                  />
                </div>
                <div className="grid gap-2">
                  <Label>部署后命令</Label>
                  <Textarea
                    value={formData.deploy_config.post_deploy_commands}
                    onChange={(e) => setFormData({
                      ...formData,
                      deploy_config: { ...formData.deploy_config, post_deploy_commands: e.target.value }
                    })}
                    placeholder={"例如\ndocker restart app"}
                    rows={3}
                  />
                </div>
                <div className="grid gap-2">
                  <Label>通知渠道</Label>
                  <Select
                    value={formData.deploy_config.notification_channel_id?.toString() || ""}
                    onValueChange={(v) => setFormData({
                      ...formData,
                      deploy_config: {
                        ...formData.deploy_config,
                        notification_channel_id: v ? parseInt(v) : null
                      }
                    })}
                  >
                    <SelectTrigger>
                      <SelectValue placeholder="使用默认渠道" />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="">使用默认渠道</SelectItem>
                      <SelectItem value="-1">不通知</SelectItem>
                      {channels.map((ch) => (
                        <SelectItem key={ch.id} value={ch.id.toString()}>
                          {ch.name}
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