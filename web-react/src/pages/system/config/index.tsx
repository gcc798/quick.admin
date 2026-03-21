import { useRef, useState } from 'react';
import type { ColumnsType } from 'antd/es/table';
import { App, Button, Popconfirm, Space, Tag } from 'antd';
import { DeleteOutlined, PlusOutlined } from '@ant-design/icons';
import { BasicTable, type BasicTableRef } from '@/components/common/BasicTable';
import { JsonViewerModal } from '@/components/common/JsonViewerModal';
import { TableAction } from '@/components/common/TableAction';
import type { FormSchema } from '@/types/form';
import type { ConfigRecord } from '@/types/system';
import { configApi } from '@/api/config';
import { ConfigModal } from './ConfigModal';

const searchSchemas: FormSchema[] = [
  {
    name: 'name',
    label: '配置名称',
    component: 'Input',
    colProps: { span: 12 },
  },
  {
    name: 'code',
    label: '配置编码',
    component: 'Input',
    colProps: { span: 12 },
  },
];

function previewJson(data: unknown) {
  const text =
    typeof data === 'string' ? data : JSON.stringify(data ?? {}, null, 2);
  return text.length > 80 ? `${text.slice(0, 80)}...` : text;
}

export default function ConfigPage() {
  const { message } = App.useApp();
  const tableRef = useRef<BasicTableRef<ConfigRecord>>(null);
  const [modalOpen, setModalOpen] = useState(false);
  const [currentConfigId, setCurrentConfigId] = useState<number>();
  const [viewerOpen, setViewerOpen] = useState(false);
  const [viewerData, setViewerData] = useState<unknown>(null);

  const columns: ColumnsType<ConfigRecord> = [
    { title: '名称', dataIndex: 'name', width: 180 },
    {
      title: '编码',
      dataIndex: 'code',
      width: 180,
      render: (value) => <Tag color="blue">{value}</Tag>,
    },
    {
      title: '内容',
      dataIndex: 'data',
      width: 320,
      render: (value) => (
        <a
          onClick={() => {
            setViewerData(value);
            setViewerOpen(true);
          }}
        >
          <span className="table-code">{previewJson(value)}</span>
        </a>
      ),
    },
    { title: '备注', dataIndex: 'remark', width: 220 },
    { title: '创建时间', dataIndex: 'createdTime', width: 180 },
    {
      title: '操作',
      key: 'action',
      width: 180,
      fixed: 'right',
      render: (_, record) => (
        <TableAction
          actions={[
            {
              key: 'edit',
              label: '编辑',
              onClick: () => {
                setCurrentConfigId(record.id);
                setModalOpen(true);
              },
            },
            {
              key: 'delete',
              label: '删除',
              danger: true,
              confirmTitle: '确定删除该配置吗？',
              onClick: async () => {
                await configApi.delete(record.id);
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
      <BasicTable<ConfigRecord>
        ref={tableRef}
        columns={columns}
        fetchData={configApi.page}
        rowKey="id"
        searchSchemas={searchSchemas}
        scroll={{ x: 1300 }}
        toolbar={
          <Space>
            <Button
              icon={<PlusOutlined />}
              type="primary"
              onClick={() => {
                setCurrentConfigId(undefined);
                setModalOpen(true);
              }}
            >
              新增
            </Button>
            <Popconfirm
              title="确定批量删除选中的配置吗？"
              onConfirm={async () => {
                const rows = tableRef.current?.getSelectedRows() ?? [];
                if (!rows.length) {
                  message.warning('请先选择配置');
                  return;
                }
                await configApi.batchDelete(rows.map((row) => row.id));
                message.success('批量删除成功');
                tableRef.current?.reload();
              }}
            >
              <Button danger icon={<DeleteOutlined />}>
                批量删除
              </Button>
            </Popconfirm>
          </Space>
        }
      />

      <ConfigModal
        open={modalOpen}
        configId={currentConfigId}
        onCancel={() => setModalOpen(false)}
        onSuccess={() => {
          setModalOpen(false);
          tableRef.current?.reload();
        }}
      />

      <JsonViewerModal
        open={viewerOpen}
        title="查看配置数据"
        data={viewerData}
        onCancel={() => setViewerOpen(false)}
      />
    </>
  );
}
