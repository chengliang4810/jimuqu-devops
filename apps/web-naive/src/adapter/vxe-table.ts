import type { VxeTableGridOptions } from '@vben/plugins/vxe-table';

import { h } from 'vue';

import { setupVbenVxeTable, useVbenVxeGrid } from '@vben/plugins/vxe-table';

import { NButton, NImage } from 'naive-ui';

import { useVbenForm } from './form';

setupVbenVxeTable({
  configVxeTable: (vxeUI) => {
    vxeUI.setConfig({
      grid: {
        align: 'center',
        border: false,
        columnConfig: {
          resizable: true,
        },
        minHeight: 180,
        formConfig: {
          // 全局禁用vxe-table的表单配置，使用formOptions
          enabled: false,
        },
        proxyConfig: {
          autoLoad: true,
          response: {
            result: 'list',
            total: 'total',
            list: 'list',
          },
          showActiveMsg: true,
          showResponseMsg: false,
        },
        round: true,
        showOverflow: true,
        size: 'small',
      } as VxeTableGridOptions,
    });

    // 表格配置项可以用 cellRender: { name: 'CellImage' },
    vxeUI.renderer.add('CellImage', {
      renderTableDefault(_renderOpts, params) {
        const { column, row } = params;
        return h(NImage, { src: row[column.field] });
      },
    });

    // 表格配置项可以用 cellRender: { name: 'CellLink' },
    vxeUI.renderer.add('CellLink', {
      renderTableDefault(renderOpts) {
        const { props } = renderOpts;
        return h(
          NButton,
          { size: 'small', type: 'primary', quaternary: true },
          { default: () => props?.text },
        );
      },
    });

    // 表格配置项可以用 cellRender: { name: 'CellOperation' },
    vxeUI.renderer.add('CellOperation', {
      renderTableDefault(renderOpts) {
        const { props } = renderOpts;
        return h(
          'div',
          { style: { display: 'flex', gap: '4px' } },
          props?.options?.map((option: any) => {
            if (typeof option === 'string') {
              // 处理内置按钮类型
              if (option === 'edit') {
                return h(
                  NButton,
                  {
                    size: 'small',
                    type: 'primary',
                    onClick: () =>
                      props.onClick?.({ code: 'edit', row: props.row }),
                  },
                  { default: () => '编辑' },
                );
              }
              if (option === 'delete') {
                return h(
                  NButton,
                  {
                    size: 'small',
                    type: 'error',
                    onClick: () =>
                      props.onClick?.({ code: 'delete', row: props.row }),
                  },
                  { default: () => '删除' },
                );
              }
            }
            // 处理自定义按钮
            if (option.code && option.text) {
              return h(
                NButton,
                {
                  size: 'small',
                  onClick: () =>
                    props.onClick?.({ code: option.code, row: props.row }),
                },
                { default: () => option.text },
              );
            }
            return null;
          }),
        );
      },
    });

    // 表格配置项可以用 cellRender: { name: 'CellTag' },
    vxeUI.renderer.add('CellTag', {
      renderTableDefault(renderOpts) {
        const { props } = renderOpts;
        return h(
          'span',
          {
            style: {
              ...props?.attrs,
              display: 'inline-block',
              padding: '2px 8px',
              borderRadius: '4px',
              fontSize: '12px',
              color: '#fff',
            },
          },
          props?.attrs?.text || '',
        );
      },
    });

    // 这里可以自行扩展 vxe-table 的全局配置，比如自定义格式化
    // vxeUI.formats.add
  },
  useVbenForm,
});

export { useVbenVxeGrid };

export type * from '@vben/plugins/vxe-table';
