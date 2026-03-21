import type { ReactNode } from 'react';
import { Modal } from 'antd';

interface BasicModalProps {
  open: boolean;
  title: string;
  width?: number;
  footer?: ReactNode | null;
  confirmLoading?: boolean;
  onOk?: () => void;
  onCancel: () => void;
  children: ReactNode;
}

export function BasicModal({
  open,
  title,
  width = 640,
  footer,
  confirmLoading,
  onOk,
  onCancel,
  children,
}: BasicModalProps) {
  return (
    <Modal
      destroyOnClose
      open={open}
      title={title}
      width={width}
      footer={footer}
      confirmLoading={confirmLoading}
      onOk={onOk}
      onCancel={onCancel}
    >
      {children}
    </Modal>
  );
}
