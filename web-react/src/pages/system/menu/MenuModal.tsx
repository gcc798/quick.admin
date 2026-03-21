import { useEffect, useMemo, useState } from 'react';
import { App, Form } from 'antd';
import type { FormSchema } from '@/types/form';
import type { MenuRecord } from '@/types/menu';
import { menuApi } from '@/api/menu';
import { BasicForm } from '@/components/common/BasicForm';
import { BasicModal } from '@/components/common/BasicModal';

interface MenuModalProps {
  open: boolean;
  menuId?: number;
  parentId?: number;
  onCancel: () => void;
  onSuccess: () => void;
}

function toTreeSelect(nodes: MenuRecord[]): Record<string, unknown>[] {
  return nodes.map((node) => ({
    title: node.menuName,
    value: node.id,
    key: node.id,
    children: node.children ? toTreeSelect(node.children) : undefined,
  }));
}

export function MenuModal({
  open,
  menuId,
  parentId,
  onCancel,
  onSuccess,
}: MenuModalProps) {
  const [form] = Form.useForm<Partial<MenuRecord>>();
  const { message } = App.useApp();
  const [loading, setLoading] = useState(false);
  const [menuOptions, setMenuOptions] = useState<Record<string, unknown>[]>([]);

  const schemas = useMemo<FormSchema[]>(
    () => [
      {
        name: 'parentId',
        label: '上级菜单',
        component: 'TreeSelect',
        initialValue: parentId ?? 0,
        props: {
          treeData: [{ title: '主类目', value: 0, key: 0 }, ...menuOptions],
          treeDefaultExpandAll: true,
          placeholder: '请选择上级菜单',
        },
      },
      {
        name: 'menuType',
        label: '菜单类型',
        component: 'RadioGroup',
        initialValue: 0,
        props: {
          optionType: 'button',
          buttonStyle: 'solid',
          options: [
            { label: '目录', value: 0 },
            { label: '菜单', value: 1 },
            { label: '按钮', value: 2 },
          ],
        },
      },
      {
        name: 'icon',
        label: '菜单图标',
        component: 'IconPicker',
        hidden: (values) => values.menuType === 2,
      },
      {
        name: 'menuName',
        label: '菜单名称',
        component: 'Input',
        rules: [{ required: true, message: '请输入菜单名称' }],
      },
      {
        name: 'sort',
        label: '显示排序',
        component: 'InputNumber',
        initialValue: 0,
        props: { min: 0 },
      },
      {
        name: 'path',
        label: '路由地址',
        component: 'Input',
        hidden: (values) => values.menuType === 2,
        helpMessage: '例如：system 或 user',
      },
      {
        name: 'component',
        label: '组件路径',
        component: 'Input',
        hidden: (values) => values.menuType !== 1,
        helpMessage: '例如：system/user/index',
      },
      {
        name: 'perms',
        label: '权限标识',
        component: 'Input',
        hidden: (values) => values.menuType === 0,
      },
      {
        name: 'query',
        label: '路由参数',
        component: 'Input',
        hidden: (values) => values.menuType !== 1,
      },
      {
        name: 'isFrame',
        label: '是否外链',
        component: 'RadioGroup',
        initialValue: 0,
        hidden: (values) => values.menuType === 2,
        props: {
          options: [
            { label: '否', value: 0 },
            { label: '是', value: 1 },
          ],
        },
      },
      {
        name: 'isCache',
        label: '是否缓存',
        component: 'RadioGroup',
        initialValue: 0,
        hidden: (values) => values.menuType !== 1,
        props: {
          options: [
            { label: '不缓存', value: 0 },
            { label: '缓存', value: 1 },
          ],
        },
      },
      {
        name: 'visible',
        label: '显示状态',
        component: 'RadioGroup',
        initialValue: 0,
        hidden: (values) => values.menuType === 2,
        props: {
          options: [
            { label: '显示', value: 0 },
            { label: '隐藏', value: 1 },
          ],
        },
      },
      {
        name: 'status',
        label: '菜单状态',
        component: 'RadioGroup',
        initialValue: 0,
        props: {
          options: [
            { label: '正常', value: 0 },
            { label: '停用', value: 1 },
          ],
        },
      },
      {
        name: 'remark',
        label: '备注',
        component: 'TextArea',
        props: { rows: 3 },
      },
    ],
    [menuOptions, parentId],
  );

  useEffect(() => {
    if (!open) {
      form.resetFields();
      return;
    }

    void (async () => {
      const tree = await menuApi.getMenuTree();
      setMenuOptions(toTreeSelect(tree));
    })();

    if (!menuId) {
      form.resetFields();
      form.setFieldsValue({
        parentId: parentId ?? 0,
        menuType: 0,
        sort: 0,
        isFrame: 0,
        isCache: 0,
        visible: 0,
        status: 0,
      });
      return;
    }

    void (async () => {
      setLoading(true);
      try {
        const detail = await menuApi.detail(menuId);
        form.setFieldsValue(detail);
      } finally {
        setLoading(false);
      }
    })();
  }, [form, menuId, open, parentId]);

  const handleSubmit = async () => {
    const values = await form.validateFields();
    setLoading(true);
    try {
      if (menuId) {
        await menuApi.update(menuId, values);
        message.success('菜单更新成功');
      } else {
        await menuApi.create(values);
        message.success('菜单创建成功');
      }
      onSuccess();
    } finally {
      setLoading(false);
    }
  };

  return (
    <BasicModal
      open={open}
      title={menuId ? '编辑菜单' : '新增菜单'}
      width={760}
      confirmLoading={loading}
      onCancel={onCancel}
      onOk={() => void handleSubmit()}
    >
      <BasicForm
        form={form}
        schemas={schemas}
        layout="vertical"
        showActionButtons={false}
      />
    </BasicModal>
  );
}
