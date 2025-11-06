import type { VbenFormSchema } from '#/adapter/form';
import type { VxeTableGridOptions } from '#/adapter/vxe-table';
import type { Project } from '#/api/project';

import { z } from '#/adapter/form';

/**
 * 获取编辑表单的字段配置
 */
export function useSchema(): VbenFormSchema[] {
  return [
    {
      component: 'Input',
      fieldName: 'name',
      label: '项目名称',
      rules: z
        .string()
        .min(2, '项目名称至少2个字符')
        .max(100, '项目名称不能超过100个字符'),
    },
    {
      component: 'Input',
      fieldName: 'code',
      label: '项目编码',
      rules: z
        .string()
        .min(2, '项目编码至少2个字符')
        .max(50, '项目编码不能超过50个字符')
        .regex(
          /^[a-zA-Z0-9_-]+$/,
          '项目编码只能包含字母、数字、连字符和下划线',
        ),
    },
    {
      component: 'Input',
      fieldName: 'git_repo',
      label: 'Git仓库地址',
      rules: z
        .string()
        .optional()
        .refine(
          (value) => {
            if (!value) return true;
            try {
              new URL(value);
              return true;
            } catch {
              return /^git@[\w.-]+:[\w.-]+\/[\w.-]+\.git$/.test(value);
            }
          },
          {
            message: '请输入有效的Git仓库地址',
          },
        ),
    },
    {
      component: 'Input',
      fieldName: 'git_username',
      label: 'Git用户名',
      rules: z.string().max(100, 'Git用户名不能超过100个字符').optional(),
    },
    {
      component: 'InputPassword',
      fieldName: 'git_password',
      label: 'Git密码/Token',
      rules: z.string().max(200, 'Git密码/Token不能超过200个字符').optional(),
    },
    {
      component: 'InputPassword',
      fieldName: 'webhook_password',
      label: 'Webhook密码',
      rules: z.string().max(100, 'Webhook密码不能超过100个字符').optional(),
    },
    {
      component: 'Textarea',
      componentProps: {
        rows: 3,
        showCount: true,
        maxLength: 500,
      },
      fieldName: 'remark',
      label: '备注',
      rules: z.string().max(500, '备注不能超过500个字符').optional(),
    },
  ];
}

/**
 * 获取表格列配置
 */
export function useColumns(): VxeTableGridOptions<Project>['columns'] {
  return [
    {
      field: 'name',
      title: '项目名称',
      minWidth: 150,
    },
    {
      field: 'code',
      title: '项目编码',
      width: 140,
    },
    {
      field: 'git_repo',
      title: 'Git仓库地址',
      minWidth: 200,
      showOverflow: 'tooltip',
    },
    {
      field: 'git_username',
      title: 'Git用户名',
      width: 120,
      showOverflow: 'tooltip',
      cellRender: {
        name: 'CellText',
        props: {
          attrs: (row: Project) => ({
            text: row.git_username || '未配置',
          }),
        },
      },
    },
    {
      field: 'remark',
      title: '备注',
      minWidth: 150,
      showOverflow: 'tooltip',
      cellRender: {
        name: 'CellText',
        props: {
          attrs: (row: Project) => ({
            text: row.remark || '无备注',
          }),
        },
      },
    },
    {
      field: 'webhook_url',
      title: 'Webhook URL',
      minWidth: 250,
      showOverflow: 'tooltip',
      cellRender: {
        name: 'CellText',
        props: {
          attrs: (row: Project) => {
            const baseUrl = window.location.origin;
            const webhookUrl = `${baseUrl}/api/webhook/${row.code}`;
            return {
              text: webhookUrl,
              title: webhookUrl,
            };
          },
        },
      },
    },
    {
      field: 'created_at',
      title: '创建时间',
      width: 160,
      cellRender: {
        name: 'CellDatetime',
        props: {
          attrs: {
            format: 'YYYY-MM-DD HH:mm:ss',
          },
        },
      },
    },
    {
      field: 'action',
      fixed: 'right',
      slots: { default: 'action' },
      title: '操作',
      resizable: false,
      width: 450,
    },
  ];
}
