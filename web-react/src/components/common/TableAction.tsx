import { Button, Popconfirm, Space } from 'antd';
import { PermissionGate } from './PermissionGate';

export interface TableActionItem {
  key: string;
  label: string;
  onClick?: () => void;
  danger?: boolean;
  permission?: string | string[];
  hidden?: boolean;
  confirmTitle?: string;
}

interface TableActionProps {
  actions: TableActionItem[];
}

export function TableAction({ actions }: TableActionProps) {
  return (
    <Space size="small">
      {actions
        .filter((action) => !action.hidden)
        .map((action) => {
          const button = (
            <Button
              danger={action.danger}
              size="small"
              type="link"
              onClick={action.confirmTitle ? undefined : action.onClick}
            >
              {action.label}
            </Button>
          );

          return (
            <PermissionGate key={action.key} permission={action.permission}>
              {action.confirmTitle ? (
                <Popconfirm title={action.confirmTitle} onConfirm={action.onClick}>
                  {button}
                </Popconfirm>
              ) : (
                button
              )}
            </PermissionGate>
          );
        })}
    </Space>
  );
}
