<template>
  <BasicModal
    v-model:visible="visible"
    title="编辑字典子项"
    :width="900"
    :footer="null"
    @cancel="handleCancel"
  >
    <div class="dict-subitems">
      <a-button type="primary" @click="handleAdd" style="margin-bottom: 16px">
        <template #icon><PlusOutlined /></template>
        新增子项
      </a-button>
      
      <BasicTable
        :columns="columns"
        :data-source="subItems"
        :pagination="false"
        :show-action-column="true"
      >
        <template #bodyCell="{ column, record }">
          <template v-if="column.dataIndex === 'isDefault'">
            <a-tag :color="record.isDefault ? 'success' : 'default'">
              {{ record.isDefault ? '是' : '否' }}
            </a-tag>
          </template>
          <template v-else-if="column.dataIndex === 'status'">
            <a-tag :color="record.status === 0 ? 'success' : 'error'">
              {{ record.status === 0 ? '正常' : '停用' }}
            </a-tag>
          </template>
          <template v-else-if="column.dataIndex === 'action'">
            <TableAction
              :actions="[
                {
                  label: '编辑',
                  onClick: () => handleEditItem(record),
                },
                {
                  label: '删除',
                  color: 'error',
                  popConfirm: {
                    title: '确定删除该子项吗？',
                    confirm: () => handleDeleteItem(record.id),
                  },
                },
              ]"
            />
          </template>
        </template>
      </BasicTable>
    </div>

    <DictModal
      v-model:visible="itemModalVisible"
      :id="currentItemId"
      :parent-id="parentId"
      @success="handleItemSuccess"
    />
  </BasicModal>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue';
import { message } from 'ant-design-vue';
import { PlusOutlined } from '@ant-design/icons-vue';
import BasicModal from '@/components/Modal/BasicModal.vue';
import BasicTable from '@/components/Table/BasicTable.vue';
import TableAction from '@/components/Table/TableAction.vue';
import DictModal from './DictModal.vue';
import { dictApi } from '@/api/dict';
import type { Dict } from '@/api/dict';

const props = defineProps<{
  visible: boolean;
  parentId?: number;
}>();

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void;
}>();

const visible = ref(false);
const subItems = ref<Dict[]>([]);
const itemModalVisible = ref(false);
const currentItemId = ref<number>();

const columns = [
  { title: '字典标签', dataIndex: 'dictLabel', width: 150 },
  { title: '字典键值', dataIndex: 'dictValue', width: 150 },
  { title: '排序', dataIndex: 'sort', width: 80 },
  { title: '是否默认', dataIndex: 'isDefault', width: 80 },
  { title: '状态', dataIndex: 'status', width: 80 },
  { title: '操作', dataIndex: 'action', width: 150 },
];

const loadSubItems = async () => {
  if (!props.parentId) return;
  try {
    const parent = await dictApi.detail(props.parentId);
    const items = await dictApi.getByType(parent.dictType, props.parentId);
    subItems.value = items;
  } catch (error) {
    console.error('加载子项失败:', error);
  }
};

const handleAdd = () => {
  currentItemId.value = undefined;
  itemModalVisible.value = true;
};

const handleEditItem = (record: any) => {
  currentItemId.value = record.id;
  itemModalVisible.value = true;
};

const handleDeleteItem = async (id: number) => {
  try {
    await dictApi.delete(id);
    message.success('删除成功');
    loadSubItems();
  } catch (error) {
    console.error('删除失败:', error);
  }
};

const handleItemSuccess = () => {
  itemModalVisible.value = false;
  loadSubItems();
};

const handleCancel = () => {
  emit('update:visible', false);
};

watch(() => props.visible, (val) => {
  visible.value = val;
  if (val) {
    loadSubItems();
  }
});

watch(visible, (val) => {
  emit('update:visible', val);
});
</script>

<style scoped lang="less">
.dict-subitems {
  min-height: 300px;
}
</style>
