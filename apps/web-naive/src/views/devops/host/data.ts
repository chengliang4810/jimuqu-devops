import type { VbenFormSchema } from '#/adapter/form';
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { Host } from '#/api/host';

import { z } from '#/adapter/form';

/**
 * 获取编辑表单的字段配置
 */
export function useSchema(): VbenFormSchema[] {
  return [
    {
      component: 'Input',
      fieldName: 'name',
      label: '主机名称',
      rules: z
        .string()
        .min(2, '主机名称至少2个字符')
        .max(50, '主机名称不能超过50个字符'),
    },
    {
      component: 'Input',
      fieldName: 'host',
      label: 'IP地址',
      rules: z
        .string()
        .min(1, 'IP地址不能为空')
        .regex(
          /^(\d{1,3}\.){3}\d{1,3}$|^([a-z0-9-]+\.)+[a-z]{2,}$/i,
          '请输入有效的IP地址或域名',
        ),
    },
    {
      component: 'InputNumber',
      fieldName: 'port',
      label: 'SSH端口',
      defaultValue: 22,
      rules: z
        .number()
        .min(1, '端口必须大于0')
        .max(65_535, '端口不能超过65535'),
    },
    {
      component: 'Input',
      fieldName: 'username',
      label: '用户名',
      rules: z
        .string()
        .min(1, '用户名不能为空')
        .max(50, '用户名不能超过50个字符'),
    },
    {
      component: 'RadioGroup',
      componentProps: {
        buttonStyle: 'solid',
        options: [
          { label: '密码认证', value: 'password' },
          // { label: '密钥认证', value: 'key' },
        ],
        optionType: 'button',
      },
      defaultValue: 'password',
      fieldName: 'auth_type',
      label: '认证方式',
    },
    {
      component: 'InputPassword',
      fieldName: 'password',
      label: '密码',
      rules: z
        .string()
        .min(1, '密码不能为空')
        .max(100, '密码不能超过100个字符'),
      dependencies: {
        triggerFields: ['auth_type'],
        show: (values) => values.auth_type === 'password',
      },
    },
    {
      component: 'Textarea',
      componentProps: {
        rows: 3,
        showCount: true,
        maxLength: 200,
      },
      fieldName: 'remark',
      label: '备注',
      rules: z.string().max(200, '备注不能超过200个字符').optional(),
    },
  ];
}

/**
 * 获取表格列配置
 */
export function useColumns(): VxeTableGridOptions<Host>['columns'] {
  return [
    {
      type: 'checkbox',
      width: 50,
    },
    {
      field: 'name',
      title: '主机名称',
      minWidth: 120,
    },
    {
      field: 'host',
      title: 'IP地址',
      width: 140,
    },
    {
      field: 'port',
      title: '端口',
      width: 80,
    },
    {
      field: 'username',
      title: '用户名',
      width: 100,
    },
    {
      field: 'auth_type',
      title: '认证方式',
      width: 120,
      cellRender: {
        name: 'CellNTag',
        props: {
          attrs: (row: Host) => ({
            type: row.auth_type === 'password' ? 'info' : 'success',
            size: 'small',
            round: true,
            text: row.auth_type === 'password' ? '密码认证' : '密钥认证',
          }),
        },
      },
    },
    {
      field: 'status',
      title: '状态',
      width: 100,
      cellRender: {
        name: 'CellNTag',
        props: {
          attrs: (row: Host) => ({
            type: row.status === 'online' ? 'success' : 'error',
            size: 'small',
            round: true,
            text: row.status === 'online' ? '在线' : '离线',
          }),
        },
      },
    },
    {
      field: 'remark',
      title: '备注',
      minWidth: 150,
      showOverflow: 'tooltip',
    },
    {
      field: 'action',
      fixed: 'right',
      slots: { default: 'action' },
      title: '操作',
      resizable: false,
      width: 'auto',
    },
  ];
}
