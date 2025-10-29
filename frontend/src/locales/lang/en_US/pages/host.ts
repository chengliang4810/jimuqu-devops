import type { Message } from 'tdesign-vue-next';

export default {
  // Host Management Page
  host: {
    title: 'Host Management',
    groupName: 'Host Group',
    groupList: 'Group List',
    createGroup: 'Create Group',
    editGroup: 'Edit Group',
    deleteGroup: 'Delete Group',
    groupNamePlaceholder: 'Please enter group name',
    groupDescPlaceholder: 'Please enter group description (optional)',
    groupNameRequired: 'Please enter group name',
    groupNameLength: 'Group name length is 1-50 characters',
    deleteGroupConfirm: 'Are you sure you want to delete group "{name}"?',
    deleteGroupWarning: 'Note: This group has {count} hosts. After deletion, these hosts will be moved to the default group.',
    defaultGroupNotEdit: 'Default group cannot be edited',
    defaultGroupNotDelete: 'Default group cannot be deleted',

    // Host List
    hostList: 'Host List',
    createHost: 'Create Host',
    editHost: 'Edit Host',
    deleteHost: 'Delete Host',
    deleteHostConfirm: 'Are you sure you want to delete this host?',
    refresh: 'Refresh',
    testConnection: 'Test Connection',

    // Host Form
    hostName: 'Host Name',
    hostNamePlaceholder: 'Please enter host name',
    hostNameRequired: 'Please enter host name',
    hostNameLength: 'Host name length is 1-50 characters',
    ipAddress: 'IP Address',
    ipAddressPlaceholder: 'Please enter IP address',
    ipAddressRequired: 'Please enter IP address',
    ipAddressInvalid: 'Please enter a valid IP address',
    port: 'Port',
    portPlaceholder: 'Please enter port number',
    portRequired: 'Please enter port number',
    portRange: 'Port range is 1-65535',
    username: 'Username',
    usernamePlaceholder: 'Please enter username',
    usernameRequired: 'Please enter username',
    usernameLength: 'Username length is 1-50 characters',
    authType: 'Authentication Type',
    authPassword: 'Password',
    authKey: 'Key',
    password: 'Password',
    passwordPlaceholder: 'Please enter password',
    passwordRequired: 'Please enter password',
    privateKey: 'Private Key',
    privateKeyPlaceholder: 'Please enter private key content',
    privateKeyRequired: 'Please enter private key',
    description: 'Description',
    descriptionPlaceholder: 'Please enter description (optional)',

    // Status
    online: 'Online',
    offline: 'Offline',
    unknown: 'Unknown',

    // Connection Test
    connectionTest: 'Connection Test',
    testResult: 'Test Result',
    connectionSuccess: 'Connection Successful',
    connectionFailed: 'Connection Failed',

    // Success Messages
    createSuccess: 'Created successfully',
    updateSuccess: 'Updated successfully',
    deleteSuccess: 'Deleted successfully',
    sortSuccess: 'Sort updated successfully',

    // Error Messages
    createFailed: 'Failed to create',
    updateFailed: 'Failed to update',
    deleteFailed: 'Failed to delete',
    sortFailed: 'Failed to update sort',
    getListFailed: 'Failed to get list',

    // Empty State
    noData: 'No host data',
    createHostBtn: 'Create Host',
  },
};