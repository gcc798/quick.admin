import { useEffect, useMemo, useState } from 'react';
import { App, Button, Form } from 'antd';
import type { SnowflakeId } from '@/types/api';
import type { FormSchema } from '@/types/form';
import type { StorageEnvRecord } from '@/types/system';
import { BasicForm } from '@/components/common/BasicForm';
import { BasicModal } from '@/components/common/BasicModal';
import { storageEnvApi } from '@/api/storageenv';

interface EnvModalProps {
  open: boolean;
  envId?: SnowflakeId;
  onCancel: () => void;
  onSuccess: () => void;
}

function safeStringify(value: unknown) {
  if (value === undefined || value === null) {
    return '';
  }
  try {
    return JSON.stringify(value, null, 2);
  } catch {
    return String(value);
  }
}

export function EnvModal({
  open,
  envId,
  onCancel,
  onSuccess,
}: EnvModalProps) {
  const [form] = Form.useForm<Partial<StorageEnvRecord> & { config?: string }>();
  const { message } = App.useApp();
  const [loading, setLoading] = useState(false);
  const [editorTheme, setEditorTheme] = useState<'vs-dark' | 'vs'>('vs-dark');

  const schemas = useMemo<FormSchema[]>(
    () => [
      {
        name: 'name',
        label: '环境名称',
        component: 'Input',
        rules: [{ required: true, message: '请输入环境名称' }],
      },
      {
        name: 'code',
        label: '环境编码',
        component: 'Input',
        rules: [{ required: true, message: '请输入环境编码' }],
      },
      {
        name: 'storageType',
        label: '存储类型',
        component: 'Select',
        initialValue: 'local',
        props: {
          options: [
            { label: '本地存储', value: 'local' },
            { label: 'MinIO', value: 'minio' },
            { label: 'S3', value: 's3' },
            { label: '阿里云 OSS', value: 'oss' },
          ],
        },
      },
      {
        name: 'isDefault',
        label: '默认环境',
        component: 'Switch',
        initialValue: false,
      },
      {
        name: 'status',
        label: '状态',
        component: 'Select',
        initialValue: 0,
        props: {
          options: [
            { label: '正常', value: 0 },
            { label: '停用', value: 1 },
          ],
        },
      },
      {
        name: 'config',
        label: '配置信息',
        component: 'MonacoEditor',
        props: {
          height: 320,
          language: 'json',
          theme: editorTheme,
        },
        rules: [
          {
            validator: async (_, value) => {
              if (!value) {
                return;
              }
              JSON.parse(value);
            },
            message: '请输入合法的 JSON',
          },
        ],
      },
      {
        name: 'remark',
        label: '备注',
        component: 'TextArea',
        props: { rows: 3 },
      },
    ],
    [editorTheme],
  );

  useEffect(() => {
    if (!open) {
      form.resetFields();
      return;
    }

    if (!envId) {
      form.resetFields();
      form.setFieldsValue({ storageType: 'local', isDefault: false, status: 0 });
      return;
    }

    void (async () => {
      setLoading(true);
      try {
        const detail = await storageEnvApi.detail(envId);
        form.setFieldsValue({
          ...detail,
          config: safeStringify(detail.config),
        });
      } finally {
        setLoading(false);
      }
    })();
  }, [envId, form, open]);

  const handleSubmit = async () => {
    const values = await form.validateFields();
    setLoading(true);
    try {
      const payload = {
        ...values,
        config: values.config ? JSON.parse(values.config) : undefined,
      };

      if (envId) {
        await storageEnvApi.update(envId, payload);
        message.success('存储环境更新成功');
      } else {
        await storageEnvApi.create(payload);
        message.success('存储环境创建成功');
      }
      onSuccess();
    } finally {
      setLoading(false);
    }
  };

  return (
    <BasicModal
      open={open}
      title={envId ? '编辑存储环境' : '新增存储环境'}
      width={760}
      confirmLoading={loading}
      onCancel={onCancel}
      onOk={() => void handleSubmit()}
    >
      <div style={{ marginBottom: 12 }}>
        <Button onClick={() => setEditorTheme((value) => (value === 'vs-dark' ? 'vs' : 'vs-dark'))}>
          切换编辑器主题
        </Button>
      </div>
      <BasicForm
        form={form}
        schemas={schemas}
        layout="vertical"
        variant="modal"
        showActionButtons={false}
      />
    </BasicModal>
  );
}
