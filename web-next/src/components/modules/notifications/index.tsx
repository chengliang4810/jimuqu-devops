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
import { Plus, Pencil, Trash2, X, Send } from "lucide-react";
import { motion, AnimatePresence } from "motion/react";

const channelTypes: { value: NotifyChannelType; label: string }[] = [
  { value: "webhook", label: "Webhook" },
  { value: "dingtalk", label: "钉钉" },
  { value: "wechat", label: "企业微信" },
  { value: "feishu", label: "飞书" },
];

// 全局缓存
let channelsCache: NotifyChannel[] | null = null;

export function Notifications() {
  const [channels, setChannels] = useState<NotifyChannel[]>(channelsCache || []);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingChannel, setEditingChannel] = useState<NotifyChannel | null>(null);
  const [deletingChannel, setDeletingChannel] = useState<NotifyChannel | null>(null);
  const loadingRef = useRef(false);
  const [formData, setFormData] = useState({
    name: "",
    type: "webhook" as NotifyChannelType,
    webhook_url: "",
    secret: "",
    remark: "",
  });

  const loadChannels = async () => {
    if (channelsCache) {
      setChannels(channelsCache);
      return;
    }
    if (loadingRef.current) return;
    loadingRef.current = true;
    try {
      const data = await notifyApi.list();
      const channels = Array.isArray(data) ? data : [];
      channelsCache = channels;
      setChannels(channels);
    } catch (error) {
      console.error(error);
    } finally {
      loadingRef.current = false;
    }
  };

  useEffect(() => {
    loadChannels();

    const handleOpenDialog = (e: CustomEvent) => {
      if (e.detail.mode === "create") {
        openCreateDialog();
      }
    };
    window.addEventListener("open-notify-dialog", handleOpenDialog as EventListener);
    return () => window.removeEventListener("open-notify-dialog", handleOpenDialog as EventListener);
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

  const handleDelete = async (channel: NotifyChannel) => {
    try {
      await notifyApi.delete(channel.id);
      channelsCache = null;
      toast.success("通知渠道删除成功");
      setDeletingChannel(null);
      loadChannels();
    } catch (error: any) {
      toast.error(error.message || "删除失败");
    }
  };

  const handleTest = async (channel: NotifyChannel) => {
    try {
      await notifyApi.test(channel.id);
      toast.success("测试通知发送成功");
    } catch (error: any) {
      toast.error(error.message || "测试通知发送失败");
    }
  };

  const getTypeLabel = (type: NotifyChannelType) => {
    return channelTypes.find((t) => t.value === type)?.label || type;
  };

  return (
    <div className="space-y-4">
      {channels.length === 0 ? (
        <Card>
          <CardContent className="p-12 text-center text-muted-foreground">
            暂无通知渠道，点击上方按钮新增渠道。
          </CardContent>
        </Card>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
          {channels.map((channel) => (
            <Card key={channel.id} className="overflow-hidden">
              <CardContent className="p-4">
                <header className="flex items-start justify-between mb-3 relative">
                  <div className="flex-1 mr-2 min-w-0">
                    <Badge variant="outline" className="mb-1">{getTypeLabel(channel.type)}</Badge>
                    <h3 className="font-semibold text-foreground truncate">{channel.name}</h3>
                  </div>
                  <div className="flex items-center gap-1 shrink-0">
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-8 w-8"
                      onClick={() => handleTest(channel)}
                    >
                      <Send className="h-4 w-4" />
                    </Button>
                    <Button
                      variant="ghost"
                      size="icon"
                      className="h-8 w-8"
                      onClick={() => openEditDialog(channel)}
                    >
                      <Pencil className="h-4 w-4" />
                    </Button>
                    {deletingChannel?.id !== channel.id ? (
                      <Button
                        variant="ghost"
                        size="icon"
                        className="h-8 w-8 text-muted-foreground hover:text-destructive"
                        onClick={() => setDeletingChannel(channel)}
                      >
                        <Trash2 className="h-4 w-4" />
                      </Button>
                    ) : null}
                  </div>

                  <AnimatePresence>
                    {deletingChannel?.id === channel.id && (
                      <motion.div
                        initial={{ opacity: 0 }}
                        animate={{ opacity: 1 }}
                        exit={{ opacity: 0 }}
                        className="absolute inset-0 flex items-center justify-center gap-2 bg-destructive p-2 rounded-lg"
                      >
                        <button
                          type="button"
                          onClick={() => setDeletingChannel(null)}
                          className="flex h-7 w-7 items-center justify-center rounded-lg bg-white/20 text-white transition-all hover:bg-white/30 active:scale-95"
                        >
                          <X className="h-4 w-4" />
                        </button>
                        <button
                          type="button"
                          onClick={() => handleDelete(channel)}
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
                  {channel.webhook_url}
                </p>
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
