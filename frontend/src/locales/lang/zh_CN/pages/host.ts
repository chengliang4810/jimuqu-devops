export default {
  // 主机管理页面
  host: {
    title: '主机管理',
    groupName: '主机分组',
    groupList: '分组列表',
    createGroup: '新建分组',
    editGroup: '编辑分组',
    deleteGroup: '删除分组',
    groupNamePlaceholder: '请输入分组名称',
    groupDescPlaceholder: '请输入分组描述（可选）',
    groupNameRequired: '请输入分组名称',
    groupNameLength: '分组名称长度为1-50个字符',
    deleteGroupConfirm: '确定要删除分组 "{name}" 吗？',
    deleteGroupWarning: '注意：该分组下还有 {count} 台主机，删除分组后这些主机将被移至默认分组。',
    defaultGroupNotEdit: '默认分组不能编辑',
    defaultGroupNotDelete: '默认分组不能删除',

    // 主机列表
    hostList: '主机列表',
    createHost: '新建主机',
    editHost: '编辑主机',
    deleteHost: '删除主机',
    deleteHostConfirm: '确定要删除该主机吗？',
    refresh: '刷新',
    testConnection: '连接测试',

    // 主机表单
    hostName: '主机名称',
    hostNamePlaceholder: '请输入主机名称',
    hostNameRequired: '请输入主机名称',
    hostNameLength: '主机名称长度为1-50个字符',
    ipAddress: 'IP地址',
    ipAddressPlaceholder: '请输入IP地址',
    ipAddressRequired: '请输入IP地址',
    ipAddressInvalid: '请输入有效的IP地址',
    port: '端口',
    portPlaceholder: '请输入端口号',
    portRequired: '请输入端口号',
    portRange: '端口范围为1-65535',
    username: '用户名',
    usernamePlaceholder: '请输入用户名',
    usernameRequired: '请输入用户名',
    usernameLength: '用户名长度为1-50个字符',
    authType: '认证方式',
    authPassword: '密码认证',
    authKey: '密钥认证',
    password: '密码',
    passwordPlaceholder: '请输入密码',
    passwordRequired: '请输入密码',
    privateKey: '私钥',
    privateKeyPlaceholder: '请输入私钥内容',
    privateKeyRequired: '请输入私钥',
    description: '备注',
    descriptionPlaceholder: '请输入备注信息（可选）',

    // 状态
    online: '在线',
    offline: '离线',
    unknown: '未知',

    // 连接测试
    connectionTest: '连接测试',
    testResult: '连接测试结果',
    connectionSuccess: '连接成功',
    connectionFailed: '连接失败',

    // 成功消息
    createSuccess: '创建成功',
    updateSuccess: '更新成功',
    deleteSuccess: '删除成功',
    sortSuccess: '排序更新成功',

    // 错误消息
    createFailed: '创建失败',
    updateFailed: '更新失败',
    deleteFailed: '删除失败',
    sortFailed: '更新排序失败',
    getListFailed: '获取列表失败',

    // 空状态
    noData: '暂无主机数据',
    createHostBtn: '新建主机',
  },
};