"use client";

import { useState, useEffect } from "react";
import { motion, AnimatePresence } from "motion/react";
import { useNavStore } from "@/stores";
import { isAuthenticated } from "@/api/client";
import { NavBar } from "@/components/modules/navbar/navbar";
import { Login } from "@/components/modules/login";
import { Home } from "@/components/modules/home";
import { Hosts } from "@/components/modules/hosts";
import { Projects } from "@/components/modules/projects";
import { Logs } from "@/components/modules/logs";
import { Notifications } from "@/components/modules/notifications";
import { Toaster } from "@/components/ui/sonner";
import { Button } from "@/components/ui/button";
import Logo from "@/components/modules/logo";
import { LogOut } from "lucide-react";

const viewTitles: Record<string, string> = {
  home: "首页",
  hosts: "主机管理",
  projects: "项目管理",
  logs: "部署记录",
  notifications: "通知渠道",
};

export function App() {
  const [authenticated, setAuthenticated] = useState(false);
  const [loading, setLoading] = useState(true);
  const { activeView } = useNavStore();

  useEffect(() => {
    setAuthenticated(isAuthenticated());
    setLoading(false);
  }, []);

  const handleLogout = () => {
    localStorage.removeItem("jwt_token");
    window.location.reload();
  };

  if (loading) {
    return (
      <div className="min-h-screen flex items-center justify-center bg-background">
        <div className="text-muted-foreground">加载中...</div>
      </div>
    );
  }

  if (!authenticated) {
    return (
      <>
        <Login onSuccess={() => setAuthenticated(true)} />
        <Toaster />
      </>
    );
  }

  const renderView = () => {
    switch (activeView) {
      case "home":
        return <Home />;
      case "hosts":
        return <Hosts />;
      case "projects":
        return <Projects />;
      case "logs":
        return <Logs />;
      case "notifications":
        return <Notifications />;
      default:
        return <Home />;
    }
  };

  return (
    <div className="mx-auto flex h-dvh max-w-6xl flex-col overflow-hidden px-3 md:grid md:grid-cols-[auto_1fr] md:gap-6 md:px-6">
      <NavBar />
      <main className="flex min-h-0 w-full min-w-0 flex-1 flex-col">
        <header className="my-6 flex flex-none items-center gap-x-2 px-2">
          <Logo size={48} />
          <div className="flex-1 overflow-hidden">
            <AnimatePresence mode="wait">
              <motion.div
                key={activeView}
                initial={{ y: 20, opacity: 0 }}
                animate={{ y: 0, opacity: 1 }}
                exit={{ y: -20, opacity: 0 }}
                transition={{ duration: 0.2 }}
                className="flex items-center"
              >
                <span className="text-3xl font-bold mt-1">{viewTitles[activeView] || "首页"}</span>
              </motion.div>
            </AnimatePresence>
          </div>
          <div className="ml-auto">
            <Button variant="outline" size="sm" onClick={handleLogout}>
              <LogOut className="h-4 w-4 mr-2" />
              退出登录
            </Button>
          </div>
        </header>
        <AnimatePresence mode="wait" initial={false}>
          <motion.div
            key={activeView}
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            transition={{ duration: 0.2 }}
            className="h-full min-h-0 flex-1 overflow-y-auto pb-6"
          >
            {renderView()}
          </motion.div>
        </AnimatePresence>
      </main>
      <Toaster />
    </div>
  );
}