/**
 * 积木区DevOps流水线 - 前端应用
 * 采用现代化设计风格，支持主机管理、项目配置和部署监控
 */
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
  isAutoSelecting: false, // 防止循环选择
};

// 移除轮询计时器，改用手动刷新
let runStream = null;
let streamingRunId = null;
const els = {};

/**
 * 应用初始化
 */
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
});

/**
 * 全局事件监听
 */
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
  }
});

/**
 * 绑定DOM元素引用
 */
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
    projectCount: document.getElementById("projectCount"),
    hostCount: document.getElementById("hostCount"),
    runCount: document.getElementById("runCount"),
    notifyCount: document.getElementById("notifyCount"),
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
    buildImage: document.getElementById("buildImage"),
    buildCommands: document.getElementById("buildCommands"),
    artifactRules: document.getElementById("artifactRules"),
    remoteSaveDir: document.getElementById("remoteSaveDir"),
    remoteDeployDir: document.getElementById("remoteDeployDir"),
    preDeployCommands: document.getElementById("preDeployCommands"),
    postDeployCommands: document.getElementById("postDeployCommands"),
    versionCount: document.getElementById("versionCount"),
    notificationChannelId: document.getElementById("notificationChannelId"),
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
    notifyChannelForm: document.getElementById("notifyChannelForm"),
    notifyChannelName: document.getElementById("notifyChannelName"),
    notifyChannelType: document.getElementById("notifyChannelType"),
    notifyChannelWebhookURL: document.getElementById("notifyChannelWebhookURL"),
    notifyChannelSecret: document.getElementById("notifyChannelSecret"),
    notifyChannelRemark: document.getElementById("notifyChannelRemark"),
    loadMoreLogsBtn: document.getElementById("loadMoreLogsBtn"),
  });
}

/**
 * 绑定事件监听器
 */
function bindEvents() {
  // 登录事件
  if (els.loginForm) {
    els.loginForm.addEventListener("submit", handleLogin);
  }

  // 退出登录事件
  if (els.logoutBtn) {
    els.logoutBtn.addEventListener("click", logout);
  }

  // 移动端菜单切换
  const mobileMenuBtn = document.getElementById("mobileMenuBtn");
  const viewNav = document.getElementById("viewNav");

  if (mobileMenuBtn && viewNav) {
    mobileMenuBtn.addEventListener("click", () => {
      viewNav.classList.toggle("mobile-open");
      // 切换图标
      const svg = mobileMenuBtn.querySelector("svg");
      if (viewNav.classList.contains("mobile-open")) {
        // 显示关闭图标
        svg.innerHTML = '<line x1="18" y1="6" x2="6" y2="18"></line><line x1="6" y1="6" x2="18" y2="18"></line>';
      } else {
        // 显示菜单图标
        svg.innerHTML = '<line x1="3" y1="6" x2="21" y2="6"></line><line x1="3" y1="12" x2="21" y2="12"></line><line x1="3" y1="18" x2="21" y2="18"></line>';
      }
    });

    // 点击导航按钮后关闭移动端菜单
    viewNav.querySelectorAll(".nav-button").forEach((button) => {
      button.addEventListener("click", () => {
        if (window.innerWidth <= 768) {
          viewNav.classList.remove("mobile-open");
          const svg = mobileMenuBtn.querySelector("svg");
          svg.innerHTML = '<line x1="3" y1="6" x2="21" y2="6"></line><line x1="3" y1="12" x2="21" y2="12"></line><line x1="3" y1="18" x2="21" y2="18"></line>';
        }
      });
    });
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
  els.cancelProjectModalBtn2?.addEventListener("click", closeProjectModal);
  els.cancelProjectModalBtn3?.addEventListener("click", closeProjectModal);
  els.saveProjectBtn.addEventListener("click", handleProjectSubmit);

  // 项目表单导航按钮
  document.getElementById("basicNextBtn")?.addEventListener("click", () => switchTab("build"));
  document.getElementById("buildPrevBtn")?.addEventListener("click", () => switchTab("basic"));
  document.getElementById("buildNextBtn")?.addEventListener("click", () => switchTab("deploy"));
  document.getElementById("deployPrevBtn")?.addEventListener("click", () => switchTab("build"));

  // 通知渠道事件
  els.addNotifyChannelBtn.addEventListener("click", openNotifyChannelModal);
  els.notifyChannelForm.addEventListener("submit", handleNotifyChannelSubmit);
  els.closeNotifyChannelModalBtn.addEventListener("click", closeNotifyChannelModal);
  els.cancelNotifyChannelModalBtn.addEventListener("click", closeNotifyChannelModal);

  // 部署记录事件
  if (els.refreshLogsBtn) {
    els.refreshLogsBtn.addEventListener("click", refreshAll);
  }
  if (els.loadMoreLogsBtn) {
    els.loadMoreLogsBtn.addEventListener("click", refreshAll);
  }

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

/**
 * 切换Tab页
 * @param {string} tabName - Tab名称
 */
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

/**
 * 获取初始视图
 * @returns {string} 视图名称
 */
function getInitialView() {
  const raw = window.location.hash.replace("#", "").trim();
  return ["home", "hosts", "projects", "notifications", "logs"].includes(raw) ? raw : "home";
}

/**
 * 切换视图
 * @param {string} view - 视图名称
 * @param {boolean} updateHash - 是否更新URL哈希
 */
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

/**
 * 刷新所有数据
 */
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
    renderNotificationChannelOptionsForConfig();
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

/**
 * 渲染概览数据
 */
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
  if (!state.hosts || !state.hosts.length) {
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
    card.style.position = "relative";

    const info = document.createElement("div");
    info.innerHTML = `
      <p class="item-title">${escapeHTML(project.name)}</p>
      <p class="item-subtitle">${escapeHTML(project.branch)}</p>
      <p class="item-meta">${escapeHTML(project.repo_url)}</p>
    `;

    // 删除按钮放在右上角
    const deleteBtn = document.createElement("button");
    deleteBtn.className = "card-delete-btn";
    deleteBtn.innerHTML = "×";
    deleteBtn.title = "删除项目";
    deleteBtn.onclick = (event) => {
      event.stopPropagation();
      handleDeleteProject(project.id, project.name);
    };

    const actions = document.createElement("div");
    actions.className = "item-actions item-actions-full";

    const buttons = [];

    // 1. 部署按钮（如果有部署配置）
    if (project.has_deploy_config) {
      buttons.push(
        createActionButton("部署", "deploy", (event) => {
          event.stopPropagation();
          handleDeployProject(project.id, project.name);
        })
      );
    }

    // 2. Webhook按钮
    buttons.push(
      createActionButton("Webhook", "webhook", (event) => {
        event.stopPropagation();
        const webhookUrl = `${window.location.origin}/api/v1/webhooks/${project.webhook_token}`;
        navigator.clipboard.writeText(webhookUrl).then(() => {
          showMessage("Webhook URL已复制到剪贴板", "success");
        }).catch(() => {
          showMessage("复制失败，请手动复制", "error");
        });
      })
    );

    // 3. 配置按钮
    buttons.push(
      createActionButton("配置", "config", (event) => {
        event.stopPropagation();
        openProjectModalForEdit(project);
      })
    );

    actions.append(...buttons);

    const head = document.createElement("div");
    head.className = "item-head";
    head.append(info, deleteBtn);

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
      createMicroButton("测试", (event) => {
        event.stopPropagation();
        handleTestNotifyChannel(channel.id, channel.name);
      }),
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

  // 如果没有选中的记录，默认选中最后一条
  let needLoadDetail = false;
  if (!state.selectedRunId) {
    state.selectedRunId = runs[0].id;
    needLoadDetail = true;
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
    // 根据触发类型显示不同的信息
    let triggerInfo = "";
    if (run.trigger_type === "webhook") {
      triggerInfo = `Webhook ${run.trigger_ref ? `/ ${escapeHTML(run.trigger_ref)}` : ""}`;
    } else if (run.trigger_type === "manual") {
      triggerInfo = "手动触发";
    } else {
      triggerInfo = run.trigger_type || "-";
    }

    info.innerHTML = `
      <p class="item-title">#${run.id} ${escapeHTML(project.name)}</p>
      <p class="item-subtitle">${escapeHTML(project.branch)} / ${triggerInfo}</p>
      <p class="item-meta">${formatDateTime(run.created_at)}</p>
    `;

    const badge = document.createElement("span");
    badge.className = `status-chip ${statusClass(run.status)}`;
    badge.textContent = statusText(run.status);

    head.append(info, badge);
    card.append(head);
    els.runList.append(card);
  });

  // 渲染详细信息
  renderRunDetail();

  // 如果需要加载详细信息且是自动选择的情况，则异步加载
  if (needLoadDetail && state.selectedRunId && !state.runDetail) {
    loadRunDetail(state.selectedRunId);
  }
}

// 新增：加载运行详细信息（不触发renderRuns循环）
async function loadRunDetail(runId) {
  try {
    state.runDetail = await api(`/api/v1/runs/${runId}`);
    startRunStream(runId);
    renderRunDetail(); // 只渲染详细信息，不重新渲染列表
  } catch (error) {
    console.error("加载运行详细信息失败:", error);
  }
}

function renderRunDetail() {
  if (!state.runDetail) {
    els.runLogOutput.textContent = "尚未加载日志。";
    return;
  }

  const run = state.runDetail;

  const sections = [];
  if (run.error_message) {
    sections.push(`[ERROR]\n${run.error_message}`);
  }
  sections.push(run.log_text || "暂无日志输出。");
  els.runLogOutput.textContent = sections.join("\n\n");

  // 自动滚动到底部显示最新日志
  els.runLogOutput.scrollTop = els.runLogOutput.scrollHeight;
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

  runStream = new EventSource(`/api/v1/runs/${runId}/stream?token=${state.token}`);
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
  // 状态显示元素已移除，此函数保留以保持兼容性
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
      host_id: parseInt(els.deployHostId.value),
      build_image: els.buildImage.value.trim(),
      build_commands: textToLines(els.buildCommands.value),
      artifact_filter_mode: document.querySelector('input[name="artifactFilterMode"]:checked')?.value || "include",
      artifact_rules: textToLines(els.artifactRules.value),
      remote_save_dir: els.remoteSaveDir.value.trim(),
      remote_deploy_dir: els.remoteDeployDir.value.trim(),
      pre_deploy_commands: textToLines(els.preDeployCommands.value),
      post_deploy_commands: textToLines(els.postDeployCommands.value),
      timeout_seconds: parseInt(els.timeoutMinutes.value) * 60, // 转换为秒
    };

    // 处理通知渠道选择
    const channelValue = els.notificationChannelId.value;
    if (channelValue === "-1") {
      configPayload.notification_channel_id = null; // 不通知
    } else if (channelValue !== "") {
      configPayload.notification_channel_id = parseInt(channelValue); // 指定渠道
    }
    // channelValue === "" 时，不设置该字段，使用默认渠道

    await api(`/api/v1/projects/${project.id}/deploy-config`, { method: "PUT", body: configPayload });

    showMessage(id ? "项目已更新" : "项目已创建", "success");
    closeProjectModal();
    await refreshAll();
  } catch (error) {
    // 处理各种错误情况
    let errorMessage = error.message;

    // UNIQUE约束错误
    if (errorMessage.includes("UNIQUE constraint failed")) {
      errorMessage = "该仓库地址和分支组合已存在，请勿重复创建项目";
    }
    // host_id为空或无效
    else if (errorMessage.includes("host_id is required") || errorMessage.includes("host_id")) {
      errorMessage = "请选择目标主机";
    }
    // 其他验证错误
    else if (errorMessage.includes("is required") || errorMessage.includes("required")) {
      errorMessage = `缺少必填字段: ${errorMessage}`;
    }

    showMessage(errorMessage, "error");
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
  document.getElementById("notifyChannelWebhookURL").value = "";
  document.getElementById("notifyChannelSecret").value = "";
  document.getElementById("notifyChannelRemark").value = "";
  els.notifyChannelModal.style.display = "flex";
}

async function openNotifyChannelModalForEdit(channel) {
  els.notifyChannelModalTitle.textContent = "编辑通知渠道";
  document.getElementById("notifyChannelId").value = channel.id;
  document.getElementById("notifyChannelName").value = channel.name;
  document.getElementById("notifyChannelType").value = channel.type;
  document.getElementById("notifyChannelRemark").value = channel.remark || "";

  // 获取完整配置信息
  try {
    const fullChannel = await api(`/api/v1/notification-channels/${channel.id}`);
    if (fullChannel.config) {
      const config = fullChannel.config;
      // 根据类型填充URL和密钥字段
      switch (channel.type) {
        case "webhook":
          document.getElementById("notifyChannelWebhookURL").value = config.url || "";
          document.getElementById("notifyChannelSecret").value = config.secret || "";
          break;
        case "wechat":
          document.getElementById("notifyChannelWebhookURL").value = config.webhook_url || "";
          document.getElementById("notifyChannelSecret").value = config.key || "";
          break;
        case "dingtalk":
          document.getElementById("notifyChannelWebhookURL").value = config.webhook_url || "";
          document.getElementById("notifyChannelSecret").value = config.secret || "";
          break;
        case "feishu":
          document.getElementById("notifyChannelWebhookURL").value = config.webhook_url || "";
          document.getElementById("notifyChannelSecret").value = "";
          break;
        default:
          document.getElementById("notifyChannelWebhookURL").value = "";
          document.getElementById("notifyChannelSecret").value = "";
      }
    } else {
      document.getElementById("notifyChannelWebhookURL").value = "";
      document.getElementById("notifyChannelSecret").value = "";
    }
  } catch (error) {
    showMessage(error.message, "error");
    document.getElementById("notifyChannelWebhookURL").value = "";
    document.getElementById("notifyChannelSecret").value = "";
  }

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
  const type = document.getElementById("notifyChannelType").value;
  const webhookURL = document.getElementById("notifyChannelWebhookURL").value;
  const secret = document.getElementById("notifyChannelSecret").value;

  // 根据渠道类型构建配置对象
  let config = {};
  switch (type) {
    case "webhook":
      config = {
        url: webhookURL,
        token: "",
        secret: secret
      };
      break;
    case "wechat":
      config = {
        webhook_url: webhookURL,
        key: secret
      };
      break;
    case "dingtalk":
      config = {
        webhook_url: webhookURL,
        secret: secret
      };
      break;
    case "feishu":
      config = {
        webhook_url: webhookURL
      };
      break;
    default:
      showMessage("不支持的渠道类型", "error");
      return;
  }

  const channelPayload = {
    name: document.getElementById("notifyChannelName").value,
    type: type,
    config: config,
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

async function handleTestNotifyChannel(channelId, channelName) {
  try {
    await api(`/api/v1/notification-channels/${channelId}/test`, {
      method: "POST",
      body: {
        title: "测试通知",
        content: "这是一条来自积木区DevOps的测试通知"
      }
    });
    showMessage(`已发送测试通知到 ${channelName}`, "success");
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

async function handleDeployProject(projectId, projectName) {
  // 先确认是否开始部署
  if (!window.confirm(`确认开始手动部署项目 "${projectName}" 吗？`)) {
    return;
  }

  try {
    const run = await api(`/api/v1/projects/${projectId}/trigger`, { method: "POST" });
    showMessage(`部署已触发，运行号 #${run.id}`, "success");

    // 询问是否查看部署日志
    if (window.confirm("部署已成功触发！是否立即查看部署日志？")) {
      // 跳转到部署记录页面并选中该运行记录
      await refreshAll();
      window.location.hash = "logs";

      // 等待页面加载后选中该运行记录
      setTimeout(async () => {
        await selectRun(run.id);
      }, 100);
    } else {
      // 如果不查看日志，也刷新数据
      await refreshAll();
    }
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

/**
 * API请求封装
 * @param {string} url - 请求URL
 * @param {Object} options - 请求选项
 * @returns {Promise<any>} 响应数据
 */
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

function createActionButton(label, type, handler) {
  const button = document.createElement("button");
  button.type = "button";
  button.className = `action-button action-button-${type}`;
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
  els.remoteSaveDir.value = "/data/jimuqu/projects"; // 设置默认值
  els.remoteDeployDir.value = "";
  els.versionCount.value = "5";
  els.preDeployCommands.value = "";
  els.postDeployCommands.value = "";
  els.notificationChannelId.value = ""; // 使用默认渠道

  await renderHostOptionsForConfig();
  await renderNotificationChannelOptionsForConfig();
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
      // 设置制品过滤模式单选按钮
      const artifactFilterMode = config.artifact_filter_mode || "include";
      const radioButton = document.querySelector(`input[name="artifactFilterMode"][value="${artifactFilterMode}"]`);
      if (radioButton) {
        radioButton.checked = true;
      }
      els.artifactRules.value = (config.artifact_rules || []).join("\n");

      // 部署配置
      els.deployHostId.value = String(config.host_id);
      els.remoteSaveDir.value = config.remote_save_dir || "";
      els.remoteDeployDir.value = config.remote_deploy_dir || "";
      els.versionCount.value = String(config.version_count || 5);
      els.preDeployCommands.value = (config.pre_deploy_commands || []).join("\n");
      els.postDeployCommands.value = (config.post_deploy_commands || []).join("\n");

      // 通知渠道设置
      if (config.notification_channel_id === null) {
        els.notificationChannelId.value = "-1"; // 不通知
      } else if (config.notification_channel_id) {
        els.notificationChannelId.value = String(config.notification_channel_id); // 指定渠道
      } else {
        els.notificationChannelId.value = ""; // 使用默认渠道
      }
    } else {
      // 重置配置表单为默认值
      els.timeoutMinutes.value = "30";
      els.buildImage.value = "";
      els.buildCommands.value = "";
      els.artifactFilterMode.value = "none";
      els.artifactRules.value = "";
      els.deployHostId.value = "";
      els.remoteSaveDir.value = "/data/jimuqu/projects"; // 设置默认值
      els.remoteDeployDir.value = "";
      els.versionCount.value = "5";
      els.preDeployCommands.value = "";
      els.postDeployCommands.value = "";
      els.notificationChannelId.value = ""; // 使用默认渠道
    }
  } catch (error) {
    console.error("Failed to load project details:", error);
  }

  await renderHostOptionsForConfig();
  await renderNotificationChannelOptionsForConfig();
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

  // 如果有当前值，使用当前值；否则默认选中第一个主机
  if (currentValue) {
    els.deployHostId.value = currentValue;
  } else if (state.hosts && state.hosts.length > 0) {
    els.deployHostId.value = String(state.hosts[0].id);
  }
}

function renderNotificationChannelOptionsForConfig() {
  const currentValue = els.notificationChannelId.value;
  els.notificationChannelId.innerHTML = '<option value="">使用默认渠道</option><option value="-1">不通知</option>';
  (state.notifyChannels || []).forEach((channel) => {
    const option = document.createElement("option");
    option.value = String(channel.id);
    const defaultMark = channel.is_default ? " (默认)" : "";
    option.textContent = `${channel.name} ${defaultMark}`;
    els.notificationChannelId.append(option);
  });
  if (currentValue) {
    els.notificationChannelId.value = currentValue;
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
