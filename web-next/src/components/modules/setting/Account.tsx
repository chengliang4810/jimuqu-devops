"use client";

import { useEffect, useState } from "react";
import { Eye, EyeOff, KeyRound, Lock, User } from "lucide-react";
import { toast } from "sonner";
import { settingApi, clearToken } from "@/api/client";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";

function forceLogout() {
  clearToken();
  window.location.href = "/";
}

export function SettingAccount() {
  const [username, setUsername] = useState("");
  const [newUsername, setNewUsername] = useState("");
  const [oldPassword, setOldPassword] = useState("");
  const [newPassword, setNewPassword] = useState("");
  const [confirmPassword, setConfirmPassword] = useState("");
  const [savingUsername, setSavingUsername] = useState(false);
  const [savingPassword, setSavingPassword] = useState(false);
  const [showOldPassword, setShowOldPassword] = useState(false);
  const [showNewPassword, setShowNewPassword] = useState(false);
  const [showConfirmPassword, setShowConfirmPassword] = useState(false);

  useEffect(() => {
    void (async () => {
      try {
        const profile = await settingApi.getProfile();
        setUsername(profile.username);
      } catch (error) {
        console.error(error);
      }
    })();
  }, []);

  const handleChangeUsername = async () => {
    if (!newUsername.trim()) {
      toast.error("新用户名不能为空");
      return;
    }

    try {
      setSavingUsername(true);
      await settingApi.changeUsername(newUsername.trim());
      toast.success("用户名修改成功，请重新登录");
      setTimeout(forceLogout, 800);
    } catch (error: any) {
      toast.error(error.message || "用户名修改失败");
    } finally {
      setSavingUsername(false);
    }
  };

  const handleChangePassword = async () => {
    if (!oldPassword || !newPassword || !confirmPassword) {
      toast.error("请完整填写密码信息");
      return;
    }
    if (newPassword !== confirmPassword) {
      toast.error("两次输入的新密码不一致");
      return;
    }

    try {
      setSavingPassword(true);
      await settingApi.changePassword(oldPassword, newPassword);
      toast.success("密码修改成功，请重新登录");
      setTimeout(forceLogout, 800);
    } catch (error: any) {
      toast.error(error.message || "密码修改失败");
    } finally {
      setSavingPassword(false);
    }
  };

  return (
    <div className="rounded-3xl border border-border bg-card p-6 space-y-6">
      <h2 className="flex items-center gap-2 text-lg font-bold text-card-foreground">
        <User className="h-5 w-5" />
        账户设置
      </h2>

      <div className="space-y-3">
        <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
          <KeyRound className="h-4 w-4" />
          当前用户名：{username || "加载中..."}
        </div>
        <div className="flex gap-2">
          <Input
            value={newUsername}
            onChange={(event) => setNewUsername(event.target.value)}
            placeholder="输入新的用户名"
            className="flex-1 rounded-xl"
          />
          <Button
            type="button"
            onClick={() => void handleChangeUsername()}
            disabled={savingUsername || !newUsername.trim()}
            className="rounded-xl"
          >
            {savingUsername ? "保存中..." : "保存"}
          </Button>
        </div>
      </div>

      <div className="border-t border-border" />

      <div className="space-y-3">
        <div className="flex items-center gap-2 text-sm font-medium text-muted-foreground">
          <Lock className="h-4 w-4" />
          修改密码
        </div>

        <div className="relative">
          <Input
            type={showOldPassword ? "text" : "password"}
            value={oldPassword}
            onChange={(event) => setOldPassword(event.target.value)}
            placeholder="当前密码"
            className="rounded-xl pr-10"
          />
          <button
            type="button"
            onClick={() => setShowOldPassword((current) => !current)}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground"
          >
            {showOldPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
          </button>
        </div>

        <div className="relative">
          <Input
            type={showNewPassword ? "text" : "password"}
            value={newPassword}
            onChange={(event) => setNewPassword(event.target.value)}
            placeholder="新密码"
            className="rounded-xl pr-10"
          />
          <button
            type="button"
            onClick={() => setShowNewPassword((current) => !current)}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground"
          >
            {showNewPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
          </button>
        </div>

        <div className="relative">
          <Input
            type={showConfirmPassword ? "text" : "password"}
            value={confirmPassword}
            onChange={(event) => setConfirmPassword(event.target.value)}
            placeholder="确认新密码"
            className="rounded-xl pr-10"
          />
          <button
            type="button"
            onClick={() => setShowConfirmPassword((current) => !current)}
            className="absolute right-3 top-1/2 -translate-y-1/2 text-muted-foreground"
          >
            {showConfirmPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
          </button>
        </div>

        <Button
          type="button"
          onClick={() => void handleChangePassword()}
          disabled={savingPassword || !oldPassword || !newPassword || !confirmPassword}
          className="w-full rounded-xl"
        >
          {savingPassword ? "保存中..." : "修改密码"}
        </Button>
      </div>
    </div>
  );
}
