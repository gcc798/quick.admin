<template>
  <BasicModal
    v-model:visible="visible"
    :title="isEdit ? '编辑字典' : '新增字典'"
    :width="600"
    :confirm-loading="loading"
    @ok="handleSubmit"
    @cancel="handleCancel"
  >
    <BasicForm
      ref="formRef"
      :schemas="formSchemas"
      :model="formData"
      :label-width="100"
      :show-action-buttons="false"
    />
  </BasicModal>
</template>

<script setup lang="ts">
import { ref, computed, watch } from 'vue';
import { message } from 'ant-design-vue';
import BasicModal from '@/components/Modal/BasicModal.vue';
import BasicForm from '@/components/Form/BasicForm.vue';
import { dictApi } from '@/api/dict';
import type { FormSchema } from '@/types/form';

const props = defineProps<{
  visible: boolean;
  id?: number;
  parentId?: number;
}>();

const emit = defineEmits<{
  (e: 'update:visible', value: boolean): void;
  (e: 'success'): void;
}>();

const visible = computed({
  get: () => props.visible,
  set: (val) => emit('update:visible', val),
});

const isEdit = computed(() => !!props.id);
const isSubItem = computed(() => !!props.parentId);
const loading = ref(false);
const formRef = ref();
const formData = ref<Record<string, any>>({});
const parentDictType = ref<string>('');

// 表单配置
const formSchemas = computed(() => {
  const schemas: FormSchema[] = [
    ...(!isSubItem.value ? [{
      field: 'dictType',
      label: '字典类型',
      component: 'Input',
      rules: [{ required: true, message: '请输入字典类型' }],
    } as FormSchema] : []),
  {
    field: 'dictLabel',
    label: '字典标签',
    component: 'Input',
    rules: [{ required: true, message: '请输入字典标签' }],
  },
  {
    field: 'dictValue',
    label: '字典键值',
    component: 'Input',
    rules: [{ required: true, message: '请输入字典键值' }],
  },
  {
    field: 'dictSort',
    label: '字典排序',
    component: 'InputNumber',
    defaultValue: 0,
    componentProps: {
      min: 0,
    },
  },
  ...(isSubItem.value ? [{
    field: 'isDefault',
    label: '是否默认',
    component: 'Switch',
    defaultValue: false,
  } as FormSchema] : []),
  {
    field: 'status',
    label: '状态',
    component: 'Select',
    defaultValue: 0,
    componentProps: {
      options: [
        { label: '正常', value: 0 },
        { label: '停用', value: 1 },
      ],
    },
  },
    {
      field: 'remark',
      label: '备注',
      component: 'Textarea',
      componentProps: {
        rows: 3,
      },
    },
  ];
  return schemas;
});

// 加载字典详情
const loadDictDetail = async () => {
  if (!props.id) return;
  
  try {
    loading.value = true;
    const data = await dictApi.detail(props.id);
    formData.value = data;
    formRef.value?.setFieldsValue(data);
  } catch (error) {
    console.error('加载字典详情失败:', error);
  } finally {
    loading.value = false;
  }
};

const handleSubmit = async () => {
  try {
    await formRef.value?.validate();
    loading.value = true;
    const values = formRef.value?.getFieldsValue();
    const data = { ...values };
    if (props.parentId) {
      data.parentId = props.parentId;
      data.dictType = parentDictType.value;
    }
    if (isEdit.value) {
      await dictApi.update({ ...data, id: props.id });
      message.success('更新成功');
    } else {
      await dictApi.create(data);
      message.success('创建成功');
    }
    emit('success');
  } catch (error) {
    console.error('提交失败:', error);
  } finally {
    loading.value = false;
  }
};

// 取消
const handleCancel = () => {
  formRef.value?.resetFields();
  formData.value = {};
};

// 加载父字典类型
const loadParentDictType = async () => {
  if (!props.parentId) return;
  try {
    const parent = await dictApi.detail(props.parentId);
    parentDictType.value = parent.dictType;
  } catch (error) {
    console.error('加载父字典类型失败:', error);
  }
};

// 监听弹窗显示
watch(
  () => props.visible,
  async (val) => {
    if (val) {
      if (props.parentId) {
        await loadParentDictType();
      }
      if (props.id) {
        loadDictDetail();
      } else {
        formRef.value?.resetFields();
        formData.value = {};
      }
    }
  }
);
</script>
