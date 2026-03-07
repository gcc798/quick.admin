<template>
  <div class="dict-page">
    <BasicTable
      :columns="columns"
      :api="loadData"
      :use-search-form="true"
      :form-config="searchFormConfig"
      :show-action-column="true"
      @register="registerTable"
    >
      <template #toolbar>
        <a-button v-permission="'dict.create'" type="primary" @click="handleCreate">
          <template #icon><PlusOutlined /></template>
            新增
        </a-button>
        <a-button v-permission="'dict.delete'" danger @click="handleBatchDelete">
          <template #icon><DeleteOutlined /></template>
          批量删除
        </a-button>
      </template>

      <template #bodyCell="{ column, record }">
        <template v-if="column.dataIndex === 'status'">
          <a-tag :color="record.status === 0 ? 'success' : 'error'">
            {{ record.status === 0 ? '正常' : '停用' }}
          </a-tag>
        </template>
        <template v-else-if="column.dataIndex === 'isDefault'">
          <a-tag :color="record.isDefault ? 'success' : 'default'">
            {{ record.isDefault ? '是' : '否' }}
          </a-tag>
        </template>
        <template v-else-if="column.dataIndex === 'action'">
          <TableAction
            :actions="[
              {
                label: '编辑',
                onClick: () => handleEdit(record),
                ifShow: hasPermission('dict.update'),
              },
              {
                label: '编辑子项',
                onClick: () => handleEditSubItems(record),
                ifShow: hasPermission('dict.update'),
              },
              {
                label: '删除',
                color: 'error',
                popConfirm: {
                  title: '确定删除该字典吗？',
                  confirm: () => handleDelete(record.id),
                },
                ifShow: hasPermission('dict.delete'),
              },
            ]"
          />
        </template>
      </template>
    </BasicTable>

    <DictModal
      v-model:visible="modalVisible"
      :id="currentDictId"
      @success="handleSuccess"
    />

    <DictSubItemsModal
      v-model:visible="subItemsModalVisible"
      :parent-id="currentParentId"
    />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { message } from 'ant-design-vue';
import { PlusOutlined, DeleteOutlined } from '@ant-design/icons-vue';
import BasicTable from '@/components/Table/BasicTable.vue';
import TableAction from '@/components/Table/TableAction.vue';
import DictModal from './DictModal.vue';
import DictSubItemsModal from './DictSubItemsModal.vue';
import { dictApi } from '@/api/dict';
import { usePermission } from '@/composables/usePermission';
import type { FormConfig } from '@/types/form';

const { hasPermission } = usePermission();

// 表格列配置
const columns = [
  {
    title: '字典ID',
    dataIndex: 'id',
    width: 80,
  },
  {
    title: '字典类型',
    dataIndex: 'dictType',
    width: 150,
  },
  {
    title: '字典标签',
    dataIndex: 'dictLabel',
    width: 150,
  },
  {
    title: '字典键值',
    dataIndex: 'dictValue',
    width: 150,
  },
  {
    title: '排序',
    dataIndex: 'sort',
    width: 80,
  },
  {
    title: '默认',
    dataIndex: 'isDefault',
    width: 80,
  },
  {
    title: '状态',
    dataIndex: 'status',
    width: 80,
  },
  {
    title: '创建时间',
    dataIndex: 'createdTime',
    width: 180,
  },
  {
    title: '操作',
    dataIndex: 'action',
    width: 150,
    fixed: 'right',
  },
];

// 搜索表单配置
const searchFormConfig: FormConfig = {
  schemas: [
    {
      field: 'dictType',
      label: '字典类型',
      component: 'Input',
      colProps: { span: 6 },
    },
    {
      field: 'dictLabel',
      label: '字典标签',
      component: 'Input',
      colProps: { span: 6 },
    },
    {
      field: 'status',
      label: '状态',
      component: 'Select',
      componentProps: {
        options: [
          { label: '正常', value: 0 },
          { label: '停用', value: 1 },
        ],
      },
      colProps: { span: 6 },
    },
  ],
};

// 表格实例
const tableRef = ref();
const registerTable = (methods: any) => {
  tableRef.value = methods;
};

// 加载数据
const loadData = async (params: any) => {
  const res = await dictApi.page(params);
  return {
    records: res.records || [],
    total: res.total || 0,
  };
};

// 弹窗状态
const modalVisible = ref(false);
const currentDictId = ref<number>();
const subItemsModalVisible = ref(false);
const currentParentId = ref<number>();

// 新增字典
const handleCreate = () => {
  currentDictId.value = undefined;
  modalVisible.value = true;
};

// 编辑字典
const handleEdit = (record: any) => {
  currentDictId.value = record.id;
  modalVisible.value = true;
};

// 删除字典
const handleDelete = async (id: number) => {
  try {
    await dictApi.delete(id);
    message.success('删除成功');
    tableRef.value?.reload();
  } catch (error) {
    console.error('删除失败:', error);
  }
};

// 批量删除
const handleBatchDelete = async () => {
  const selectedRows = tableRef.value?.getSelectRows();
  if (!selectedRows || selectedRows.length === 0) {
    message.warning('请选择要删除的字典');
    return;
  }
  try {
    const ids = selectedRows.map((row: any) => row.id);
    await dictApi.batchDelete(ids);
    message.success('批量删除成功');
    tableRef.value?.reload();
  } catch (error) {
    console.error('批量删除失败:', error);
  }
};

// 编辑子项
const handleEditSubItems = (record: any) => {
  currentParentId.value = record.id;
  subItemsModalVisible.value = true;
};

// 操作成功回调
const handleSuccess = () => {
  modalVisible.value = false;
  tableRef.value?.reload();
};
</script>

<style scoped lang="less">
.dict-page {
  padding: 0;
}
</style>
