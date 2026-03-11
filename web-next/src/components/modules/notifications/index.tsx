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
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select";
import { Badge } from "@/components/ui/badge";
import { notifyApi } from "@/api/client";
import type { NotifyChannel, NotifyChannelType } from "@/types";
import { toast } from "sonner";
import { Plus, Pencil, Trash2 } from "lucide-react";

const channelTypes: { value: NotifyChannelType; label: string }[] = [
  { value: "webhook", label: "Webhook" },
  { value: "dingtalk", label: "钉钉" },
  { value: "wechat", label: "企业微信" },
  { value: "feishu", label: "飞书" },
];

export function Notifications() {
  const [channels, setChannels] = useState<NotifyChannel[]>([]);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingChannel, setEditingChannel] = useState<NotifyChannel | null>(null);
  const [formData, setFormData] = useState({
    name: "",
    type: "webhook" as NotifyChannelType,
    webhook_url: "",
    secret: "",
    remark: "",
  });

  const loadChannels = async () => {
    try {
      const data = await notifyApi.list();
      setChannels(data?.channels || []);
    } catch (error) {
      console.error(error);
    }
  };

  useEffect(() => {
    loadChannels();
  }, []);

  const openCreateDialog = () => {
    setEditingChannel(null);
    setFormData({ name: "", type: "webhook", webhook_url: "", secret: "", remark: "" });
    setDialogOpen(true);
  };

  const openEditDialog = (channel: NotifyChannel) => {
    setEditingChannel(channel);
    setFormData({
      name: channel.name,
      type: channel.type,
      webhook_url: channel.webhook_url,
      secret: channel.secret || "",
      remark: channel.remark || "",
    });
    setDialogOpen(true);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      if (editingChannel) {
        await notifyApi.update(editingChannel.id, formData);
        toast.success("通知渠道更新成功");
      } else {
        await notifyApi.create(formData);
        toast.success("通知渠道创建成功");
      }
      setDialogOpen(false);
      loadChannels();
    } catch (error: any) {
      toast.error(error.message || "操作失败");
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm("确定要删除此通知渠道吗？")) return;
    try {
      await notifyApi.delete(id);
      toast.success("通知渠道删除成功");
      loadChannels();
    } catch (error: any) {
      toast.error(error.message || "删除失败");
    }
  };

  const getTypeLabel = (type: NotifyChannelType) => {
    return channelTypes.find((t) => t.value === type)?.label || type;
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-end">
        <div className="flex items-center gap-4">
          <Badge variant="secondary">{channels.length} 个渠道</Badge>
          <Button onClick={openCreateDialog}>
            <Plus className="h-4 w-4 mr-2" />
            新增渠道
          </Button>
        </div>
      </div>

      {channels.length === 0 ? (
        <Card>
          <CardContent className="p-12 text-center text-muted-foreground">
            暂无通知渠道，点击上方按钮新增渠道。
          </CardContent>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {channels.map((channel) => (
            <Card key={channel.id} className="relative group">
              <CardContent className="p-6">
                <button
                  onClick={() => handleDelete(channel.id)}
                  className="absolute top-4 right-4 p-2 rounded-md text-destructive hover:bg-destructive/10 opacity-0 group-hover:opacity-100 transition-opacity"
                >
                  <Trash2 className="h-4 w-4" />
                </button>
                <div className="mb-2">
                  <Badge variant="outline">{getTypeLabel(channel.type)}</Badge>
                </div>
                <h3 className="font-semibold text-foreground mb-2">{channel.name}</h3>
                <p className="text-sm text-muted-foreground truncate mb-4">
                  {channel.webhook_url}
                </p>
                <Button variant="outline" size="sm" onClick={() => openEditDialog(channel)}>
                  <Pencil className="h-4 w-4 mr-2" />
                  编辑
                </Button>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle>{editingChannel ? "编辑通知渠道" : "新增通知渠道"}</DialogTitle>
          </DialogHeader>
          <form onSubmit={handleSubmit}>
            <div className="grid gap-4 py-4">
              <div className="grid gap-2">
                <Label>渠道名称</Label>
                <Input
                  value={formData.name}
                  onChange={(e) => setFormData({ ...formData, name: e.target.value })}
                  placeholder="企业微信通知"
                  required
                />
              </div>
              <div className="grid gap-2">
                <Label>渠道类型</Label>
                <Select
                  value={formData.type}
                  onValueChange={(v) => setFormData({ ...formData, type: v as NotifyChannelType })}
                >
                  <SelectTrigger>
                    <SelectValue />
                  </SelectTrigger>
                  <SelectContent>
                    {channelTypes.map((type) => (
                      <SelectItem key={type.value} value={type.value}>
                        {type.label}
                      </SelectItem>
                    ))}
                  </SelectContent>
                </Select>
              </div>
              <div className="grid gap-2">
                <Label>Webhook URL</Label>
                <Input
                  value={formData.webhook_url}
                  onChange={(e) => setFormData({ ...formData, webhook_url: e.target.value })}
                  placeholder="https://oapi.dingtalk.com/robot/send?access_token=xxx"
                  type="url"
                  required
                />
              </div>
              <div className="grid gap-2">
                <Label>签名密钥（可选）</Label>
                <Input
                  value={formData.secret}
                  onChange={(e) => setFormData({ ...formData, secret: e.target.value })}
                  placeholder="钉钉群机器人加签密钥，其他类型留空"
                />
              </div>
              <div className="grid gap-2">
                <Label>备注</Label>
                <Input
                  value={formData.remark}
                  onChange={(e) => setFormData({ ...formData, remark: e.target.value })}
                  placeholder="可选的备注信息"
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
