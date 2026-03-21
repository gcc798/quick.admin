import { forwardRef, useCallback, useEffect, useImperativeHandle, useMemo, useState } from 'react';
import type { ReactNode } from 'react';
import { Card, Form, Table } from 'antd';
import type { TableColumnsType, TableProps } from 'antd';
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
  const [searchForm] = Form.useForm();
  const [dataSource, setDataSource] = useState<T[]>([]);
  const [loading, setLoading] = useState(false);
  const [pageNum, setPageNum] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [total, setTotal] = useState(0);
  const [searchValues, setSearchValues] = useState<Record<string, unknown>>({});
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);

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
    <Card variant="borderless">
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
          />
        </div>
      ) : null}

      {toolbar ? <div className="page-toolbar">{toolbar}</div> : null}

      <Table<T>
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
  );
}

export const BasicTable = forwardRef(InnerBasicTable) as <T extends object>(
  props: BasicTableProps<T> & { ref?: React.Ref<BasicTableRef<T>> },
) => ReturnType<typeof InnerBasicTable>;
