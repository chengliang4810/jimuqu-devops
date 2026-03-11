"use client";

import { useEffect, useState } from "react";
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
import { Plus, Pencil, Trash2, Play, Copy } from "lucide-react";

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
  const [projects, setProjects] = useState<Project[]>([]);
  const [hosts, setHosts] = useState<Host[]>([]);
  const [channels, setChannels] = useState<NotifyChannel[]>([]);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingProject, setEditingProject] = useState<Project | null>(null);
  const [activeTab, setActiveTab] = useState("basic");
  const [formData, setFormData] = useState({
    name: "",
    branch: "main",
    repo_url: "",
    description: "",
    timeout_minutes: 30,
    deploy_config: { ...defaultDeployConfig } as Partial<DeployConfig>,
  });

  const loadData = async () => {
    try {
      const [projRes, hostRes, chanRes] = await Promise.all([
        projectApi.list(),
        hostApi.list(),
        notifyApi.list(),
      ]);
      setProjects(projRes?.projects || []);
      setHosts(hostRes?.hosts || []);
      setChannels(chanRes?.channels || []);
    } catch (error) {
      console.error(error);
    }
  };

  useEffect(() => {
    loadData();
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

  const handleDelete = async (id: number) => {
    if (!confirm("确定要删除此项目吗？")) return;
    try {
      await projectApi.delete(id);
      toast.success("项目删除成功");
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
      <div className="flex items-center justify-end">
        <div className="flex items-center gap-4">
          <Badge variant="secondary">{projects.length} 个项目</Badge>
          <Button onClick={openCreateDialog}>
            <Plus className="h-4 w-4 mr-2" />
            新增项目
          </Button>
        </div>
      </div>

      {projects.length === 0 ? (
        <Card>
          <CardContent className="p-12 text-center text-muted-foreground">
            暂无项目，点击上方按钮新增项目。
          </CardContent>
        </Card>
      ) : (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-4">
          {projects.map((project) => (
            <Card key={project.id} className="relative group">
              <CardContent className="p-6">
                <button
                  onClick={() => handleDelete(project.id)}
                  className="absolute top-4 right-4 p-2 rounded-md text-destructive hover:bg-destructive/10 opacity-0 group-hover:opacity-100 transition-opacity"
                >
                  <Trash2 className="h-4 w-4" />
                </button>
                <div className="mb-4">
                  <h3 className="font-semibold text-foreground">{project.name}</h3>
                  <p className="text-sm text-muted-foreground">{project.branch}</p>
                  <p className="text-sm text-muted-foreground truncate">{project.repo_url}</p>
                </div>
                <div className="flex flex-wrap gap-2">
                  <Button variant="outline" size="sm" onClick={() => openEditDialog(project)}>
                    <Pencil className="h-4 w-4 mr-2" />
                    编辑
                  </Button>
                  <Button variant="outline" size="sm" onClick={() => handleTrigger(project.id)}>
                    <Play className="h-4 w-4 mr-2" />
                    构建
                  </Button>
                  <Button variant="outline" size="sm" onClick={() => copyWebhook(project.webhook_token)}>
                    <Copy className="h-4 w-4 mr-2" />
                    Webhook
                  </Button>
                </div>
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