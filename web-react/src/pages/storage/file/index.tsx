import { useRef, useState } from 'react';
import type { ColumnsType } from 'antd/es/table';
import { App, Button, Modal, Popconfirm, Space, Tag, Upload } from 'antd';
import { DeleteOutlined, UploadOutlined } from '@ant-design/icons';
import { BasicTable, type BasicTableRef } from '@/components/common/BasicTable';
import { PermissionGate } from '@/components/common/PermissionGate';
import { TableAction } from '@/components/common/TableAction';
import type { FormSchema } from '@/types/form';
import type { AttachmentRecord } from '@/types/system';
import { attachmentApi } from '@/api/attachment';

interface UploadRequestLike {
  file: File | { originFileObj?: File };
  onSuccess?: (body: unknown, file?: File) => void;
  onError?: (error: Error) => void;
}

const searchSchemas: FormSchema[] = [
  {
    name: 'fileName',
    label: '文件名',
    component: 'Input',
    colProps: { span: 8 },
  },
  {
    name: 'fileType',
    label: '文件类型',
    component: 'Input',
    colProps: { span: 8 },
  },
  {
    name: 'businessType',
    label: '业务类型',
    component: 'Input',
    colProps: { span: 8 },
  },
];

function formatFileSize(size?: number) {
  if (!size) {
    return '0 B';
  }
  if (size < 1024) return `${size} B`;
  if (size < 1024 * 1024) return `${(size / 1024).toFixed(2)} KB`;
  if (size < 1024 * 1024 * 1024) return `${(size / 1024 / 1024).toFixed(2)} MB`;
  return `${(size / 1024 / 1024 / 1024).toFixed(2)} GB`;
}

function isPreviewable(fileType?: string) {
  return Boolean(fileType?.startsWith('image/') || fileType === 'application/pdf');
}

export default function FilePage() {
  const { message } = App.useApp();
  const tableRef = useRef<BasicTableRef<AttachmentRecord>>(null);
  const [previewOpen, setPreviewOpen] = useState(false);
  const [previewRecord, setPreviewRecord] = useState<AttachmentRecord>();
  const [previewUrl, setPreviewUrl] = useState('');

  const columns: ColumnsType<AttachmentRecord> = [
    {
      title: '文件名',
      dataIndex: 'fileName',
      width: 240,
      render: (value, record) => (
        <a
          onClick={async () => {
            if (!isPreviewable(record.fileType)) {
              return;
            }
            const data = await attachmentApi.getUrl(record.attachmentId);
            setPreviewRecord(record);
            setPreviewUrl(data.url);
            setPreviewOpen(true);
          }}
        >
          {value}
        </a>
      ),
    },
    {
      title: '文件大小',
      dataIndex: 'fileSize',
      width: 120,
      render: (value) => formatFileSize(value),
    },
    {
      title: '文件类型',
      dataIndex: 'fileType',
      width: 180,
      render: (value) => <Tag>{value}</Tag>,
    },
    { title: '业务类型', dataIndex: 'businessType', width: 160 },
    { title: '上传人', dataIndex: 'uploadBy', width: 120 },
    { title: '上传时间', dataIndex: 'uploadTime', width: 180 },
    {
      title: '操作',
      key: 'action',
      width: 220,
      fixed: 'right',
      render: (_, record) => (
        <TableAction
          actions={[
            {
              key: 'preview',
              label: '预览',
              hidden: !isPreviewable(record.fileType),
              onClick: async () => {
                const data = await attachmentApi.getUrl(record.attachmentId);
                setPreviewRecord(record);
                setPreviewUrl(data.url);
                setPreviewOpen(true);
              },
            },
            {
              key: 'download',
              label: '下载',
              permission: 'attachment.download',
              onClick: () => attachmentApi.download(record.attachmentId),
            },
            {
              key: 'delete',
              label: '删除',
              permission: 'attachment.delete',
              danger: true,
              confirmTitle: '确定删除该文件吗？',
              onClick: async () => {
                await attachmentApi.delete(record.attachmentId);
                message.success('删除成功');
                tableRef.current?.reload();
              },
            },
          ]}
        />
      ),
    },
  ];

  const handleUpload = async (options: UploadRequestLike) => {
    const file =
      options.file instanceof File
        ? options.file
        : options.file.originFileObj;

    if (!file) {
      options.onError?.(new Error('无法读取上传文件'));
      return;
    }
    try {
      await attachmentApi.uploadFile(file);
      message.success('上传成功');
      options.onSuccess?.({}, file);
      tableRef.current?.reload();
    } catch (error) {
      message.error('上传失败');
      options.onError?.(error as Error);
    }
  };

  return (
    <>
      <BasicTable<AttachmentRecord>
        ref={tableRef}
        columns={columns}
        fetchData={attachmentApi.page}
        rowKey="attachmentId"
        searchSchemas={searchSchemas}
        scroll={{ x: 1300 }}
        toolbar={
          <Space>
            <PermissionGate permission="attachment.upload">
              <Upload
                customRequest={(options) =>
                  void handleUpload(options as unknown as UploadRequestLike)
                }
                multiple
                showUploadList={false}
              >
                <Button icon={<UploadOutlined />} type="primary">
                  上传文件
                </Button>
              </Upload>
            </PermissionGate>
            <PermissionGate permission="attachment.delete">
              <Popconfirm
                title="确定批量删除选中的文件吗？"
                onConfirm={async () => {
                  const rows = tableRef.current?.getSelectedRows() ?? [];
                  if (!rows.length) {
                    message.warning('请先选择文件');
                    return;
                  }
                  await attachmentApi.batchDelete(rows.map((row) => row.attachmentId));
                  message.success('批量删除成功');
                  tableRef.current?.reload();
                }}
              >
                <Button danger icon={<DeleteOutlined />}>
                  批量删除
                </Button>
              </Popconfirm>
            </PermissionGate>
          </Space>
        }
      />

      <Modal
        destroyOnClose
        footer={null}
        open={previewOpen}
        title="文件预览"
        width={900}
        onCancel={() => setPreviewOpen(false)}
      >
        {previewRecord?.fileType?.startsWith('image/') ? (
          <img alt={previewRecord.fileName} src={previewUrl} style={{ width: '100%' }} />
        ) : previewRecord?.fileType === 'application/pdf' ? (
          <iframe src={previewUrl} style={{ width: '100%', height: 640, border: 'none' }} />
        ) : (
          <Space direction="vertical">
            <span>该文件类型暂不支持在线预览。</span>
            <Button onClick={() => previewRecord && attachmentApi.download(previewRecord.attachmentId)}>
              下载文件
            </Button>
          </Space>
        )}
      </Modal>
    </>
  );
}
