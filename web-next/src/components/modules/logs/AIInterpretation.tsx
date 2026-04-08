"use client";

import { useState } from "react";
import { Bot, LoaderCircle, Sparkles } from "lucide-react";

import type { AIInterpretationResponse, RunStatus } from "@/types";
import { Button } from "@/components/ui/button";

type FailedRunAIInterpretationProps = {
  runId: number;
  runStatus: RunStatus;
  enabled: boolean;
  interpret: (runId: number) => Promise<AIInterpretationResponse>;
};

type InterpretationState =
  | { status: "idle" }
  | { status: "loading" }
  | { status: "success"; result: AIInterpretationResponse }
  | { status: "error"; message: string };

export function FailedRunAIInterpretation({
  runId,
  runStatus,
  enabled,
  interpret,
}: FailedRunAIInterpretationProps) {
  const [state, setState] = useState<InterpretationState>({ status: "idle" });

  if (!enabled || runStatus !== "failed") {
    return null;
  }

  const open = state.status !== "idle";

  const handleInterpret = async () => {
    setState({ status: "loading" });
    try {
      const result = await interpret(runId);
      setState({ status: "success", result });
    } catch (error: any) {
      setState({ status: "error", message: error?.message || "AI 解读失败" });
    }
  };

  return (
    <>
      <div className="pointer-events-none absolute bottom-6 right-6 z-10">
        <Button
          type="button"
          className="pointer-events-auto rounded-full px-5 shadow-lg"
          aria-label="AI 一键解读"
          onClick={() => void handleInterpret()}
        >
          <Sparkles className="mr-2 h-4 w-4" />
          AI 一键解读
        </Button>
      </div>

      {open ? (
        <aside className="flex h-64 shrink-0 flex-col border-t border-border bg-card/95 md:h-auto md:w-[360px] md:border-l md:border-t-0">
          <div className="flex items-center gap-2 border-b border-border px-4 py-3">
            <Bot className="h-4 w-4 text-primary" />
            <h3 className="text-sm font-semibold">AI 解读结果</h3>
          </div>

          <div className="flex-1 overflow-y-auto p-4 text-sm leading-6 text-card-foreground">
            {state.status === "loading" ? (
              <div className="flex h-full flex-col items-center justify-center gap-3 text-muted-foreground">
                <LoaderCircle className="h-5 w-5 animate-spin" />
                <p>AI 正在解读这次失败记录...</p>
              </div>
            ) : null}

            {state.status === "error" ? (
              <div className="rounded-2xl border border-destructive/20 bg-destructive/5 p-4 text-destructive">
                {state.message}
              </div>
            ) : null}

            {state.status === "success" ? (
              <div className="space-y-3">
                <div className="flex flex-wrap items-center gap-2 text-xs text-muted-foreground">
                  <span className="rounded-full bg-muted px-2.5 py-1">{state.result.protocol}</span>
                  <span className="rounded-full bg-muted px-2.5 py-1">{state.result.model}</span>
                </div>
                {state.result.log_truncated ? (
                  <p className="rounded-2xl border border-amber-200 bg-amber-50 px-3 py-2 text-xs text-amber-700">
                    已基于截断日志解读，当前只分析了最近一段失败日志。
                  </p>
                ) : null}
                <div className="whitespace-pre-wrap break-words">{state.result.content}</div>
              </div>
            ) : null}
          </div>
        </aside>
      ) : null}
    </>
  );
}
