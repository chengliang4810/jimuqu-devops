"use client";

import { useEffect, useState } from "react";
import { Copy, KeyRound, Plus, RotateCcw, Trash2 } from "lucide-react";
import { toast } from "sonner";
import { settingApi } from "@/api/client";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog";
import { Input } from "@/components/ui/input";
import type { ApiToken } from "@/types";

function formatDate(value?: string | null) {
  if (!value) return "从不过期";
  return new Date(value).toLocaleString();
}

function defaultExpiresAt() {
  return "";
}

export function SettingApiTokens() {
  const [tokens, setTokens] = useState<ApiToken[]>([]);
  const [loading, setLoading] = useState(false);
  const [creating, setCreating] = useState(false);
  const [dialogOpen, setDialogOpen] = useState(false);
  const [name, setName] = useState("");
  const [expiresAt, setExpiresAt] = useState(defaultExpiresAt);
  const [createdToken, setCreatedToken] = useState("");

  const loadTokens = async () => {
    try {
      setLoading(true);
      setTokens(await settingApi.listApiTokens());
    } catch (error: any) {
      toast.error(error.message || "加载 API Token 失败");
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    void loadTokens();
  }, []);

  const handleCreate = async () => {
    if (!name.trim()) {
      toast.error("Token 名称不能为空");
      return;
    }
    try {
      setCreating(true);
      const response = await settingApi.createApiToken({
        name: name.trim(),
        expires_at: expiresAt ? new Date(expiresAt).toISOString() : null,
      });
      setCreatedToken(response.token);
      setTokens((current) => [response.api_token, ...current]);
      setName("");
      setExpiresAt(defaultExpiresAt());
      toast.success("API Token 已创建");
    } catch (error: any) {
      toast.error(error.message || "创建 API Token 失败");
    } finally {
      setCreating(false);
    }
  };

  const handleCopy = async (value: string) => {
    await navigator.clipboard.writeText(value);
    toast.success("已复制");
  };

  const handleRevoke = async (token: ApiToken) => {
    if (!window.confirm(`确认撤销 API Token「${token.name}」？`)) {
      return;
    }
    try {
      await settingApi.revokeApiToken(token.id);
      await loadTokens();
      toast.success("API Token 已撤销");
    } catch (error: any) {
      toast.error(error.message || "撤销 API Token 失败");
    }
  };

  return (
    <div className="rounded-3xl border border-border bg-card p-6 space-y-5">
      <div className="flex items-center justify-between gap-3">
        <h2 className="flex items-center gap-2 text-lg font-bold text-card-foreground">
          <KeyRound className="h-5 w-5" />
          API Token
        </h2>
        <div className="flex gap-2">
          <Button type="button" variant="outline" size="sm" onClick={() => void loadTokens()} disabled={loading}>
            <RotateCcw className="h-4 w-4" />
          </Button>
          <Button type="button" size="sm" onClick={() => setDialogOpen(true)}>
            <Plus className="h-4 w-4" />
            新建
          </Button>
        </div>
      </div>

      <div className="space-y-3">
        {tokens.length === 0 ? (
          <div className="rounded-2xl border border-dashed border-border p-4 text-sm text-muted-foreground">
            暂无 API Token。
          </div>
        ) : (
          tokens.map((token) => (
            <div key={token.id} className="rounded-2xl border border-border p-4 space-y-2">
              <div className="flex items-center justify-between gap-3">
                <div>
                  <div className="font-medium text-card-foreground">{token.name}</div>
                  <div className="text-xs text-muted-foreground">{token.prefix}...</div>
                </div>
                {token.revoked_at ? <Badge variant="secondary">已撤销</Badge> : <Badge variant="success">可用</Badge>}
              </div>
              <div className="grid gap-1 text-xs text-muted-foreground sm:grid-cols-2">
                <span>过期时间：{formatDate(token.expires_at)}</span>
                <span>最后使用：{token.last_used_at ? formatDate(token.last_used_at) : "尚未使用"}</span>
              </div>
              {!token.revoked_at && (
                <Button type="button" variant="outline" size="sm" onClick={() => void handleRevoke(token)}>
                  <Trash2 className="h-4 w-4" />
                  撤销
                </Button>
              )}
            </div>
          ))
        )}
      </div>

      <Dialog open={dialogOpen} onOpenChange={(open) => {
        setDialogOpen(open);
        if (!open) {
          setCreatedToken("");
        }
      }}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>新建 API Token</DialogTitle>
            <DialogDescription>
              Token 只会显示一次，请复制后保存到 Agent 或 CI 的环境变量中。
            </DialogDescription>
          </DialogHeader>

          {createdToken ? (
            <div className="space-y-3">
              <div className="rounded-xl border border-border bg-muted p-3 break-all font-mono text-sm">
                {createdToken}
              </div>
              <Button type="button" className="w-full" onClick={() => void handleCopy(createdToken)}>
                <Copy className="h-4 w-4" />
                复制 Token
              </Button>
            </div>
          ) : (
            <div className="space-y-3">
              <Input value={name} onChange={(event) => setName(event.target.value)} placeholder="Token 名称，例如 codex-agent" />
              <Input
                type="datetime-local"
                value={expiresAt}
                onChange={(event) => setExpiresAt(event.target.value)}
                placeholder="过期时间"
              />
              <div className="text-xs text-muted-foreground">不填写过期时间表示默认不过期。</div>
            </div>
          )}

          <DialogFooter>
            {createdToken ? (
              <Button type="button" onClick={() => setDialogOpen(false)}>完成</Button>
            ) : (
              <Button type="button" onClick={() => void handleCreate()} disabled={creating || !name.trim()}>
                {creating ? "创建中..." : "创建"}
              </Button>
            )}
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  );
}
