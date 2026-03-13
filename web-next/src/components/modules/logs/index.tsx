"use client";

import { useEffect, useState, useRef, type MouseEvent } from "react";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { runApi } from "@/api/client";
import type { PipelineRun } from "@/types";
import { getApiOrigin } from "@/lib/api-base";
import { formatDate, getStatusVariant, getStatusText, calculateDuration, formatDuration, formatShortDateTime } from "@/lib/utils";
import { X, Square } from "lucide-react";
import { motion, AnimatePresence } from "motion/react";
import { useNavStore } from "@/stores";
import { toast } from "sonner";

// 全局缓存
let runsCache: PipelineRun[] | null = null;

function LogDialog({
  run,
  onClose,
  onCancelRequest,
  onCancelConfirm,
  onCancelDismiss,
  cancelling,
  confirmingCancel,
}: {
  run: PipelineRun;
  onClose: () => void;
  onCancelRequest: (run: PipelineRun) => void;
  onCancelConfirm: (run: PipelineRun) => void;
  onCancelDismiss: () => void;
  cancelling: boolean;
  confirmingCancel: boolean;
}) {
  const [currentRun, setCurrentRun] = useState(run);
  const [logContent, setLogContent] = useState(run.log_text || "");
  const logRef = useRef<HTMLDivElement>(null);
  const logContentRef = useRef<HTMLPreElement>(null);
  const eventSourceRef = useRef<EventSource | null>(null);

  useEffect(() => {
    const frameId = window.requestAnimationFrame(() => {
      if (logRef.current) {
        logRef.current.scrollTop = logRef.current.scrollHeight;
      }
    });

    return () => window.cancelAnimationFrame(frameId);
  }, [logContent]);

  useEffect(() => {
    setCurrentRun(run);
    setLogContent(run.log_text || "");

    // 关闭之前的 EventSource
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
    }

    // 如果是运行中的任务，开启流式日志
    if (run.status === "running") {
      const token = typeof window !== "undefined" ? localStorage.getItem("jwt_token") : null;
      const query = token ? `?token=${encodeURIComponent(token)}` : "";
      const es = new EventSource(`${getApiOrigin()}/api/v1/runs/${run.id}/stream${query}`);
      es.addEventListener("run", (event) => {
        const nextRun = JSON.parse((event as MessageEvent<string>).data) as PipelineRun;
        setCurrentRun(nextRun);
        setLogContent(nextRun.log_text || "");
        if (nextRun.status !== "running") {
          es.close();
        }
      });
      es.onerror = () => {
        es.close();
      };
      eventSourceRef.current = es;
    }

    return () => {
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
      }
    };
  }, [run.id, run.status]);

  useEffect(() => {
    const handleSelectAllLogs = (event: KeyboardEvent) => {
      if (!(event.ctrlKey || event.metaKey) || event.altKey || event.shiftKey) {
        return;
      }
      if (event.key.toLowerCase() !== "a") {
        return;
      }

      const target = event.target as HTMLElement | null;
      if (target?.closest("input, textarea, [contenteditable='true']")) {
        return;
      }

      if (!logContentRef.current) {
        return;
      }

      event.preventDefault();
      const selection = window.getSelection();
      if (!selection) {
        return;
      }

      const range = document.createRange();
      range.selectNodeContents(logContentRef.current);
      selection.removeAllRanges();
      selection.addRange(range);
    };

    document.addEventListener("keydown", handleSelectAllLogs);
    return () => document.removeEventListener("keydown", handleSelectAllLogs);
  }, []);

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
            <h2 className="text-lg font-semibold">
              #{currentRun.id} {currentRun.project_name}
            </h2>
            <Badge variant={getStatusVariant(currentRun.status)}>
              {getStatusText(currentRun.status)}
            </Badge>
            {(currentRun.status === "running" || currentRun.status === "queued") ? (
              confirmingCancel ? (
                <div className="flex items-center gap-2">
                  <Button
                    type="button"
                    variant="destructive"
                    size="sm"
                    onClick={() => onCancelConfirm(currentRun)}
                    disabled={cancelling}
                  >
                    {cancelling ? "取消中" : "确认取消"}
                  </Button>
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={onCancelDismiss}
                    disabled={cancelling}
                  >
                    再想想
                  </Button>
                </div>
              ) : (
                <Button
                  type="button"
                  variant="destructive"
                  size="sm"
                  onClick={() => onCancelRequest(currentRun)}
                  disabled={cancelling}
                >
                  <Square className="mr-2 h-4 w-4" />
                  取消部署
                </Button>
              )
            ) : null}
          </div>
          <div className="flex items-center gap-4">
            <span className="text-sm text-muted-foreground">
              {currentRun.branch} • {formatDate(currentRun.started_at)}
            </span>
            <Button variant="ghost" size="icon" onClick={onClose}>
              <X className="h-5 w-5" />
            </Button>
          </div>
        </div>

        {/* 日志内容 */}
        <div className="flex-1 overflow-hidden bg-slate-900">
          <div
            ref={logRef}
            className="h-full overflow-y-scroll overflow-x-auto pr-1 [scrollbar-width:auto] [-ms-overflow-style:auto] [&::-webkit-scrollbar]:w-3 [&::-webkit-scrollbar]:h-3 [&::-webkit-scrollbar-thumb]:rounded-full [&::-webkit-scrollbar-thumb]:bg-slate-600 [&::-webkit-scrollbar-track]:bg-slate-800/80"
          >
            <pre
              ref={logContentRef}
              className="min-h-full p-4 text-sm whitespace-pre-wrap font-mono text-green-200 select-text"
            >
              {logContent || "暂无日志内容"}
            </pre>
          </div>
        </div>
      </motion.div>
    </div>
  );
}

export function Logs() {
  const { pendingRunId, clearPendingRunId } = useNavStore();
  const [runs, setRuns] = useState<PipelineRun[]>(runsCache || []);
  const [selectedRun, setSelectedRun] = useState<PipelineRun | null>(null);
  const [cancellingRunId, setCancellingRunId] = useState<number | null>(null);
  const [confirmingCancelRunId, setConfirmingCancelRunId] = useState<number | null>(null);
  const loadingRef = useRef(false);
  const openingRunIdRef = useRef<number | null>(null);

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
    void loadRuns(Boolean(pendingRunId));

    const handleRefresh = () => {
      runsCache = null; // 清除缓存以强制刷新
      void loadRuns(true);
    };
    window.addEventListener("refresh-logs", handleRefresh);
    return () => window.removeEventListener("refresh-logs", handleRefresh);
  }, [pendingRunId]);

  useEffect(() => {
    if (!pendingRunId || openingRunIdRef.current === pendingRunId) {
      return;
    }

    const matchedRun = runs.find((run) => run.id === pendingRunId);
    if (matchedRun) {
      setSelectedRun(matchedRun);
      clearPendingRunId();
      return;
    }

    openingRunIdRef.current = pendingRunId;

    void (async () => {
      try {
        const run = await runApi.get(pendingRunId);
        runsCache = [run, ...(runsCache ?? []).filter((item) => item.id !== run.id)];
        setRuns((prevRuns) => [run, ...prevRuns.filter((item) => item.id !== run.id)]);
        setSelectedRun(run);
      } catch (error) {
        console.error(error);
      } finally {
        openingRunIdRef.current = null;
        clearPendingRunId();
      }
    })();
  }, [runs, pendingRunId, clearPendingRunId]);

  const handleSelectRun = (run: PipelineRun) => {
    setConfirmingCancelRunId(null);
    setSelectedRun(run);
  };

  const handleCancelRequest = (run: PipelineRun, event?: MouseEvent<HTMLButtonElement>) => {
    event?.stopPropagation();
    setConfirmingCancelRunId(run.id);
  };

  const handleCancelDismiss = (event?: MouseEvent<HTMLButtonElement>) => {
    event?.stopPropagation();
    setConfirmingCancelRunId(null);
  };

  const handleCancelRun = async (run: PipelineRun, event?: MouseEvent<HTMLButtonElement>) => {
    event?.stopPropagation();
    if (cancellingRunId === run.id) {
      return;
    }

    setCancellingRunId(run.id);
    try {
      const nextRun = await runApi.cancel(run.id);
      toast.success("部署任务已取消");
      setConfirmingCancelRunId(null);
      runsCache = null;
      await loadRuns(true);
      if (selectedRun?.id === run.id) {
        setSelectedRun(nextRun);
      }
    } catch (error: any) {
      toast.error(error.message || "取消部署失败");
    } finally {
      setCancellingRunId(null);
    }
  };

  const handleCloseDialog = () => {
    setSelectedRun(null);
    setConfirmingCancelRunId(null);
    runsCache = null;
    void loadRuns(true);
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
                    <span className="font-medium text-foreground shrink-0">
                      #{run.id} {run.project_name}
                    </span>
                    <Badge variant={getStatusVariant(run.status)} className="shrink-0">
                      {getStatusText(run.status)}
                    </Badge>
                    <span className="text-sm text-muted-foreground truncate">{run.commit_message}</span>
                  </div>
                  <div className="flex items-center gap-4 text-sm text-muted-foreground shrink-0 ml-4">
                    <span>{formatShortDateTime(run.started_at)} · {formatDuration(calculateDuration(run.started_at, run.finished_at))}</span>
                    {(run.status === "running" || run.status === "queued") ? (
                      confirmingCancelRunId === run.id ? (
                        <div className="flex items-center gap-2">
                          <Button
                            type="button"
                            size="sm"
                            variant="destructive"
                            onClick={(event) => void handleCancelRun(run, event)}
                            disabled={cancellingRunId === run.id}
                          >
                            {cancellingRunId === run.id ? "取消中" : "确认取消"}
                          </Button>
                          <Button
                            type="button"
                            size="sm"
                            variant="outline"
                            onClick={handleCancelDismiss}
                            disabled={cancellingRunId === run.id}
                          >
                            再想想
                          </Button>
                        </div>
                      ) : (
                        <Button
                          type="button"
                          size="sm"
                          variant="destructive"
                          onClick={(event) => handleCancelRequest(run, event)}
                          disabled={cancellingRunId === run.id}
                        >
                          <Square className="mr-2 h-4 w-4" />
                          取消部署
                        </Button>
                      )
                    ) : null}
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
          <LogDialog
            run={selectedRun}
            onClose={handleCloseDialog}
            onCancelRequest={(run) => handleCancelRequest(run)}
            onCancelConfirm={(run) => void handleCancelRun(run)}
            onCancelDismiss={handleCancelDismiss}
            cancelling={cancellingRunId === selectedRun.id}
            confirmingCancel={confirmingCancelRunId === selectedRun.id}
          />
        )}
      </AnimatePresence>
    </div>
  );
}
