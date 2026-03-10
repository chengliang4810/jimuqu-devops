const state = {
  hosts: [],
  projects: [],
  runs: [],
  notifyChannels: [],
  activeView: "home",
  selectedRunId: null,
  runDetail: null,
  token: localStorage.getItem("jwt_token") || null,
  isAuthenticated: !!localStorage.getItem("jwt_token"),
};

// 移除轮询计时器，改用手动刷新
// let pollTimer = null;
let runStream = null;
let streamingRunId = null;
const els = {};

document.addEventListener("DOMContentLoaded", () => {
  bindElements();
  bindEvents();
  els.serverOrigin.textContent = window.location.origin;

  // 检查认证状态
  if (state.isAuthenticated) {
    showMainApp();
  } else {
    showLoginScreen();
  }

  // 移除自动轮询，用户可以通过刷新按钮手动刷新
  // startPolling();
});

window.addEventListener("beforeunload", () => {
  stopRunStream();
});

window.addEventListener("hashchange", () => {
  switchView(getInitialView(), false);
});

window.addEventListener("keydown", (event) => {
  if (event.key === "Escape") {
    closeHostModal();
    closeProjectModal();
    // closeProjectConfigModal(); // 移除，已合并到项目模态框
  }
});

function bindElements() {
  Object.assign(els, {
    // 登录相关
    loginScreen: document.getElementById("loginScreen"),
    loginForm: document.getElementById("loginForm"),
    loginUsername: document.getElementById("loginUsername"),
    loginPassword: document.getElementById("loginPassword"),
    loginMessage: document.getElementById("loginMessage"),
    logoutBtn: document.getElementById("logoutBtn"),
    pageShell: document.querySelector(".page-shell"),

    // 主应用
    navButtons: Array.from(document.querySelectorAll(".nav-button")),
    viewSections: Array.from(document.querySelectorAll(".view-section")),
    messageBar: document.getElementById("messageBar"),
    serverOrigin: document.getElementById("serverOrigin"),
    homeProjectChip: document.getElementById("homeProjectChip"),
    homeSelectedProjectMeta: document.getElementById("homeSelectedProjectMeta"),
    homeGoProjectsBtn: document.getElementById("homeGoProjectsBtn"),
    homeGoLogsBtn: document.getElementById("homeGoLogsBtn"),
    projectCount: document.getElementById("projectCount"),
    hostCount: document.getElementById("hostCount"),
    runCount: document.getElementById("runCount"),
    notifyCount: document.getElementById("notifyCount"),
    selectedRunStatus: document.getElementById("selectedRunStatus"),
    lastSyncText: document.getElementById("lastSyncText"),
    hostSummaryChip: document.getElementById("hostSummaryChip"),
    projectSummaryChip: document.getElementById("projectSummaryChip"),
    notifyChannelSummaryChip: document.getElementById("notifyChannelSummaryChip"),
    runSummaryChip: document.getElementById("runSummaryChip"),
    hostList: document.getElementById("hostList"),
    projectList: document.getElementById("projectList"),
    runList: document.getElementById("runList"),
    runDetailTitle: document.getElementById("runDetailTitle"),
    runDetailMeta: document.getElementById("runDetailMeta"),
    runLogOutput: document.getElementById("runLogOutput"),
    runStreamState: document.getElementById("runStreamState"),
    refreshLogsBtn: document.getElementById("refreshLogsBtn"),
    hostForm: document.getElementById("hostForm"),
    projectForm: document.getElementById("projectForm"),
    buildForm: document.getElementById("buildForm"),
    deployForm: document.getElementById("deployForm"),
    hostFormId: document.getElementById("hostFormId"),
    deployHostId: document.getElementById("hostId"),
    hostName: document.getElementById("hostName"),
    hostAddress: document.getElementById("hostAddress"),
    hostPort: document.getElementById("hostPort"),
    hostUsername: document.getElementById("hostUsername"),
    hostPassword: document.getElementById("hostPassword"),
    projectId: document.getElementById("projectId"),
    projectName: document.getElementById("projectName"),
    projectBranch: document.getElementById("projectBranch"),
    projectRepoURL: document.getElementById("projectRepoURL"),
    projectDescription: document.getElementById("projectDescription"),
    timeoutMinutes: document.getElementById("timeoutMinutes"),
    webhookToken: document.getElementById("webhookToken"),
    buildImage: document.getElementById("buildImage"),
    buildCommands: document.getElementById("buildCommands"),
    artifactFilterMode: document.getElementById("artifactFilterMode"),
    artifactRules: document.getElementById("artifactRules"),
    remoteSaveDir: document.getElementById("remoteSaveDir"),
    remoteDeployDir: document.getElementById("remoteDeployDir"),
    preDeployCommands: document.getElementById("preDeployCommands"),
    postDeployCommands: document.getElementById("postDeployCommands"),
    versionCount: document.getElementById("versionCount"),
    notifyWebhookURL: document.getElementById("notifyWebhookURL"),
    notifyBearerToken: document.getElementById("notifyBearerToken"),
    hostModal: document.getElementById("hostModal"),
    hostModalTitle: document.getElementById("hostModalTitle"),
    closeHostModalBtn: document.getElementById("closeHostModalBtn"),
    cancelHostModalBtn: document.getElementById("cancelHostModalBtn"),
    addHostBtn: document.getElementById("addHostBtn"),
    projectModal: document.getElementById("projectModal"),
    projectModalTitle: document.getElementById("projectModalTitle"),
    closeProjectModalBtn: document.getElementById("closeProjectModalBtn"),
    cancelProjectModalBtn: document.getElementById("cancelProjectModalBtn"),
    addProjectBtn: document.getElementById("addProjectBtn"),
    saveProjectBtn: document.getElementById("saveProjectBtn"),
    projectConfigModal: document.getElementById("projectConfigModal"),
    projectConfigModalTitle: document.getElementById("projectConfigModalTitle"),
    closeProjectConfigModalBtn: document.getElementById("closeProjectConfigModalBtn"),
    cancelProjectConfigModalBtn: document.getElementById("cancelProjectConfigModalBtn"),
    notifyChannelList: document.getElementById("notifyChannelList"),
    addNotifyChannelBtn: document.getElementById("addNotifyChannelBtn"),
    notifyChannelModal: document.getElementById("notifyChannelModal"),
    notifyChannelModalTitle: document.getElementById("notifyChannelModalTitle"),
    closeNotifyChannelModalBtn: document.getElementById("closeNotifyChannelModalBtn"),
    cancelNotifyChannelModalBtn: document.getElementById("cancelNotifyChannelModalBtn"),
  });
}

function bindEvents() {
  // 登录事件
  if (els.loginForm) {
    els.loginForm.addEventListener("submit", handleLogin);
  }

  // 退出登录事件
  if (els.logoutBtn) {
    els.logoutBtn.addEventListener("click", logout);
  }

  els.hostForm.addEventListener("submit", handleHostSubmit);
  els.projectForm.addEventListener("submit", handleProjectSubmit);
  // 移除刷新和首页按钮事件
  // els.refreshAllBtn.addEventListener("click", refreshAll);
  // els.homeGoProjectsBtn.addEventListener("click", () => switchView("projects"));
  // els.homeGoLogsBtn.addEventListener("click", () => switchView("logs"));
  els.navButtons.forEach((button) => {
    button.addEventListener("click", () => switchView(button.dataset.viewTarget));
  });
  // 主机模态框事件
  els.addHostBtn.addEventListener("click", openHostModalForCreate);
  els.closeHostModalBtn.addEventListener("click", closeHostModal);
  els.cancelHostModalBtn.addEventListener("click", closeHostModal);
  // 项目模态框事件
  els.addProjectBtn.addEventListener("click", openProjectModalForCreate);
  els.closeProjectModalBtn.addEventListener("click", closeProjectModal);
  els.cancelProjectModalBtn.addEventListener("click", closeProjectModal);
  els.saveProjectBtn.addEventListener("click", handleProjectSubmit);

  // 通知渠道事件
  els.addNotifyChannelBtn.addEventListener("click", openNotifyChannelModal);
  els.notifyChannelForm.addEventListener("submit", handleNotifyChannelSubmit);
  els.closeNotifyChannelModalBtn.addEventListener("click", closeNotifyChannelModal);
  els.cancelNotifyChannelModalBtn.addEventListener("click", closeNotifyChannelModal);

  // 部署记录事件
  els.refreshLogsBtn.addEventListener("click", refreshRuns);
  els.loadMoreLogsBtn.addEventListener("click", loadMoreRuns);

  // Tab切换事件
  document.querySelectorAll('.tab-button').forEach(button => {
    button.addEventListener('click', (event) => {
      switchTab(event.target.dataset.tab);
    });
  });

  // 移除项目配置相关事件
  // els.projectConfigForm.addEventListener("submit", handleProjectConfigSubmit);
  // els.closeProjectConfigModalBtn.addEventListener("click", closeProjectConfigModal);
  // els.cancelProjectConfigModalBtn.addEventListener("click", closeProjectConfigModal);
}

function switchTab(tabName) {
  // 更新tab按钮状态
  document.querySelectorAll('.tab-button').forEach(button => {
    button.classList.remove('active');
    if (button.dataset.tab === tabName) {
      button.classList.add('active');
    }
  });

  // 更新tab内容显示
  document.querySelectorAll('.tab-pane').forEach(pane => {
    pane.classList.remove('active');
  });
  document.getElementById(`tab-${tabName}`).classList.add('active');
}

function getInitialView() {
  const raw = window.location.hash.replace("#", "").trim();
  return ["home", "hosts", "projects", "notifications", "logs"].includes(raw) ? raw : "home";
}

function switchView(view, updateHash = true) {
  const normalized = ["home", "hosts", "projects", "notifications", "logs"].includes(view) ? view : "home";
  state.activeView = normalized;

  els.navButtons.forEach((button) => {
    button.classList.toggle("active", button.dataset.viewTarget === normalized);
  });
  els.viewSections.forEach((section) => {
    section.classList.toggle("active", section.id === `view-${normalized}`);
  });

  if (updateHash) {
    window.location.hash = normalized;
  }
}

async function refreshAll() {
  try {
    const [hosts, projects, runs, notifyChannels] = await Promise.all([
      api("/api/v1/hosts"),
      api("/api/v1/projects"),
      api("/api/v1/runs"),
      api("/api/v1/notification-channels").catch(() => []),
    ]);
    state.hosts = hosts;
    state.projects = projects;
    state.runs = runs;
    state.notifyChannels = notifyChannels || [];
    renderHosts();
    renderProjects();
    renderNotifyChannels();
    renderHostOptionsForConfig();
    renderOverview();
    renderRuns();
  } catch (error) {
    showMessage(error.message, "error");
  }
}

// 移除自动轮询功能，用户可以通过刷新按钮手动刷新
// function startPolling() {
//   if (pollTimer) {
//     window.clearInterval(pollTimer);
//   }
//   pollTimer = window.setInterval(() => {
//     refreshAll();
//   }, 5000);
// }

function renderOverview() {
  els.projectCount.textContent = String(state.projects?.length || 0);
  els.hostCount.textContent = String(state.hosts?.length || 0);
  els.runCount.textContent = String(state.runs?.length || 0);

  // 计算有配置通知的项目数量
  const notifyProjects = (state.projects || []).filter((project) => {
    return project.notify_webhook_url && project.notify_webhook_url.trim() !== "";
  });
  els.notifyCount.textContent = String(notifyProjects.length);

  els.hostSummaryChip.textContent = `${state.hosts?.length || 0} 台主机`;
  els.projectSummaryChip.textContent = `${state.projects?.length || 0} 个项目`;

  // 移除首页相关逻辑
  // const runs = state.runs || [];
  // const latestRun = runs[0];
  // if (latestRun) {
  //   els.selectedRunStatus.textContent = statusText(latestRun.status);
  //   els.lastSyncText.textContent = `最近部署 ${formatDateTime(latestRun.created_at)}`;
  // } else {
  //   els.selectedRunStatus.textContent = "暂无部署";
  //   els.lastSyncText.textContent = `最近同步 ${formatDateTime(new Date().toISOString())}`;
  // }
  // renderHomeSpotlight();
}

// 移除首页相关函数
// function renderHomeSpotlight() {
//   els.homeGoLogsBtn.disabled = false;
//   els.homeProjectChip.textContent = "查看部署记录";
//   els.homeProjectChip.className = "chip chip-teal";
//   els.homeSelectedProjectMeta.innerHTML = `
//     <div class="meta-card"><span>总项目数</span><strong>${(state.projects || []).length} 个</strong></div>
//     <div class="meta-card"><span>已配置</span><strong>${(state.projects || []).filter((p) => p.has_deploy_config).length} 个</strong></div>
//     <div class="meta-card"><span>总主机数</span><strong>${(state.hosts || []).length} 台</strong></div>
//   `;
// }
// 移除首页相关函数结束

function renderHosts() {
  if (!state.hosts.length) {
    els.hostList.className = "list-grid empty-state";
    els.hostList.textContent = "暂无主机，先创建一个部署目标。";
    return;
  }

  els.hostList.className = "list-grid";
  els.hostList.innerHTML = "";

  state.hosts.forEach((host) => {
    const card = document.createElement("article");
    card.className = "list-card";

    const head = document.createElement("div");
    head.className = "item-head";

    const info = document.createElement("div");
    info.innerHTML = `
      <p class="item-title">${escapeHTML(host.name)}</p>
      <p class="item-subtitle">${escapeHTML(host.username)}@${escapeHTML(host.address)}:${host.port}</p>
      <p class="item-meta">${host.has_password ? "已保存 SSH 密码" : "未设置 SSH 密码"}</p>
    `;

    const badge = document.createElement("span");
    badge.className = "chip chip-teal";
    badge.textContent = "SSH";

    const actions = document.createElement("div");
    actions.className = "item-actions";
    actions.append(
      createMicroButton("编辑", (event) => {
        event.stopPropagation();
        openHostModalForEdit(host);
      }),
      createMicroButton("删除", (event) => {
        event.stopPropagation();
        handleDeleteHost(host.id, host.name);
      }, true),
    );

    head.append(info, badge);
    card.append(head, actions);
    els.hostList.append(card);
  });
}

function renderProjects() {
  if (!state.projects || !state.projects.length) {
    els.projectList.className = "list-grid empty-state";
    els.projectList.textContent = "暂无项目，点击上方按钮新增项目。";
    return;
  }

  els.projectList.className = "list-grid";
  els.projectList.innerHTML = "";

  state.projects.forEach((project) => {
    const card = document.createElement("article");
    card.className = "list-card";

    const head = document.createElement("div");
    head.className = "item-head";

    const info = document.createElement("div");
    info.innerHTML = `
      <p class="item-title">${escapeHTML(project.name)}</p>
      <p class="item-subtitle">${escapeHTML(project.branch)}</p>
      <p class="item-meta">${escapeHTML(project.repo_url)}</p>
    `;

    const actions = document.createElement("div");
    actions.className = "item-actions";

    const buttons = [
      createMicroButton("Webhook", (event) => {
        event.stopPropagation();
        const webhookUrl = `${window.location.origin}/api/v1/webhooks/${project.webhook_token}`;
        navigator.clipboard.writeText(webhookUrl).then(() => {
          showMessage("Webhook URL已复制到剪贴板", "success");
        }).catch(() => {
          showMessage("复制失败，请手动复制", "error");
        });
      }),
      createMicroButton("配置", (event) => {
        event.stopPropagation();
        openProjectModalForEdit(project);
      }),
    ];

    if (project.has_deploy_config) {
      buttons.push(
        createMicroButton("触发", (event) => {
          event.stopPropagation();
          handleTriggerProject(project.id, project.name);
        })
      );
    }

    buttons.push(
      createMicroButton("删除", (event) => {
        event.stopPropagation();
        handleDeleteProject(project.id, project.name);
      }, true)
    );

    actions.append(...buttons);

    head.append(info);
    card.append(head, actions);
    els.projectList.append(card);
  });
}

function renderNotifyChannels() {
  if (!state.notifyChannels || !state.notifyChannels.length) {
    els.notifyChannelSummaryChip.textContent = "0 个渠道";
    els.notifyChannelList.className = "list-grid empty-state";
    els.notifyChannelList.textContent = "暂无通知渠道，点击上方按钮新增渠道。";
    return;
  }

  els.notifyChannelSummaryChip.textContent = `${state.notifyChannels.length} 个渠道`;
  els.notifyChannelList.className = "list-grid";
  els.notifyChannelList.innerHTML = "";

  state.notifyChannels.forEach((channel) => {
    const card = document.createElement("article");
    card.className = "list-card";

    const head = document.createElement("div");
    head.className = "item-head";

    const info = document.createElement("div");
    info.innerHTML = `
      <p class="item-title">${escapeHTML(channel.name)}</p>
      <p class="item-subtitle">${escapeHTML(channel.type)}</p>
      <p class="item-meta">${channel.is_default ? "默认渠道" : "普通渠道"} ${channel.remark ? `- ${escapeHTML(channel.remark)}` : ""}</p>
    `;

    const actions = document.createElement("div");
    actions.className = "item-actions";

    const buttons = [
      createMicroButton("编辑", (event) => {
        event.stopPropagation();
        openNotifyChannelModalForEdit(channel);
      }),
      createMicroButton("删除", (event) => {
        event.stopPropagation();
        handleDeleteNotifyChannel(channel.id, channel.name);
      }, true)
    ];

    actions.append(...buttons);
    head.append(info);
    card.append(head, actions);
    els.notifyChannelList.append(card);
  });
}

function renderRuns() {
  const runs = state.runs || [];
  if (!runs.length) {
    els.runSummaryChip.textContent = "0 条记录";
    els.runList.className = "run-list empty-state";
    els.runList.textContent = "暂无部署记录。";
    renderRunDetail();
    return;
  }

  els.runSummaryChip.textContent = `${runs.length} 条记录`;
  els.runList.className = "run-list";
  els.runList.innerHTML = "";

  runs.forEach((run) => {
    const project = (state.projects || []).find(p => p.id === run.project_id) || { name: "未知项目", branch: "-" };

    const card = document.createElement("article");
    card.className = `run-card${state.selectedRunId === run.id ? " selected" : ""}`;
    card.addEventListener("click", () => selectRun(run.id));

    const head = document.createElement("div");
    head.className = "run-card-head";

    const info = document.createElement("div");
    info.innerHTML = `
      <p class="item-title">#${run.id} ${escapeHTML(project.name)}</p>
      <p class="item-subtitle">${escapeHTML(project.branch)} / ${escapeHTML(run.trigger_type)} / ${escapeHTML(run.trigger_ref || "-")}</p>
      <p class="item-meta">${formatDateTime(run.created_at)}</p>
    `;

    const badge = document.createElement("span");
    badge.className = `status-chip ${statusClass(run.status)}`;
    badge.textContent = statusText(run.status);

    head.append(info, badge);
    card.append(head);
    els.runList.append(card);
  });

  renderRunDetail();
}

function renderRunDetail() {
  if (!state.runDetail) {
    els.runDetailTitle.textContent = "暂无运行详情";
    els.runDetailMeta.textContent = "点击左侧记录查看日志";
    els.runLogOutput.textContent = "尚未加载日志。";
    setRunStreamState("未连接", "queued");
    return;
  }

  const run = state.runDetail;
  els.runDetailTitle.textContent = `运行 #${run.id} / ${statusText(run.status)}`;
  els.runDetailMeta.textContent = `${formatDateTime(run.created_at)} / ${run.trigger_type} / ${run.trigger_ref || "-"}`;

  const sections = [];
  if (run.error_message) {
    sections.push(`[ERROR]\n${run.error_message}`);
  }
  sections.push(run.log_text || "暂无日志输出。");
  els.runLogOutput.textContent = sections.join("\n\n");

  // 自动滚动到底部显示最新日志
  els.runLogOutput.scrollTop = els.runLogOutput.scrollHeight;

  if (isTerminalStatus(run.status)) {
    setRunStreamState("已结束", run.status);
  } else if (streamingRunId === run.id) {
    setRunStreamState("实时刷新中", "running");
  }
}

function startRunStream(runId) {
  if (!runId) {
    stopRunStream();
    return;
  }
  if (streamingRunId === runId && runStream) {
    return;
  }

  stopRunStream(false);
  streamingRunId = runId;
  setRunStreamState("连接中", "running");

  runStream = new EventSource(`/api/v1/runs/${runId}/stream`);
  runStream.addEventListener("run", (event) => {
    const run = JSON.parse(event.data);
    state.runDetail = run;
    syncRunIntoList(run);
    renderRuns();
    renderOverview();

    if (isTerminalStatus(run.status)) {
      setRunStreamState("已结束", run.status);
      stopRunStream(false);
      return;
    }

    setRunStreamState("实时刷新中", run.status);
  });

  runStream.onerror = () => {
    if (state.runDetail && state.runDetail.id === runId && isTerminalStatus(state.runDetail.status)) {
      setRunStreamState("已结束", state.runDetail.status);
      stopRunStream(false);
      return;
    }
    setRunStreamState("连接中断", "failed");
  };
}

function stopRunStream(resetState = true) {
  if (runStream) {
    runStream.close();
  }
  runStream = null;
  streamingRunId = null;
  if (resetState) {
    setRunStreamState("未连接", "queued");
  }
}

function syncRunIntoList(run) {
  const runs = state.runs || [];
  const index = runs.findIndex((item) => item.id === run.id);
  if (index >= 0) {
    runs[index] = { ...runs[index], ...run };
    return;
  }
  runs.unshift(run);
}

function setRunStreamState(text, status) {
  els.runStreamState.textContent = text;
  els.runStreamState.className = `status-chip ${statusClass(status || "queued")}`;
}

async function selectRun(runId) {
  state.selectedRunId = runId;
  try {
    state.runDetail = await api(`/api/v1/runs/${runId}`);
    startRunStream(runId);
    renderRuns();
  } catch (error) {
    showMessage(error.message, "error");
  }
}

async function handleHostSubmit(event) {
  event.preventDefault();
  const id = els.hostFormId.value.trim();
  const payload = {
    name: els.hostName.value.trim(),
    address: els.hostAddress.value.trim(),
    port: Number(els.hostPort.value || 22),
    username: els.hostUsername.value.trim(),
  };
  if (els.hostPassword.value) {
    payload.password = els.hostPassword.value;
  }

  try {
    if (id) {
      await api(`/api/v1/hosts/${id}`, { method: "PUT", body: payload });
      showMessage("主机已更新", "success");
    } else {
      if (!payload.password) {
        throw new Error("创建主机时必须填写密码");
      }
      await api("/api/v1/hosts", { method: "POST", body: payload });
      showMessage("主机已创建", "success");
    }
    closeHostModal();
    await refreshAll();
  } catch (error) {
    showMessage(error.message, "error");
  }
}

async function handleProjectSubmit(event) {
  event.preventDefault();
  const id = els.projectId.value.trim();

  // 基础信息
  const projectPayload = {
    name: els.projectName.value.trim(),
    repo_url: els.projectRepoURL.value.trim(),
    branch: els.projectBranch.value.trim(),
    description: els.projectDescription.value.trim(),
    webhook_token: els.webhookToken.value.trim() || undefined,
  };

  try {
    // 先创建或更新项目基础信息
    let project;
    if (id) {
      project = await api(`/api/v1/projects/${id}`, { method: "PUT", body: projectPayload });
    } else {
      project = await api("/api/v1/projects", { method: "POST", body: projectPayload });
    }

    // 然后保存部署配置（合并编译和部署Tab的数据）
    const configPayload = {
      host_id: parseInt(els.deployHostId.value) || 0,
      build_image: els.buildImage.value.trim(),
      build_commands: textToLines(els.buildCommands.value),
      artifact_filter_mode: els.artifactFilterMode.value,
      artifact_rules: textToLines(els.artifactRules.value),
      remote_save_dir: els.remoteSaveDir.value.trim(),
      remote_deploy_dir: els.remoteDeployDir.value.trim(),
      pre_deploy_commands: textToLines(els.preDeployCommands.value),
      post_deploy_commands: textToLines(els.postDeployCommands.value),
      timeout_seconds: parseInt(els.timeoutMinutes.value) * 60, // 转换为秒
      notify_webhook_url: els.notifyWebhookURL.value.trim(),
    };

    // 只有在有BearerToken时才添加
    if (els.notifyBearerToken.value.trim()) {
      configPayload.notify_bearer_token = els.notifyBearerToken.value.trim();
    }

    await api(`/api/v1/projects/${project.id}/config`, { method: "PUT", body: configPayload });

    showMessage(id ? "项目已更新" : "项目已创建", "success");
    closeProjectModal();
    await refreshAll();
  } catch (error) {
    showMessage(error.message, "error");
  }
}

async function handleDeleteHost(hostId, hostName) {
  if (!window.confirm(`确认删除主机 ${hostName} 吗？`)) {
    return;
  }
  try {
    await api(`/api/v1/hosts/${hostId}`, { method: "DELETE" });
    showMessage("主机已删除", "success");
    await refreshAll();
  } catch (error) {
    showMessage(error.message, "error");
  }
}

async function handleDeleteProject(projectId, projectName) {
  if (!window.confirm(`确认删除项目 ${projectName} 吗？`)) {
    return;
  }
  try {
    await api(`/api/v1/projects/${projectId}`, { method: "DELETE" });
    showMessage("项目已删除", "success");
    await refreshAll();
  } catch (error) {
    showMessage(error.message, "error");
  }
}

// 通知渠道相关函数
function openNotifyChannelModal() {
  els.notifyChannelModalTitle.textContent = "新增通知渠道";
  document.getElementById("notifyChannelId").value = "";
  document.getElementById("notifyChannelName").value = "";
  document.getElementById("notifyChannelType").value = "webhook";
  document.getElementById("notifyChannelConfig").value = "";
  document.getElementById("notifyChannelRemark").value = "";
  els.notifyChannelModal.style.display = "flex";
}

function openNotifyChannelModalForEdit(channel) {
  els.notifyChannelModalTitle.textContent = "编辑通知渠道";
  document.getElementById("notifyChannelId").value = channel.id;
  document.getElementById("notifyChannelName").value = channel.name;
  document.getElementById("notifyChannelType").value = channel.type;
  document.getElementById("notifyChannelConfig").value = channel.config_json || "";
  document.getElementById("notifyChannelRemark").value = channel.remark || "";
  els.notifyChannelModal.style.display = "flex";
}

function closeNotifyChannelModal() {
  els.notifyChannelModal.style.display = "none";
  els.notifyChannelForm.reset();
  document.getElementById("notifyChannelId").value = "";
}

async function handleNotifyChannelSubmit(event) {
  event.preventDefault();
  const id = document.getElementById("notifyChannelId").value;
  const channelPayload = {
    name: document.getElementById("notifyChannelName").value,
    type: document.getElementById("notifyChannelType").value,
    config_json: document.getElementById("notifyChannelConfig").value,
    remark: document.getElementById("notifyChannelRemark").value,
  };

  try {
    if (id) {
      await api(`/api/v1/notification-channels/${id}`, { method: "PUT", body: channelPayload });
      showMessage("通知渠道已更新", "success");
    } else {
      await api("/api/v1/notification-channels", { method: "POST", body: channelPayload });
      showMessage("通知渠道已创建", "success");
    }
    closeNotifyChannelModal();
    await refreshAll();
  } catch (error) {
    showMessage(error.message, "error");
  }
}

async function handleDeleteNotifyChannel(channelId, channelName) {
  if (!window.confirm(`确认删除通知渠道 ${channelName} 吗？`)) {
    return;
  }
  try {
    await api(`/api/v1/notification-channels/${channelId}`, { method: "DELETE" });
    showMessage("通知渠道已删除", "success");
    await refreshAll();
  } catch (error) {
    showMessage(error.message, "error");
  }
}

async function handleTriggerProject(projectId, projectName) {
  try {
    const run = await api(`/api/v1/projects/${projectId}/trigger`, { method: "POST" });
    showMessage(`已触发部署，运行号 #${run.id}`, "success");
    state.selectedRunId = run.id;
    await refreshAll();
  } catch (error) {
    showMessage(error.message, "error");
  }
}

function textToLines(text) {
  if (typeof text !== "string") {
    return [];
  }
  return text.split(/\r?\n/).map((line) => line.trim()).filter(Boolean);
}

function showMessage(text, type) {
  els.messageBar.textContent = text;
  els.messageBar.className = `message-bar ${type || "success"}`;
  window.clearTimeout(showMessage.timer);
  showMessage.timer = window.setTimeout(() => {
    els.messageBar.className = "message-bar hidden";
  }, 4200);
}

async function api(url, options = {}) {
  const config = {
    method: options.method || "GET",
    headers: {},
  };

  // 添加JWT token（除了登录接口）
  if (state.token && !url.includes("/admin/login")) {
    config.headers["Authorization"] = `Bearer ${state.token}`;
  }

  if (options.body !== undefined) {
    config.headers["Content-Type"] = "application/json";
    config.body = JSON.stringify(options.body);
  }

  const response = await fetch(url, config);
  const raw = await response.text();
  const data = raw ? JSON.parse(raw) : null;
  if (!response.ok) {
    throw new Error(data?.error || `请求失败: ${response.status}`);
  }
  return data;
}

function createMicroButton(label, handler, danger = false) {
  const button = document.createElement("button");
  button.type = "button";
  button.className = `micro-button${danger ? " danger" : ""}`;
  button.textContent = label;
  button.addEventListener("click", handler);
  return button;
}

function statusText(status) {
  return {
    queued: "排队中",
    running: "运行中",
    success: "成功",
    failed: "失败",
  }[status] || status || "-";
}

function statusClass(status) {
  return {
    queued: "status-queued",
    running: "status-running",
    success: "status-success",
    failed: "status-failed",
  }[status] || "status-queued";
}

function isTerminalStatus(status) {
  return status === "success" || status === "failed";
}

function formatDateTime(value) {
  if (!value) {
    return "-";
  }
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) {
    return value;
  }
  return new Intl.DateTimeFormat("zh-CN", {
    year: "numeric",
    month: "2-digit",
    day: "2-digit",
    hour: "2-digit",
    minute: "2-digit",
    second: "2-digit",
  }).format(date);
}

function escapeHTML(value) {
  return String(value ?? "")
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}

// 主机模态框函数
function openHostModalForCreate() {
  els.hostFormId.value = "";
  els.hostName.value = "";
  els.hostAddress.value = "";
  els.hostPort.value = "22";
  els.hostUsername.value = "";
  els.hostPassword.value = "";
  els.hostModalTitle.textContent = "新增主机";
  els.hostModal.style.display = "flex";
}

function openHostModalForEdit(host) {
  els.hostFormId.value = String(host.id);
  els.hostName.value = host.name || "";
  els.hostAddress.value = host.address || "";
  els.hostPort.value = host.port || 22;
  els.hostUsername.value = host.username || "";
  els.hostPassword.value = "";
  els.hostModalTitle.textContent = "编辑主机";
  els.hostModal.style.display = "flex";
}

function closeHostModal() {
  els.hostModal.style.display = "none";
  els.hostForm.reset();
  els.hostFormId.value = "";
  els.hostPort.value = "22";
}

// 项目模态框函数
async function openProjectModalForCreate() {
  els.projectId.value = "";
  els.projectName.value = "";
  els.projectBranch.value = "";
  els.projectRepoURL.value = "";
  els.projectDescription.value = "";
  els.timeoutMinutes.value = "30";
  els.webhookToken.value = "";

  // 重置编译配置
  els.buildImage.value = "";
  els.buildCommands.value = "";
  els.artifactFilterMode.value = "none";
  els.artifactRules.value = "";

  // 重置部署配置
  els.deployHostId.value = "";
  els.remoteSaveDir.value = "";
  els.remoteDeployDir.value = "";
  els.versionCount.value = "5";
  els.preDeployCommands.value = "";
  els.postDeployCommands.value = "";
  els.notifyWebhookURL.value = "";
  els.notifyBearerToken.value = "";

  await renderHostOptionsForConfig();
  els.projectModalTitle.textContent = "新增项目";
  switchTab("basic");
  els.projectModal.style.display = "flex";
}

async function openProjectModalForEdit(project) {
  els.projectId.value = String(project.id);
  els.projectName.value = project.name || "";
  els.projectBranch.value = project.branch || "";
  els.projectRepoURL.value = project.repo_url || "";
  els.projectDescription.value = project.description || "";
  els.webhookToken.value = project.webhook_token || "";

  // 获取完整的项目详情
  try {
    const detail = await api(`/api/v1/projects/${project.id}`);
    const config = detail.deploy_config;

    if (config) {
      // 基础信息
      els.timeoutMinutes.value = String(Math.floor((config.timeout_seconds || 1800) / 60));

      // 编译配置
      els.buildImage.value = config.build_image || "";
      els.buildCommands.value = (config.build_commands || []).join("\n");
      els.artifactFilterMode.value = config.artifact_filter_mode || "none";
      els.artifactRules.value = (config.artifact_rules || []).join("\n");

      // 部署配置
      els.deployHostId.value = String(config.host_id);
      els.remoteSaveDir.value = config.remote_save_dir || "";
      els.remoteDeployDir.value = config.remote_deploy_dir || "";
      els.versionCount.value = String(config.version_count || 5);
      els.preDeployCommands.value = (config.pre_deploy_commands || []).join("\n");
      els.postDeployCommands.value = (config.post_deploy_commands || []).join("\n");
      els.notifyWebhookURL.value = config.notify_webhook_url || "";
      els.notifyBearerToken.value = "";
    } else {
      // 重置配置表单为默认值
      els.timeoutMinutes.value = "30";
      els.buildImage.value = "";
      els.buildCommands.value = "";
      els.artifactFilterMode.value = "none";
      els.artifactRules.value = "";
      els.deployHostId.value = "";
      els.remoteSaveDir.value = "";
      els.remoteDeployDir.value = "";
      els.versionCount.value = "5";
      els.preDeployCommands.value = "";
      els.postDeployCommands.value = "";
      els.notifyWebhookURL.value = "";
      els.notifyBearerToken.value = "";
    }
  } catch (error) {
    console.error("Failed to load project details:", error);
  }

  await renderHostOptionsForConfig();
  els.projectModalTitle.textContent = "编辑项目";
  switchTab("basic");
  els.projectModal.style.display = "flex";
}

function closeProjectModal() {
  els.projectModal.style.display = "none";
  els.projectForm.reset();
  els.buildForm.reset();
  els.deployForm.reset();
  els.projectId.value = "";
}

function renderHostOptionsForConfig() {
  const currentValue = els.deployHostId.value;
  els.deployHostId.innerHTML = '<option value="">请选择主机</option>';
  (state.hosts || []).forEach((host) => {
    const option = document.createElement("option");
    option.value = String(host.id);
    option.textContent = `${host.name} / ${host.username}@${host.address}:${host.port}`;
    els.deployHostId.append(option);
  });
  if (currentValue) {
    els.deployHostId.value = currentValue;
  }
}

// 认证相关函数
function showLoginScreen() {
  if (els.loginScreen) {
    els.loginScreen.style.display = "flex";
    els.loginScreen.classList.add("active");
  }
  if (els.pageShell) {
    els.pageShell.style.display = "none";
    els.pageShell.classList.remove("visible");
  }
}

function showMainApp() {
  if (els.loginScreen) {
    els.loginScreen.style.display = "none";
    els.loginScreen.classList.remove("active");
  }
  if (els.pageShell) {
    els.pageShell.style.display = "block";
    els.pageShell.classList.add("visible");
  }
  switchView(getInitialView(), false);
  refreshAll();
}

async function handleLogin(event) {
  event.preventDefault();
  const username = els.loginUsername.value.trim();
  const password = els.loginPassword.value.trim();

  try {
    const response = await api("/api/v1/admin/login", {
      method: "POST",
      body: { username, password }
    });

    if (response.token) {
      state.token = response.token;
      state.isAuthenticated = true;
      localStorage.setItem("jwt_token", response.token);

      showMainApp();
      showMessage("登录成功", "success");
    } else {
      throw new Error("登录失败，未收到token");
    }
  } catch (error) {
    showLoginMessage(error.message || "登录失败", "error");
  }
}

function showLoginMessage(text, type) {
  if (els.loginMessage) {
    els.loginMessage.textContent = text;
    els.loginMessage.className = `message-bar ${type || "success"}`;
    window.clearTimeout(showLoginMessage.timer);
    showLoginMessage.timer = window.setTimeout(() => {
      els.loginMessage.className = "message-bar hidden";
    }, 4200);
  }
}

function logout() {
  state.token = null;
  state.isAuthenticated = false;
  localStorage.removeItem("jwt_token");
  showLoginScreen();
  showMessage("已退出登录", "success");
}
