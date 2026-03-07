<template>
  <div class="user-page">
    <BasicTable
      :columns="columns"
      :api="loadData"
      :use-search-form="true"
      :form-config="searchFormConfig"
      :show-action-column="true"
      @register="registerTable"
    >
      <template #toolbar>
        <a-space>
          <a-button v-permission="'user.create'" type="primary" size="middle" @click="handleCreate">
            <template #icon><PlusOutlined /></template>
            新增
          </a-button>
          <a-button v-permission="'user.delete'" danger size="middle" @click="handleBatchDelete">
            <template #icon><DeleteOutlined /></template>
            批量删除
          </a-button>
        </a-space>
      </template>

      <template #bodyCell="{ column, record }">
        <template v-if="column.dataIndex === 'status'">
          <a-tag :color="record.status === 0 ? 'success' : 'error'">
            {{ record.status === 0 ? '正常' : '停用' }}
          </a-tag>
        </template>
        <template v-else-if="column.dataIndex === 'action'">
          <a-space>
            <a-tooltip title="编辑">
              <a-button 
                v-permission="'user.update'"
                type="link" 
                size="small"
                @click="handleEdit(record)"
              >
                <template #icon><EditOutlined /></template>
              </a-button>
            </a-tooltip>
            <a-tooltip title="删除">
              <a-popconfirm
                title="确定删除该用户吗？"
                @confirm="handleDelete(record.userId)"
              >
                <a-button 
                  v-permission="'user.delete'"
                  type="link" 
                  size="small"
                  danger
                >
                  <template #icon><DeleteOutlined /></template>
                </a-button>
              </a-popconfirm>
            </a-tooltip>
            <a-tooltip title="重置密码">
              <a-button 
                v-permission="'user.update'"
                type="link" 
                size="small"
                @click="handleResetPassword(record)"
              >
                <template #icon><KeyOutlined /></template>
              </a-button>
            </a-tooltip>
          </a-space>
        </template>
      </template>
    </BasicTable>

    <UserModal
      v-model:visible="modalVisible"
      :user-id="currentUserId"
      @success="handleSuccess"
    />
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue';
import { message } from 'ant-design-vue';
import { PlusOutlined, DeleteOutlined, EditOutlined, KeyOutlined } from '@ant-design/icons-vue';
import BasicTable from '@/components/Table/BasicTable.vue';
import TableAction from '@/components/Table/TableAction.vue';
import UserModal from './UserModal.vue';
import { userApi } from '@/api/user';
import { usePermission } from '@/composables/usePermission';

const { hasPermission } = usePermission();

// 表格列配置
const columns = [
  {
    title: '用户ID',
    dataIndex: 'id',
    width: 80,
  },
  {
    title: '用户名',
    dataIndex: 'userName',
    width: 120,
  },
  {
    title: '昵称',
    dataIndex: 'nickName',
    width: 120,
  },
  {
    title: '用户类型',
    dataIndex: 'userType',
    width: 120,
  },
  {
    title: '邮箱',
    dataIndex: 'email',
    width: 180,
  },
  {
    title: '手机号',
    dataIndex: 'phonenumber',
    width: 120,
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
const searchFormConfig = {
  schemas: [
    {
      field: 'username',
      label: '用户名',
      component: 'Input',
      colProps: { span: 6 },
    },
    {
      field: 'phonenumber',
      label: '手机号',
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
  return await userApi.list(params);
};

// 弹窗状态
const modalVisible = ref(false);
const currentUserId = ref<number>();

// 新增用户
const handleCreate = () => {
  currentUserId.value = undefined;
  modalVisible.value = true;
};

// 编辑用户
const handleEdit = (record: any) => {
  currentUserId.value = record.userId;
  modalVisible.value = true;
};

// 删除用户
const handleDelete = async (userId: number) => {
  try {
    await userApi.delete(userId);
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
    message.warning('请选择要删除的用户');
    return;
  }
  try {
    const userIds = selectedRows.map((row: any) => row.userId);
    await userApi.batchDelete(userIds);
    message.success('批量删除成功');
    tableRef.value?.reload();
  } catch (error) {
    console.error('批量删除失败:', error);
  }
};

// 重置密码
const handleResetPassword = async (record: any) => {
  try {
    await userApi.resetPassword(record.userId, '123456');
    message.success('密码已重置为默认密码');
  } catch (error) {
    console.error('重置密码失败:', error);
  }
};

// 操作成功回调
const handleSuccess = () => {
  modalVisible.value = false;
  tableRef.value?.reload();
};
</script>

<style scoped lang="less">
.user-page {
  background: #fff;
  padding: 24px;
  border-radius: 4px;
}
</style>
