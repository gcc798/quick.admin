import { useRef, useState } from 'react';
import type { ColumnsType } from 'antd/es/table';
import { App, Button, Space, Tag } from 'antd';
import { PlusOutlined } from '@ant-design/icons';
import { BasicTable, type BasicTableRef } from '@/components/common/BasicTable';
import { JsonViewerModal } from '@/components/common/JsonViewerModal';
import { PermissionGate } from '@/components/common/PermissionGate';
import { TableAction } from '@/components/common/TableAction';
import type { SnowflakeId } from '@/types/api';
import type { FormSchema } from '@/types/form';
import type { StorageEnvRecord } from '@/types/system';
import { storageEnvApi } from '@/api/storageenv';
import { EnvModal } from './EnvModal';

const searchSchemas: FormSchema[] = [
  {
    name: 'name',
    label: '名称',
    component: 'Input',
    colProps: { span: 12 },
  },
  {
    name: 'storageType',
    label: '存储类型',
    component: 'Select',
    colProps: { span: 12 },
    props: {
      allowClear: true,
      options: [
        { label: '本地存储', value: 'local' },
        { label: 'MinIO', value: 'minio' },
        { label: 'S3', value: 's3' },
        { label: '阿里云 OSS', value: 'oss' },
      ],
    },
  },
];

function previewJson(data: unknown) {
  const text =
    typeof data === 'string' ? data : JSON.stringify(data ?? {}, null, 2);
  const singleLine = text.replace(/\s+/g, ' ').trim();
  return singleLine.length > 72 ? `${singleLine.slice(0, 72)}...` : singleLine;
}

export default function StorageEnvPage() {
  const { message } = App.useApp();
  const tableRef = useRef<BasicTableRef<StorageEnvRecord>>(null);
  const [modalOpen, setModalOpen] = useState(false);
  const [currentEnvId, setCurrentEnvId] = useState<SnowflakeId>();
  const [viewerOpen, setViewerOpen] = useState(false);
  const [viewerData, setViewerData] = useState<unknown>(null);

  const columns: ColumnsType<StorageEnvRecord> = [
    { title: '名称', dataIndex: 'name', width: 180 },
    { title: '编码', dataIndex: 'code', width: 160 },
    {
      title: '存储类型',
      dataIndex: 'storageType',
      width: 140,
      render: (value) => {
        const map: Record<string, { color: string; label: string }> = {
          local: { color: 'blue', label: '本地存储' },
          minio: { color: 'green', label: 'MinIO' },
          s3: { color: 'orange', label: 'S3' },
          oss: { color: 'purple', label: '阿里云 OSS' },
        };
        const current = map[value] ?? { color: 'default', label: String(value) };
        return <Tag color={current.color}>{current.label}</Tag>;
      },
    },
    {
      title: '配置',
      dataIndex: 'config',
      width: 340,
      render: (value) => (
        <span
          className="json-preview-inline table-code"
          role="button"
          tabIndex={0}
          onClick={() => {
            setViewerData(value);
            setViewerOpen(true);
          }}
          onKeyDown={(event) => {
            if (event.key === 'Enter' || event.key === ' ') {
              event.preventDefault();
              setViewerData(value);
              setViewerOpen(true);
            }
          }}
        >
          {previewJson(value)}
        </span>
      ),
    },
    {
      title: '默认',
      dataIndex: 'isDefault',
      width: 100,
      render: (value) => <Tag color={value ? 'success' : 'default'}>{value ? '是' : '否'}</Tag>,
    },
    {
      title: '状态',
      dataIndex: 'status',
      width: 100,
      render: (value) => (
        <Tag color={value === 0 ? 'success' : 'error'}>
          {value === 0 ? '正常' : '停用'}
        </Tag>
      ),
    },
    {
      title: '操作',
      key: 'action',
      width: 260,
      fixed: 'right',
      render: (_, record) => (
        <TableAction
          actions={[
            {
              key: 'edit',
              label: '编辑',
              permission: 'storage_env.update',
              onClick: () => {
                setCurrentEnvId(record.id);
                setModalOpen(true);
              },
            },
            {
              key: 'setDefault',
              label: '设为默认',
              permission: 'storage_env.update',
              hidden: record.isDefault === true,
              onClick: async () => {
                await storageEnvApi.setDefault(record.id);
                message.success('设置成功');
                tableRef.current?.reload();
              },
            },
            {
              key: 'test',
              label: '测试连接',
              permission: 'storage_env.update',
              onClick: async () => {
                await storageEnvApi.testConnection(record.id);
                message.success('连接测试成功');
              },
            },
            {
              key: 'delete',
              label: '删除',
              permission: 'storage_env.delete',
              danger: true,
              confirmTitle: '确定删除该存储环境吗？',
              onClick: async () => {
                await storageEnvApi.delete(record.id);
                message.success('删除成功');
                tableRef.current?.reload();
              },
            },
          ]}
        />
      ),
    },
  ];

  return (
    <>
      <BasicTable<StorageEnvRecord>
        ref={tableRef}
        columns={columns}
        fetchData={storageEnvApi.page}
        rowKey="id"
        searchSchemas={searchSchemas}
        scroll={{ x: 1400 }}
        toolbar={
          <Space>
            <PermissionGate permission="storage_env.create">
              <Button
                icon={<PlusOutlined />}
                type="primary"
                onClick={() => {
                  setCurrentEnvId(undefined);
                  setModalOpen(true);
                }}
              >
                新增
              </Button>
            </PermissionGate>
          </Space>
        }
      />

      <EnvModal
        open={modalOpen}
        envId={currentEnvId}
        onCancel={() => setModalOpen(false)}
        onSuccess={() => {
          setModalOpen(false);
          tableRef.current?.reload();
        }}
      />

      <JsonViewerModal
        open={viewerOpen}
        title="查看存储环境配置"
        data={viewerData}
        onCancel={() => setViewerOpen(false)}
      />
    </>
  );
}
