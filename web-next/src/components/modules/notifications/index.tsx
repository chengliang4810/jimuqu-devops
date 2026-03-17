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
import { Badge } from "@/components/ui/badge";
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { notifyApi } from "@/api/client";
import type { NotifyChannel, NotifyChannelDetail, NotifyChannelType } from "@/types";
import { cn } from "@/lib/utils";
import { toast } from "sonner";
import { GripVertical, Pencil, Send, Star, Trash2, X } from "lucide-react";
import { motion, AnimatePresence } from "motion/react";
import { useToolbarSearchStore } from "@/components/modules/toolbar/search-store";
import { useToolbarViewOptionsStore } from "@/components/modules/toolbar/view-options-store";

const channelTypes: { value: NotifyChannelType; label: string }[] = [
  { value: "webhook", label: "Webhook" },
  { value: "dingtalk", label: "钉钉" },
  { value: "wechat", label: "企业微信" },
  { value: "feishu", label: "飞书" },
];

type NotificationFormState = {
  name: string;
  type: NotifyChannelType;
  webhook_url: string;
  secret: string;
  remark: string;
};

// 全局缓存
let channelsCache: NotifyChannel[] | null = null;

type NotificationCardViewProps = {
  channel: NotifyChannel;
  dragHandle?: React.ReactNode;
  isDeleting: boolean;
  isDragging?: boolean;
  isOverlay?: boolean;
  typeLabel: string;
  onSetDefault: () => void;
  onTest: () => void;
  onEdit: () => void;
  onDeleteRequest: () => void;
  onDeleteCancel: () => void;
  onDeleteConfirm: () => void;
};

function NotificationCardView({
  channel,
  dragHandle,
  isDeleting,
  isDragging = false,
  isOverlay = false,
  typeLabel,
  onSetDefault,
  onTest,
  onEdit,
  onDeleteRequest,
  onDeleteCancel,
  onDeleteConfirm,
}: NotificationCardViewProps) {
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
            <div className="mb-1 flex items-center gap-2">
              <Badge variant="outline">{typeLabel}</Badge>
            </div>
            <Tooltip>
              <TooltipTrigger asChild>
                <h3 className="truncate font-semibold text-foreground">{channel.name}</h3>
              </TooltipTrigger>
              <TooltipContent>{channel.name}</TooltipContent>
            </Tooltip>
          </div>
          <div className="flex shrink-0 items-center gap-1">
            {dragHandle}
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant="ghost"
                  size="icon"
                  className={cn(
                    "h-8 w-8",
                    channel.is_default && "text-amber-600 hover:text-amber-600"
                  )}
                  aria-label={channel.is_default ? "当前默认渠道" : "设为默认渠道"}
                  onClick={channel.is_default ? undefined : onSetDefault}
                >
                  <Star className={cn("h-4 w-4", channel.is_default && "fill-current")} />
                </Button>
              </TooltipTrigger>
              <TooltipContent>{channel.is_default ? "当前默认渠道" : "设为默认渠道"}</TooltipContent>
            </Tooltip>
            <Tooltip>
              <TooltipTrigger asChild>
                <Button
                  variant="ghost"
                  size="icon"
                  className="h-8 w-8"
                  aria-label="测试发送"
                  onClick={onTest}
                >
                  <Send className="h-4 w-4" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>测试发送</TooltipContent>
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

        <p className="truncate text-sm text-muted-foreground">
          {channel.remark || "未填写备注"}
        </p>
      </CardContent>
    </Card>
  );
}

type SortableNotificationCardProps = {
  channel: NotifyChannel;
  typeLabel: string;
  isDeleting: boolean;
  onSetDefault: () => void;
  onTest: () => void;
  onEdit: () => void;
  onDeleteRequest: () => void;
  onDeleteCancel: () => void;
  onDeleteConfirm: () => void;
};

function SortableNotificationCard(props: SortableNotificationCardProps) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: props.channel.id,
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
      <NotificationCardView
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

function getWebhookUrl(channel: NotifyChannelDetail): string {
  if (channel.type === "webhook") {
    return typeof channel.config.url === "string" ? channel.config.url : "";
  }

  return typeof channel.config.webhook_url === "string" ? channel.config.webhook_url : "";
}

function getSecret(channel: NotifyChannelDetail): string {
  return typeof channel.config.secret === "string" ? channel.config.secret : "";
}

function buildNotificationPayload(formData: NotificationFormState) {
  const config =
    formData.type === "webhook"
      ? {
          url: formData.webhook_url,
          ...(formData.secret ? { secret: formData.secret } : {}),
        }
      : {
          webhook_url: formData.webhook_url,
          ...(formData.secret ? { secret: formData.secret } : {}),
        };

  return {
    name: formData.name,
    type: formData.type,
    remark: formData.remark,
    config,
  };
}

export function Notifications() {
  const [channels, setChannels] = useState<NotifyChannel[]>(channelsCache || []);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingChannel, setEditingChannel] = useState<NotifyChannel | null>(null);
  const [deletingChannel, setDeletingChannel] = useState<NotifyChannel | null>(null);
  const [activeChannelId, setActiveChannelId] = useState<number | null>(null);
  const loadingRef = useRef(false);
  const [formData, setFormData] = useState<NotificationFormState>({
    name: "",
    type: "webhook",
    webhook_url: "",
    secret: "",
    remark: "",
  });
  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: { distance: 2 },
    }),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  );
  const searchTerm = useToolbarSearchStore((state) => state.searchTerms.notifications || "");
  const notificationFilter = useToolbarViewOptionsStore((state) => state.notificationFilter);
  const sortOrder = useToolbarViewOptionsStore((state) => state.getSortOrder("notifications"));
  const deferredSearchTerm = useDeferredValue(searchTerm);

  const loadChannels = async (forceRefresh = false) => {
    if (!forceRefresh && channelsCache) {
      setChannels(channelsCache);
      return;
    }
    if (loadingRef.current) return;
    loadingRef.current = true;
    try {
      const data = await notifyApi.list();
      const nextChannels = Array.isArray(data) ? data : [];
      channelsCache = nextChannels;
      setChannels(nextChannels);
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

  const openEditDialog = async (channel: NotifyChannel) => {
    try {
      const detail = await notifyApi.get(channel.id);
      setEditingChannel(channel);
      setFormData({
        name: detail.name,
        type: detail.type,
        webhook_url: getWebhookUrl(detail),
        secret: getSecret(detail),
        remark: detail.remark || "",
      });
      setDialogOpen(true);
    } catch (error: any) {
      toast.error(error.message || "加载渠道详情失败");
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    try {
      const payload = buildNotificationPayload(formData);
      if (editingChannel) {
        await notifyApi.update(editingChannel.id, payload);
        toast.success("通知渠道更新成功");
      } else {
        await notifyApi.create(payload);
        toast.success("通知渠道创建成功");
      }
      channelsCache = null;
      setDialogOpen(false);
      await loadChannels(true);
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
      await loadChannels(true);
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

  const handleSetDefault = async (channel: NotifyChannel) => {
    try {
      await notifyApi.setDefault(channel.id);
      const nextChannels = channels.map((item) => ({
        ...item,
        is_default: item.id === channel.id,
      }));
      channelsCache = nextChannels;
      setChannels(nextChannels);
      toast.success("默认通知渠道已更新");
    } catch (error: any) {
      toast.error(error.message || "设置默认渠道失败");
    }
  };

  const handleDragStart = (event: DragStartEvent) => {
    setActiveChannelId(Number(event.active.id));
  };

  const handleDragCancel = () => {
    setActiveChannelId(null);
  };

  const activeChannel =
    activeChannelId == null ? null : channels.find((item) => item.id === activeChannelId) ?? null;

  const normalizedSearchTerm = deferredSearchTerm.trim().toLowerCase();
  const visibleChannels = channels
    .filter((channel) => {
      if (notificationFilter === "default") {
        return channel.is_default === true;
      }
      if (notificationFilter !== "all") {
        return channel.type === notificationFilter;
      }
      return true;
    })
    .filter((channel) => {
      if (!normalizedSearchTerm) {
        return true;
      }

      return channel.name.toLowerCase().includes(normalizedSearchTerm);
    });

  if (sortOrder === "name-asc") {
    visibleChannels.sort((left, right) => left.name.localeCompare(right.name, "zh-CN"));
  } else if (sortOrder === "name-desc") {
    visibleChannels.sort((left, right) => right.name.localeCompare(left.name, "zh-CN"));
  }

  const reorderEnabled =
    sortOrder === "manual" &&
    notificationFilter === "all" &&
    normalizedSearchTerm.length === 0;

  const handleDragEnd = async (event: DragEndEvent) => {
    const { active, over } = event;
    setActiveChannelId(null);

    if (!over || active.id === over.id) {
      return;
    }

    const oldIndex = channels.findIndex((item) => item.id === Number(active.id));
    const newIndex = channels.findIndex((item) => item.id === Number(over.id));
    if (oldIndex === -1 || newIndex === -1 || oldIndex === newIndex) {
      return;
    }

    const nextChannels = arrayMove(channels, oldIndex, newIndex);
    const previousChannels = channels;
    channelsCache = nextChannels;
    setChannels(nextChannels);

    try {
      await notifyApi.reorder(nextChannels.map((item) => item.id));
    } catch (error: any) {
      channelsCache = previousChannels;
      setChannels(previousChannels);
      toast.error(error.message || "通知渠道排序保存失败");
    }
  };

  const getTypeLabel = (type: NotifyChannelType) => {
    return channelTypes.find((item) => item.value === type)?.label || type;
  };

  return (
    <div className="space-y-4">
      {channels.length === 0 ? (
        <Card>
          <CardContent className="p-12 text-center text-muted-foreground">
            暂无通知渠道，点击上方按钮新增渠道。
          </CardContent>
        </Card>
      ) : visibleChannels.length === 0 ? (
        <Card>
          <CardContent className="p-12 text-center text-muted-foreground">
            没有找到符合当前查询条件的通知渠道。
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
            items={visibleChannels.map((channel) => channel.id)}
            strategy={rectSortingStrategy}
          >
            <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
              {visibleChannels.map((channel) => (
                <SortableNotificationCard
                  key={channel.id}
                  channel={channel}
                  typeLabel={getTypeLabel(channel.type)}
                  isDeleting={deletingChannel?.id === channel.id}
                  onSetDefault={() => void handleSetDefault(channel)}
                  onTest={() => void handleTest(channel)}
                  onEdit={() => void openEditDialog(channel)}
                  onDeleteRequest={() => setDeletingChannel(channel)}
                  onDeleteCancel={() => setDeletingChannel(null)}
                  onDeleteConfirm={() => void handleDelete(channel)}
                />
              ))}
            </div>
          </SortableContext>
          <DragOverlay>
            {activeChannel ? (
              <div className="w-full">
                <NotificationCardView
                  channel={activeChannel}
                  typeLabel={getTypeLabel(activeChannel.type)}
                  isDeleting={false}
                  isDragging
                  isOverlay
                  onSetDefault={() => {}}
                  onTest={() => {}}
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
        <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
          {visibleChannels.map((channel) => (
            <NotificationCardView
              key={channel.id}
              channel={channel}
              typeLabel={getTypeLabel(channel.type)}
              isDeleting={deletingChannel?.id === channel.id}
              onSetDefault={() => void handleSetDefault(channel)}
              onTest={() => void handleTest(channel)}
              onEdit={() => void openEditDialog(channel)}
              onDeleteRequest={() => setDeletingChannel(channel)}
              onDeleteCancel={() => setDeletingChannel(null)}
              onDeleteConfirm={() => void handleDelete(channel)}
            />
          ))}
        </div>
      )}

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="sm:max-w-[500px]">
          <DialogHeader>
            <DialogTitle>{editingChannel ? "编辑通知渠道" : "新增通知渠道"}</DialogTitle>
            <DialogDescription className="sr-only">
              配置通知渠道类型、Webhook 地址和备注信息。
            </DialogDescription>
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
                  onValueChange={(value) =>
                    setFormData({
                      ...formData,
                      type: value as NotifyChannelType,
                      secret: value === "webhook" || value === "dingtalk" ? formData.secret : "",
                    })
                  }
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
                  placeholder="https://example.com/webhook"
                  type="url"
                  required
                />
              </div>
              <div className="grid gap-2">
                <Label>签名密钥（可选）</Label>
                <Input
                  value={formData.secret}
                  onChange={(e) => setFormData({ ...formData, secret: e.target.value })}
                  placeholder="Webhook 或钉钉签名密钥，其他类型留空"
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
