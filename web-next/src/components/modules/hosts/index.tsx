"use client";

import { useEffect, useState, useRef } from "react";
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
import { Tooltip, TooltipContent, TooltipTrigger } from "@/components/ui/tooltip";
import { hostApi } from "@/api/client";
import type { Host } from "@/types";
import { cn } from "@/lib/utils";
import { toast } from "sonner";
import { GripVertical, Pencil, Trash2, X } from "lucide-react";
import { motion, AnimatePresence } from "motion/react";

// 全局请求缓存
let hostsCache: Host[] | null = null;

type HostCardViewProps = {
  host: Host;
  dragHandle?: React.ReactNode;
  isDeleting: boolean;
  isDragging?: boolean;
  isOverlay?: boolean;
  onEdit: () => void;
  onDeleteRequest: () => void;
  onDeleteCancel: () => void;
  onDeleteConfirm: () => void;
};

function HostCardView({
  host,
  dragHandle,
  isDeleting,
  isDragging = false,
  isOverlay = false,
  onEdit,
  onDeleteRequest,
  onDeleteCancel,
  onDeleteConfirm,
}: HostCardViewProps) {
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
                <h3 className="truncate font-semibold text-foreground">{host.name}</h3>
              </TooltipTrigger>
              <TooltipContent>{host.name}</TooltipContent>
            </Tooltip>
          </div>
          <div className="flex shrink-0 items-center gap-1">
            {dragHandle}
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

        <div className="space-y-1 text-sm text-muted-foreground">
          <p>
            {host.address}:{host.port}
          </p>
          <p>{host.username}</p>
        </div>
      </CardContent>
    </Card>
  );
}

type SortableHostCardProps = {
  host: Host;
  isDeleting: boolean;
  onEdit: () => void;
  onDeleteRequest: () => void;
  onDeleteCancel: () => void;
  onDeleteConfirm: () => void;
};

function SortableHostCard(props: SortableHostCardProps) {
  const { attributes, listeners, setNodeRef, transform, transition, isDragging } = useSortable({
    id: props.host.id,
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
      <HostCardView
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

export function Hosts() {
  const [hosts, setHosts] = useState<Host[]>(hostsCache || []);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingHost, setEditingHost] = useState<Host | null>(null);
  const [deletingHost, setDeletingHost] = useState<Host | null>(null);
  const [activeHostId, setActiveHostId] = useState<number | null>(null);
  const loadingRef = useRef(false);
  const [formData, setFormData] = useState({
    name: "",
    address: "",
    port: 22,
    username: "",
    password: "",
  });
  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: { distance: 2 },
    }),
    useSensor(KeyboardSensor, {
      coordinateGetter: sortableKeyboardCoordinates,
    })
  );

  const loadHosts = async (forceRefresh = false) => {
    if (loadingRef.current) return;
    if (!forceRefresh && hostsCache) {
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
      hostsCache = null;
      setDialogOpen(false);
      await loadHosts(true);
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

  const handleDragStart = (event: DragStartEvent) => {
    setActiveHostId(Number(event.active.id));
  };

  const handleDragCancel = () => {
    setActiveHostId(null);
  };

  const activeHost = activeHostId == null ? null : hosts.find((item) => item.id === activeHostId) ?? null;

  const handleDragEnd = async (event: DragEndEvent) => {
    const { active, over } = event;
    setActiveHostId(null);

    if (!over || active.id === over.id) {
      return;
    }

    const oldIndex = hosts.findIndex((item) => item.id === Number(active.id));
    const newIndex = hosts.findIndex((item) => item.id === Number(over.id));
    if (oldIndex === -1 || newIndex === -1 || oldIndex === newIndex) {
      return;
    }

    const nextHosts = arrayMove(hosts, oldIndex, newIndex);
    const previousHosts = hosts;
    hostsCache = nextHosts;
    setHosts(nextHosts);

    try {
      await hostApi.reorder(nextHosts.map((item) => item.id));
    } catch (error: any) {
      hostsCache = previousHosts;
      setHosts(previousHosts);
      toast.error(error.message || "主机排序保存失败");
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
        <DndContext
          sensors={sensors}
          collisionDetection={closestCenter}
          modifiers={[restrictToParentElement]}
          onDragStart={handleDragStart}
          onDragCancel={handleDragCancel}
          onDragEnd={(event) => void handleDragEnd(event)}
        >
          <SortableContext items={hosts.map((host) => host.id)} strategy={rectSortingStrategy}>
            <div className="grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-3">
              {hosts.map((host) => (
                <SortableHostCard
                  key={host.id}
                  host={host}
                  isDeleting={deletingHost?.id === host.id}
                  onEdit={() => openEditDialog(host)}
                  onDeleteRequest={() => setDeletingHost(host)}
                  onDeleteCancel={() => setDeletingHost(null)}
                  onDeleteConfirm={() => void handleDelete(host)}
                />
              ))}
            </div>
          </SortableContext>
          <DragOverlay>
            {activeHost ? (
              <div className="w-full">
                <HostCardView
                  host={activeHost}
                  isDeleting={false}
                  isDragging
                  isOverlay
                  onEdit={() => {}}
                  onDeleteRequest={() => {}}
                  onDeleteCancel={() => {}}
                  onDeleteConfirm={() => {}}
                />
              </div>
            ) : null}
          </DragOverlay>
        </DndContext>
      )}

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="sm:max-w-[425px]">
          <DialogHeader>
            <DialogTitle>{editingHost ? "编辑主机" : "新增主机"}</DialogTitle>
            <DialogDescription className="sr-only">
              填写主机连接信息并保存到部署目标列表。
            </DialogDescription>
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
