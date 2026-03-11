"use client";

import { useEffect, useState, useRef } from "react";
import { Button } from "@/components/ui/button";
import { Card, CardContent } from "@/components/ui/card";
import { Badge } from "@/components/ui/badge";
import { runApi } from "@/api/client";
import type { PipelineRun } from "@/types";
import { formatDate, formatDuration, getStatusColor, getStatusText } from "@/lib/utils";
import { toast } from "sonner";
import { RefreshCw } from "lucide-react";

export function Logs() {
  const [runs, setRuns] = useState<PipelineRun[]>([]);
  const [selectedRun, setSelectedRun] = useState<PipelineRun | null>(null);
  const [logContent, setLogContent] = useState("");
  const [loading, setLoading] = useState(false);
  const logRef = useRef<HTMLPreElement>(null);
  const eventSourceRef = useRef<EventSource | null>(null);

  const loadRuns = async () => {
    try {
      const data = await runApi.list();
      const runs = data?.runs || [];
      setRuns(runs);
      if (runs.length > 0 && !selectedRun) {
        selectRun(runs[0]);
      }
    } catch (error) {
      console.error(error);
    }
  };

  useEffect(() => {
    loadRuns();
    return () => {
      if (eventSourceRef.current) {
        eventSourceRef.current.close();
      }
    };
  }, []);

  const selectRun = (run: PipelineRun) => {
    setSelectedRun(run);
    setLogContent(run.log || "");

    // 关闭之前的 EventSource
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
    }

    // 如果是运行中的任务，开启流式日志
    if (run.status === "running") {
      const es = new EventSource(`/api/v1/runs/${run.id}/log/stream`);
      es.onmessage = (event) => {
        setLogContent((prev) => prev + event.data);
        if (logRef.current) {
          logRef.current.scrollTop = logRef.current.scrollHeight;
        }
      };
      es.onerror = () => {
        es.close();
        loadRuns();
      };
      eventSourceRef.current = es;
    }
  };

  const handleRefresh = async () => {
    setLoading(true);
    await loadRuns();
    toast.success("记录已刷新");
    setLoading(false);
  };

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-end">
        <div className="flex items-center gap-4">
          <Badge variant="secondary">{runs.length} 条记录</Badge>
          <Button variant="outline" size="sm" onClick={handleRefresh} disabled={loading}>
            <RefreshCw className={`h-4 w-4 mr-2 ${loading ? "animate-spin" : ""}`} />
            刷新记录
          </Button>
        </div>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-[300px_1fr] gap-4">
        {/* 左侧列表 */}
        <div className="max-h-[600px] overflow-y-auto space-y-2">
          {runs.length === 0 ? (
            <Card>
              <CardContent className="p-6 text-center text-muted-foreground">
                暂无部署记录
              </CardContent>
            </Card>
          ) : (
            runs.map((run) => (
              <Card
                key={run.id}
                className={`cursor-pointer transition-colors ${
                  selectedRun?.id === run.id ? "border-primary" : ""
                }`}
                onClick={() => selectRun(run)}
              >
                <CardContent className="p-4">
                  <div className="flex items-center justify-between mb-2">
                    <span className="font-medium text-foreground">{run.project_name}</span>
                    <Badge className={getStatusColor(run.status)}>
                      {getStatusText(run.status)}
                    </Badge>
                  </div>
                  <p className="text-sm text-muted-foreground mb-1">{run.branch}</p>
                  <div className="flex items-center justify-between text-xs text-muted-foreground">
                    <span>{formatDate(run.started_at)}</span>
                    <span>{formatDuration(run.duration)}</span>
                  </div>
                </CardContent>
              </Card>
            ))
          )}
        </div>

        {/* 右侧日志 */}
        <Card>
          <CardContent className="p-0">
            <pre
              ref={logRef}
              className="h-[600px] overflow-auto p-4 bg-slate-900 text-green-200 font-mono text-sm whitespace-pre-wrap"
            >
              {logContent || "尚未加载日志。"}
            </pre>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
