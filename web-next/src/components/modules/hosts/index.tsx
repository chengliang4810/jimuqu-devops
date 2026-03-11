"use client";

import { useEffect, useState } from "react";
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
import { Plus, Pencil, Trash2 } from "lucide-react";

export function Hosts() {
  const [hosts, setHosts] = useState<Host[]>([]);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingHost, setEditingHost] = useState<Host | null>(null);
  const [formData, setFormData] = useState({
    name: "",
    address: "",
    port: 22,
    username: "",
    password: "",
  });

  const loadHosts = async () => {
    try {
      const data = await hostApi.list();
      setHosts(data?.hosts || []);
    } catch (error) {
      console.error(error);
    }
  };

  useEffect(() => {
    loadHosts();
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

  const handleDelete = async (id: number) => {
    if (!confirm("确定要删除此主机吗？")) return;
    try {
      await hostApi.delete(id);
      toast.success("主机删除成功");
      loadHosts();
    } catch (error: any) {
      toast.error(error.message || "删除失败");
    }
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-end">
        <div className="flex items-center gap-4">
          <Badge variant="secondary">{hosts.length} 台主机</Badge>
          <Button onClick={openCreateDialog}>
            <Plus className="h-4 w-4 mr-2" />
            新增主机
          </Button>
        </div>
      </div>

      {hosts.length === 0 ? (
        <Card>
          <CardContent className="p-12 text-center text-muted-foreground">
            暂无主机，点击上方按钮新增部署目标。
          </CardContent>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {hosts.map((host) => (
            <Card key={host.id} className="relative group">
              <CardContent className="p-6">
                <button
                  onClick={() => handleDelete(host.id)}
                  className="absolute top-4 right-4 p-2 rounded-md text-destructive hover:bg-destructive/10 opacity-0 group-hover:opacity-100 transition-opacity"
                >
                  <Trash2 className="h-4 w-4" />
                </button>
                <h3 className="font-semibold text-foreground mb-2">{host.name}</h3>
                <p className="text-sm text-muted-foreground mb-1">{host.address}:{host.port}</p>
                <p className="text-sm text-muted-foreground mb-4">{host.username}</p>
                <Button variant="outline" size="sm" onClick={() => openEditDialog(host)}>
                  <Pencil className="h-4 w-4 mr-2" />
                  编辑
                </Button>
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
