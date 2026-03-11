"use client";

import { motion } from "motion/react";
import { cn } from "@/lib/utils";
import { useNavStore } from "@/stores";
import { Home, Server, FolderGit2, History, Bell } from "lucide-react";

const navItems = [
  { id: "home", label: "首页", icon: Home },
  { id: "hosts", label: "主机", icon: Server },
  { id: "projects", label: "项目", icon: FolderGit2 },
  { id: "logs", label: "部署记录", icon: History },
  { id: "notifications", label: "通知渠道", icon: Bell },
];

export function NavBar() {
  const { activeView, setActiveView } = useNavStore();

  return (
    <div className="relative z-50 md:min-h-screen">
      <motion.nav
        aria-label="主导航"
        className={cn(
          "fixed bottom-6 left-1/2 -translate-x-1/2 flex items-center gap-1 p-3",
          "md:sticky md:top-30 md:left-auto md:bottom-auto md:translate-x-0 md:flex-col md:gap-3",
          "bg-sidebar text-sidebar-foreground border border-sidebar-border rounded-3xl custom-shadow"
        )}
        initial={{ opacity: 0, y: 20 }}
        animate={{ opacity: 1, y: 0 }}
        transition={{ duration: 0.3 }}
      >
        {navItems.map((item, index) => {
          const isActive = activeView === item.id;
          const Icon = item.icon;
          return (
            <motion.button
              key={item.id}
              type="button"
              onClick={() => setActiveView(item.id)}
              className={cn(
                "relative p-2 md:p-3 rounded-2xl z-20 transition-colors",
                isActive
                  ? "text-sidebar-primary-foreground"
                  : "text-sidebar-foreground/60 hover:bg-sidebar-accent"
              )}
              initial={{ opacity: 0, scale: 0.8 }}
              animate={{
                opacity: 1,
                scale: 1,
                transition: {
                  delay: index * 0.05,
                  duration: 0.3,
                },
              }}
              whileHover={{ scale: 1.1, zIndex: 30 }}
              whileTap={{ scale: 0.95 }}
            >
              {isActive && (
                <motion.div
                  layoutId="navbar-indicator"
                  className="absolute inset-0 bg-sidebar-primary rounded-2xl z-0"
                  transition={{ type: "spring", stiffness: 300, damping: 30 }}
                />
              )}
              <span className="relative z-10">
                <Icon strokeWidth={2} className="h-6 w-6" />
              </span>
            </motion.button>
          );
        })}
      </motion.nav>
    </div>
  );
}