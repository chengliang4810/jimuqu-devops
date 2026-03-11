"use client";

import { useEffect, useState, useRef } from "react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Label } from "@/components/ui/label";
import { Card, CardContent } from "@/components/ui/card";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from "@/components/ui/dialog";
import { Badge } from "@/components/ui/badge";
import { hostApi } from "@/api/client";
import type { Host } from "@/types";
import { toast } from "sonner";
import { Plus, Pencil, Trash2, X } from "lucide-react";
import { motion, AnimatePresence } from "motion/react";

// 全局请求缓存
let hostsCache: Host[] | null = null;

export function Hosts() {
  const [hosts, setHosts] = useState<Host[]>(hostsCache || []);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingHost, setEditingHost] = useState<Host | null>(null);
  const [deletingHost, setDeletingHost] = useState<Host | null>(null);
  const loadingRef = useRef(false);
  const [formData, setFormData] = useState({
    name: "",
    address: "",
    port: 22,
    username: "",
    password: "",
  });

  const loadHosts = async () => {
    if (loadingRef.current) return;
    if (hostsCache) {
      setHosts(hostsCache);
      return;
    }
    loadingRef.current = true;
    try {
      const data = await hostApi.list();
      const hosts = Array.isArray(data) ? data : [];
      hostsCache = hosts;
      setHosts(hosts);
    } catch (error) {
      console.error(error);
    } finally {
      loadingRef.current = false;
    }
  };

  useEffect(() => {
    loadHosts();

    const handleOpenDialog = (e: CustomEvent) => {
      if (e.detail.mode === "create") {
        openCreateDialog();
      }
    };
    window.addEventListener("open-host-dialog", handleOpenDialog as EventListener);
    return () => window.removeEventListener("open-host-dialog", handleOpenDialog as EventListener);
  }, []);

  const openCreateDialog = () => {
    setEditingHost(null);
    setFormData({ name: "", address: "", port: 22, username: "", password: "" });
    setDialogOpen(true);
  };

  const openEditDialog = (host: Host) => {
    setEditingHost(host);
    setFormData({
      name: host.name,
      address: host.address,
      port: host.port,
      username: host.username,
      password: "",
    });
    setDialogOpen(true);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      if (editingHost) {
        const updateData: any = { ...formData };
        if (!updateData.password) delete updateData.password;
        await hostApi.update(editingHost.id, updateData);
        toast.success("主机更新成功");
      } else {
        await hostApi.create(formData);
        toast.success("主机创建成功");
      }
      setDialogOpen(false);
      loadHosts();
    } catch (error: any) {
      toast.error(error.message || "操作失败");
    }
  };

  const handleDelete = async (host: Host) => {
    try {
      await hostApi.delete(host.id);
      hostsCache = null; // 清除缓存
      toast.success("主机删除成功");
      setDeletingHost(null);
      loadHosts();
    } catch (error: any) {
      toast.error(error.message || "删除失败");
    }
  };

  return (
    <div className="space-y-4">
      {hosts.length === 0 ? (
        <Card>
          <CardContent className="p-12 text-center text-muted-foreground">
            暂无主机，点击上方按钮新增部署目标。
          </CardContent>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {hosts.map((host) => (
            <Card key={host.id} className="overflow-hidden">
              <CardContent className="p-4">
                <header className="flex items-start justify-between mb-3 relative">
                  <h3 className="font-semibold text-foreground truncate flex-1 mr-2">{host.name}</h3>
                  <div className="flex items-center gap-1 shrink-0">
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-8 w-8"
                      onClick={() => openEditDialog(host)}
                    >
                      <Pencil className="h-4 w-4" />
                    </Button>
                    {deletingHost?.id !== host.id ? (
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8 text-muted-foreground hover:text-destructive"
                        onClick={() => setDeletingHost(host)}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    ) : null}
                  </div>

                  <AnimatePresence>
                    {deletingHost?.id === host.id && (
                      <motion.div
                        initial={{ opacity: 0 }}
                        animate={{ opacity: 1 }}
                        exit={{ opacity: 0 }}
                        className="absolute inset-0 flex items-center justify-center gap-2 bg-destructive p-2 rounded-lg"
                      >
                        <button
                          type="button"
                          onClick={() => setDeletingHost(null)}
                          className="flex h-7 w-7 items-center justify-center rounded-lg bg-white/20 text-white transition-all hover:bg-white/30 active:scale-95"
                        >
                          <X className="h-4 w-4" />
                        </button>
                        <button
                          type="button"
                          onClick={() => handleDelete(host)}
                          className="flex-1 h-7 flex items-center justify-center gap-2 rounded-lg bg-white text-destructive text-sm font-semibold transition-all hover:bg-white/90 active:scale-[0.98]"
                        >
                          <Trash2 className="h-3.5 w-3.5" />
                          确认删除
                        </button>
                      </motion.div>
                    )}
                  </AnimatePresence>
                </header>

                <div className="space-y-1 text-sm text-muted-foreground">
                  <p>{host.address}:{host.port}</p>
                  <p>{host.username}</p>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle>{editingHost ? "编辑主机" : "新增主机"}</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmit}>
            <div className="grid gap-4 py-4">
              <div className="grid gap-2">
                <Label htmlFor="name">主机名称</Label>
                <Input
                  id="name"
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  placeholder="prod-1"
                  required
                />
              </div>
              <div className="grid grid-cols-3 gap-4">
                <div className="col-span-2 grid gap-2">
                  <Label htmlFor="address">IP / 域名</Label>
                  <Input
                    id="address"
                    value={formData.address}
                    onChange={(e) => setFormData({ ...formData, address: e.target.value })}
                    placeholder="192.168.1.10"
                    required
                  />
                </div>
                <div className="grid gap-2">
                  <Label htmlFor="port">SSH 端口</Label>
                  <Input
                    id="port"
                    type="number"
                    value={formData.port}
                    onChange={(e) => setFormData({ ...formData, port: parseInt(e.target.value) })}
                    required
                  />
                </div>
              </div>
              <div className="grid gap-2">
                <Label htmlFor="username">用户名</Label>
                <Input
                  id="username"
                  value={formData.username}
                  onChange={(e) => setFormData({ ...formData, username: e.target.value })}
                  placeholder="root"
                  required
                />
              </div>
              <div className="grid gap-2">
                <Label htmlFor="password">密码</Label>
                <Input
                  id="password"
                  type="password"
                  value={formData.password}
                  onChange={(e) => setFormData({ ...formData, password: e.target.value })}
                  placeholder={editingHost ? "留空表示不变" : ""}
                  required={!editingHost}
                />
              </div>
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setDialogOpen(false)}>
                取消
              </Button>
              <Button type="submit">保存</Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  );
}
