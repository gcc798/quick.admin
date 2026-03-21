import { useEffect, useState } from 'react';
import type { Key } from 'react';
import { App, Spin, Tree } from 'antd';
import { BasicModal } from '@/components/common/BasicModal';
import { menuApi } from '@/api/menu';
import { roleApi } from '@/api/role';
import type { MenuRecord } from '@/types/menu';

interface PermissionModalProps {
  open: boolean;
  roleId?: number;
  onCancel: () => void;
  onSuccess: () => void;
}

export function PermissionModal({
  open,
  roleId,
  onCancel,
  onSuccess,
}: PermissionModalProps) {
  const { message } = App.useApp();
  const [loading, setLoading] = useState(false);
  const [treeData, setTreeData] = useState<MenuRecord[]>([]);
  const [checkedKeys, setCheckedKeys] = useState<Key[]>([]);

  useEffect(() => {
    if (!open || !roleId) {
      setCheckedKeys([]);
      return;
    }

    void (async () => {
      setLoading(true);
      try {
        const [menus, roleMenus] = await Promise.all([
          menuApi.getMenuTree(),
          roleApi.getMenus(roleId),
        ]);
        setTreeData(menus);
        setCheckedKeys(roleMenus);
      } finally {
        setLoading(false);
      }
    })();
  }, [open, roleId]);

  const handleSubmit = async () => {
    if (!roleId) {
      return;
    }

    setLoading(true);
    try {
      await roleApi.assignMenus(
        roleId,
        checkedKeys.map((key) => Number(key)),
      );
      message.success('角色菜单分配成功');
      onSuccess();
    } finally {
      setLoading(false);
    }
  };

  return (
    <BasicModal
      open={open}
      title="分配菜单权限"
      width={700}
      confirmLoading={loading}
      onCancel={onCancel}
      onOk={() => void handleSubmit()}
    >
      <Spin spinning={loading}>
        <Tree
          checkable
          checkedKeys={checkedKeys}
          defaultExpandAll
          fieldNames={{ title: 'menuName', key: 'id', children: 'children' }}
          selectable={false}
          treeData={treeData}
          onCheck={(keys) => setCheckedKeys(Array.isArray(keys) ? keys : keys.checked)}
        />
      </Spin>
    </BasicModal>
  );
}
