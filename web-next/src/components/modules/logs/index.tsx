"use client";

import { useEffect, useState, useRef } from "react";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { runApi } from "@/api/client";
import type { PipelineRun } from "@/types";
import { formatDate, getStatusVariant, getStatusText, calculateDuration, formatDuration, formatShortDateTime } from "@/lib/utils";
import { X, Loader2 } from "lucide-react";
import { motion, AnimatePresence } from "motion/react";

// 全局缓存
let runsCache: PipelineRun[] | null = null;

function LogDialog({
  run,
  onClose,
}: {
  run: PipelineRun;
  onClose: () => void;
}) {
  const [logContent, setLogContent] = useState(run.log_text || "");
  const [loading, setLoading] = useState(false);
  const logRef = useRef<HTMLPreElement>(null);
  const eventSourceRef = useRef<EventSource | null>(null);

  useEffect(() => {
    // 关闭之前的 EventSource
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
    }

    // 如果是运行中的任务，开启流式日志
    if (run.status === "running") {
      setLoading(true);
      const es = new EventSource(`/api/v1/runs/${run.id}/log/stream`);
      es.onmessage = (event) => {
        setLogContent((prev) => prev + event.data);
        if (logRef.current) {
          logRef.current.scrollTop = logRef.current.scrollHeight;
        }
      };
      es.onerror = () => {
        es.close();
        setLoading(false);
      };
      eventSourceRef.current = es;
    }

    return () => {
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
      }
    };
  }, [run.id, run.status]);

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center">
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: 1 }}
        exit={{ opacity: 0 }}
        className="absolute inset-0 bg-black/50"
        onClick={onClose}
      />
      <motion.div
        initial={{ opacity: 0, scale: 0.95, y: 20 }}
        animate={{ opacity: 1, scale: 1, y: 0 }}
        exit={{ opacity: 0, scale: 0.95, y: 20 }}
        transition={{ type: "spring", stiffness: 300, damping: 30 }}
        className="relative w-[90vw] max-w-4xl h-[80vh] bg-card rounded-3xl border shadow-2xl flex flex-col overflow-hidden"
      >
        {/* 头部 */}
        <div className="flex items-center justify-between px-6 py-4 border-b shrink-0">
          <div className="flex items-center gap-4">
            <h2 className="text-lg font-semibold">{run.project_name}</h2>
            <Badge variant={getStatusVariant(run.status)}>
              {getStatusText(run.status)}
            </Badge>
          </div>
          <div className="flex items-center gap-4">
            <span className="text-sm text-muted-foreground">
              {run.branch} • {formatDate(run.started_at)}
            </span>
            <Button variant="ghost" size="icon" onClick={onClose}>
              <X className="h-5 w-5" />
            </Button>
          </div>
        </div>

        {/* 日志内容 */}
        <div className="flex-1 overflow-hidden bg-slate-900">
          {loading && (
            <div className="flex items-center justify-center h-12 border-b border-slate-700">
              <Loader2 className="h-5 w-5 animate-spin text-muted-foreground mr-2" />
              <span className="text-sm text-muted-foreground">实时获取日志中...</span>
            </div>
          )}
          <pre
            ref={logRef}
            className="h-full overflow-auto p-4 text-green-200 font-mono text-sm whitespace-pre-wrap"
          >
            {logContent || "暂无日志内容"}
          </pre>
        </div>
      </motion.div>
    </div>
  );
}

export function Logs() {
  const [runs, setRuns] = useState<PipelineRun[]>(runsCache || []);
  const [selectedRun, setSelectedRun] = useState<PipelineRun | null>(null);
  const loadingRef = useRef(false);

  const loadRuns = async (forceRefresh = false) => {
    if (!forceRefresh && runsCache) {
      setRuns(runsCache);
      return;
    }
    if (loadingRef.current) return;
    loadingRef.current = true;
    try {
      const data = await runApi.list();
      const runs = Array.isArray(data) ? data : [];
      runsCache = runs;
      setRuns(runs);
    } catch (error) {
      console.error(error);
    } finally {
      loadingRef.current = false;
    }
  };

  useEffect(() => {
    loadRuns();

    const handleRefresh = () => {
      runsCache = null; // 清除缓存以强制刷新
      loadRuns(true);
    };
    window.addEventListener("refresh-logs", handleRefresh);
    return () => window.removeEventListener("refresh-logs", handleRefresh);
  }, []);

  const handleSelectRun = (run: PipelineRun) => {
    setSelectedRun(run);
  };

  const handleCloseDialog = () => {
    setSelectedRun(null);
    loadRuns(); // 刷新列表以获取最新状态
  };

  return (
    <div className="space-y-2">
      {runs.length === 0 ? (
        <Card>
          <CardContent className="p-12 text-center text-muted-foreground">
            暂无部署记录
          </CardContent>
        </Card>
      ) : (
        <div className="space-y-2">
          {runs.map((run) => (
            <Card
              key={run.id}
              className="cursor-pointer hover:border-primary transition-colors"
              onClick={() => handleSelectRun(run)}
            >
              <CardContent className="p-4">
                <div className="flex items-center justify-between">
                  <div className="flex items-center gap-4 flex-1 min-w-0">
                    <span className="font-medium text-foreground shrink-0">#{run.id} {run.project_name}</span>
                    <Badge variant={getStatusVariant(run.status)} className="shrink-0">
                      {getStatusText(run.status)}
                    </Badge>
                    <span className="text-sm text-muted-foreground truncate">{run.commit_message}</span>
                  </div>
                  <div className="flex items-center gap-4 text-sm text-muted-foreground shrink-0 ml-4">
                    <span>{formatShortDateTime(run.started_at)} · {formatDuration(calculateDuration(run.started_at, run.finished_at))}</span>
                  </div>
                </div>
              </CardContent>
            </Card>
          ))}
        </div>
      )}

      {/* 日志详情对话框 */}
      <AnimatePresence>
        {selectedRun && (
          <LogDialog run={selectedRun} onClose={handleCloseDialog} />
        )}
      </AnimatePresence>
    </div>
  );
}
