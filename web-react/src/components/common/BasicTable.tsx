import { forwardRef, useCallback, useEffect, useImperativeHandle, useMemo, useState } from 'react';
import type { ReactNode } from 'react';
import { PushpinFilled, PushpinOutlined, ReloadOutlined } from '@ant-design/icons';
import { Button, Card, Form, Space, Table, Tooltip } from 'antd';
import type { TableColumnsType, TableProps } from 'antd';
import { useLocation } from 'react-router-dom';
import type { FormSchema } from '@/types/form';
import type { PageData } from '@/types/api';
import { BasicForm } from './BasicForm';

export interface BasicTableRef<T> {
  reload: () => void;
  reset: () => void;
  getSelectedRows: () => T[];
}

interface BasicTableProps<T extends object> {
  columns: TableColumnsType<T>;
  fetchData: (params: Record<string, unknown>) => Promise<PageData<T>>;
  rowKey: keyof T | ((record: T) => string | number);
  searchSchemas?: FormSchema[];
  toolbar?: ReactNode;
  scroll?: TableProps<T>['scroll'];
  selectable?: boolean;
}

function hasFixedColumns<T extends object>(columns: TableColumnsType<T>): boolean {
  return columns.some((column) => {
    const tableColumn = column as TableColumnsType<T>[number] & {
      children?: TableColumnsType<T>;
      fixed?: 'left' | 'right' | boolean;
    };

    if (tableColumn.fixed) {
      return true;
    }

    return Array.isArray(tableColumn.children) ? hasFixedColumns(tableColumn.children) : false;
  });
}

function InnerBasicTable<T extends object>(
  {
    columns,
    fetchData,
    rowKey,
    searchSchemas = [],
    toolbar,
    scroll,
    selectable = true,
  }: BasicTableProps<T>,
  ref: React.ForwardedRef<BasicTableRef<T>>,
) {
  const location = useLocation();
  const [searchForm] = Form.useForm();
  const [dataSource, setDataSource] = useState<T[]>([]);
  const [loading, setLoading] = useState(false);
  const [pageNum, setPageNum] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [total, setTotal] = useState(0);
  const [searchValues, setSearchValues] = useState<Record<string, unknown>>({});
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
  const fixedColumnStorageKey = `web-react-fixed-columns:${location.pathname}`;
  const [fixedColumnsEnabled, setFixedColumnsEnabled] = useState(() => {
    if (typeof window === 'undefined') {
      return true;
    }

    return window.localStorage.getItem(fixedColumnStorageKey) !== '0';
  });

  useEffect(() => {
    if (typeof window === 'undefined') {
      return;
    }

    setFixedColumnsEnabled(window.localStorage.getItem(fixedColumnStorageKey) !== '0');
  }, [fixedColumnStorageKey]);

  // 统一把“分页 + 查询条件 + 表格刷新”收敛到一个组件里，
  // 这样后续迁移 CRUD 页面时只需要关注列定义和接口调用。
  const loadData = useCallback(async (extraParams?: Record<string, unknown>) => {
    setLoading(true);
    try {
      const result = await fetchData({
        pageNum,
        pageSize,
        ...searchValues,
        ...extraParams,
      });
      setDataSource(result.records ?? []);
      setTotal(result.total ?? 0);
      setSelectedRowKeys([]);
    } finally {
      setLoading(false);
    }
  }, [fetchData, pageNum, pageSize, searchValues]);

  useEffect(() => {
    void loadData();
  }, [loadData]);

  const selectedRows = useMemo(() => {
    const resolveRowKey = (record: T) => {
      if (typeof rowKey === 'function') {
        return rowKey(record);
      }
      return (record as Record<string, unknown>)[rowKey as string] as React.Key;
    };

    return dataSource.filter((record) => selectedRowKeys.includes(resolveRowKey(record)));
  }, [dataSource, rowKey, selectedRowKeys]);

  const supportsFixedColumnToggle = useMemo(() => hasFixedColumns(columns), [columns]);

  useImperativeHandle(
    ref,
    () => ({
      // 暴露给页面层的只有少量必要动作：
      // 重新加载、重置筛选、读取当前勾选行。
      reload: () => void loadData(),
      reset: () => {
        searchForm.resetFields();
        setSearchValues({});
        setPageNum(1);
        void loadData({ pageNum: 1 });
      },
      getSelectedRows: () => selectedRows,
    }),
    [loadData, searchForm, selectedRows],
  );

  return (
    <>
      {searchSchemas.length ? (
        <div className="page-search">
          <BasicForm
            form={searchForm}
            schemas={searchSchemas}
            layout="vertical"
            onReset={() => {
              setSearchValues({});
              setPageNum(1);
              void loadData({ pageNum: 1 });
            }}
            onSubmit={(values) => {
              setSearchValues(values);
              setPageNum(1);
              void loadData({ ...values, pageNum: 1 });
            }}
            resetText="重置"
            submitText="查询"
            variant="search"
          />
        </div>
      ) : null}

      <Card className="page-card data-table-card" variant="borderless">
        <div className="page-toolbar">
          <div className="page-toolbar-main">{toolbar}</div>
          <Space className="page-toolbar-tools" size={8}>
            <Button
              className="table-utility-btn"
              icon={<ReloadOutlined />}
              loading={loading}
              onClick={() => void loadData()}
            >
              刷新
            </Button>
            {supportsFixedColumnToggle ? (
              <Tooltip title={fixedColumnsEnabled ? '取消固定操作列' : '固定操作列'}>
                <Button
                  aria-label={fixedColumnsEnabled ? '取消固定操作列' : '固定操作列'}
                  className="table-fixed-toggle-btn table-toolbar-icon-btn"
                  icon={fixedColumnsEnabled ? <PushpinFilled /> : <PushpinOutlined />}
                  onClick={() => {
                    const nextValue = !fixedColumnsEnabled;
                    setFixedColumnsEnabled(nextValue);

                    if (typeof window !== 'undefined') {
                      window.localStorage.setItem(fixedColumnStorageKey, nextValue ? '1' : '0');
                    }
                  }}
                />
              </Tooltip>
            ) : null}
          </Space>
        </div>

        {selectable && selectedRowKeys.length > 0 ? (
          <div className="table-selection-bar">
            <span>
              已选择 <strong>{selectedRowKeys.length}</strong> 项
            </span>
            <Button size="small" type="link" onClick={() => setSelectedRowKeys([])}>
              清空选择
            </Button>
          </div>
        ) : null}

        <Table<T>
          className={fixedColumnsEnabled ? undefined : 'table-fixed-disabled'}
          columns={columns}
          dataSource={dataSource}
          loading={loading}
          rowKey={rowKey as TableProps<T>['rowKey']}
          rowSelection={
            selectable
              ? {
                  selectedRowKeys,
                  onChange: setSelectedRowKeys,
                }
              : undefined
          }
          scroll={scroll}
          pagination={{
            current: pageNum,
            pageSize,
            total,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (currentTotal) => `共 ${currentTotal} 条`,
            onChange: (nextPage, nextSize) => {
              setPageNum(nextPage);
              setPageSize(nextSize);
            },
          }}
        />
      </Card>
    </>
  );
}

export const BasicTable = forwardRef(InnerBasicTable) as <T extends object>(
  props: BasicTableProps<T> & { ref?: React.Ref<BasicTableRef<T>> },
) => ReturnType<typeof InnerBasicTable>;
