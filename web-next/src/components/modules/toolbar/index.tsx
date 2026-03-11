"use client";

import { motion, AnimatePresence } from "motion/react";
import { useNavStore } from "@/stores";
import { Button } from "@/components/ui/button";
import { useState } from "react";
import { Plus, RefreshCw } from "lucide-react";

function HostsToolbar() {
  const handleAdd = () => {
    window.dispatchEvent(new CustomEvent("open-host-dialog", { detail: { mode: "create" } }));
  };

  return (
    <Button size="sm" onClick={handleAdd}>
      <Plus className="h-4 w-4 mr-2" />
      新增主机
    </Button>
  );
}

function ProjectsToolbar() {
  const handleAdd = () => {
    window.dispatchEvent(new CustomEvent("open-project-dialog", { detail: { mode: "create" } }));
  };

  return (
    <Button size="sm" onClick={handleAdd}>
      <Plus className="h-4 w-4 mr-2" />
      新增项目
    </Button>
  );
}

function LogsToolbar() {
  const [loading, setLoading] = useState(false);

  const handleRefresh = () => {
    setLoading(true);
    window.dispatchEvent(new CustomEvent("refresh-logs"));
    setTimeout(() => setLoading(false), 1000);
  };

  return (
    <Button variant="outline" size="sm" onClick={handleRefresh} disabled={loading}>
      <RefreshCw className={`h-4 w-4 mr-2 ${loading ? "animate-spin" : ""}`} />
      刷新记录
    </Button>
  );
}

function NotificationsToolbar() {
  const handleAdd = () => {
    window.dispatchEvent(new CustomEvent("open-notify-dialog", { detail: { mode: "create" } }));
  };

  return (
    <Button size="sm" onClick={handleAdd}>
      <Plus className="h-4 w-4 mr-2" />
      新增渠道
    </Button>
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
