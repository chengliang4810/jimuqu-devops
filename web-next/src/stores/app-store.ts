import { create } from "zustand";

interface Host {
  id: number;
  name: string;
  address: string;
  port: number;
  username: string;
}

interface Project {
  id: number;
  name: string;
  branch: string;
  repo_url: string;
  description?: string;
}

interface NotifyChannel {
  id: number;
  name: string;
  type: string;
  remark?: string;
}

interface PipelineRun {
  id: number;
  project_id: number;
  status: "queued" | "running" | "success" | "failed";
  started_at?: string;
  finished_at?: string;
}

interface AppState {
  // 认证状态
  isAuthenticated: boolean;
  setAuthenticated: (value: boolean) => void;

  // 数据状态
  hosts: Host[];
  projects: Project[];
  notifyChannels: NotifyChannel[];
  runs: PipelineRun[];

  // 设置数据
  setHosts: (hosts: Host[]) => void;
  setProjects: (projects: Project[]) => void;
  setNotifyChannels: (channels: NotifyChannel[]) => void;
  setRuns: (runs: PipelineRun[]) => void;

  // 增删改
  addHost: (host: Host) => void;
  updateHost: (host: Host) => void;
  removeHost: (id: number) => void;

  addProject: (project: Project) => void;
  updateProject: (project: Project) => void;
  removeProject: (id: number) => void;

  addNotifyChannel: (channel: NotifyChannel) => void;
  updateNotifyChannel: (channel: NotifyChannel) => void;
  removeNotifyChannel: (id: number) => void;

  addRun: (run: PipelineRun) => void;
  updateRun: (run: PipelineRun) => void;
}

export const useAppStore = create<AppState>((set) => ({
  // 初始状态
  isAuthenticated: false,
  hosts: [],
  projects: [],
  notifyChannels: [],
  runs: [],

  // 设置认证状态
  setAuthenticated: (value) => set({ isAuthenticated: value }),

  // 设置数据
  setHosts: (hosts) => set({ hosts }),
  setProjects: (projects) => set({ projects }),
  setNotifyChannels: (channels) => set({ notifyChannels: channels }),
  setRuns: (runs) => set({ runs }),

  // 主机操作
  addHost: (host) => set((state) => ({ hosts: [...state.hosts, host] })),
  updateHost: (host) =>
    set((state) => ({
      hosts: state.hosts.map((h) => (h.id === host.id ? host : h)),
    })),
  removeHost: (id) =>
    set((state) => ({ hosts: state.hosts.filter((h) => h.id !== id) })),

  // 项目操作
  addProject: (project) =>
    set((state) => ({ projects: [...state.projects, project] })),
  updateProject: (project) =>
    set((state) => ({
      projects: state.projects.map((p) => (p.id === project.id ? project : p)),
    })),
  removeProject: (id) =>
    set((state) => ({ projects: state.projects.filter((p) => p.id !== id) })),

  // 通知渠道操作
  addNotifyChannel: (channel) =>
    set((state) => ({ notifyChannels: [...state.notifyChannels, channel] })),
  updateNotifyChannel: (channel) =>
    set((state) => ({
      notifyChannels: state.notifyChannels.map((c) =>
        c.id === channel.id ? channel : c
      ),
    })),
  removeNotifyChannel: (id) =>
    set((state) => ({
      notifyChannels: state.notifyChannels.filter((c) => c.id !== id),
    })),

  // 部署记录操作
  addRun: (run) => set((state) => ({ runs: [run, ...state.runs] })),
  updateRun: (run) =>
    set((state) => ({
      runs: state.runs.map((r) => (r.id === run.id ? run : r)),
    })),
}));
