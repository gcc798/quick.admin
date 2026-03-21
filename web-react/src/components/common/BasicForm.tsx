import { useMemo } from 'react';
import type { ReactNode } from 'react';
import { Button, Col, Form, Input, InputNumber, Radio, Row, Select, Space, Switch, TreeSelect } from 'antd';
import type { FormInstance } from 'antd';
import type { FormSchema } from '@/types/form';
import { MonacoEditor } from './MonacoEditor';
import { IconPicker } from './IconPicker';

interface BasicFormProps {
  form: FormInstance;
  schemas: FormSchema[];
  initialValues?: Record<string, unknown>;
  layout?: 'horizontal' | 'vertical' | 'inline';
  showActionButtons?: boolean;
  submitText?: string;
  resetText?: string;
  onSubmit?: (values: Record<string, unknown>) => void;
  onReset?: () => void;
}

function renderField(schema: FormSchema) {
  switch (schema.component) {
    case 'Password':
      return <Input.Password {...schema.props} />;
    case 'InputNumber':
      return <InputNumber style={{ width: '100%' }} {...schema.props} />;
    case 'Select':
      return <Select {...schema.props} />;
    case 'TreeSelect':
      return <TreeSelect {...schema.props} />;
    case 'RadioGroup':
      return <Radio.Group {...schema.props} />;
    case 'TextArea':
      return <Input.TextArea {...schema.props} />;
    case 'Switch':
      return <Switch {...schema.props} />;
    case 'MonacoEditor':
      return <MonacoEditor {...schema.props} />;
    case 'IconPicker':
      return <IconPicker {...schema.props} />;
    default:
      return <Input {...schema.props} />;
  }
}

function SchemaFormItem({
  schema,
  form,
}: {
  schema: FormSchema;
  form: FormInstance;
}) {
  const values = Form.useWatch([], form) ?? {};
  const hidden =
    typeof schema.hidden === 'function'
      ? schema.hidden(values)
      : schema.hidden;

  if (hidden) {
    return null;
  }

  const valuePropName = schema.component === 'Switch' ? 'checked' : 'value';

  return (
    <Col {...schema.colProps} span={schema.colProps?.span ?? 24}>
      <Form.Item
        extra={schema.helpMessage}
        initialValue={schema.initialValue}
        label={schema.label}
        name={schema.name}
        rules={schema.rules}
        valuePropName={valuePropName}
      >
        {renderField(schema)}
      </Form.Item>
    </Col>
  );
}

export function BasicForm({
  form,
  schemas,
  initialValues,
  layout = 'vertical',
  showActionButtons = true,
  submitText = '提交',
  resetText = '重置',
  onSubmit,
  onReset,
}: BasicFormProps) {
  const formInitialValues = useMemo(() => {
    const defaults = schemas.reduce<Record<string, unknown>>((acc, schema) => {
      if (schema.initialValue !== undefined) {
        acc[schema.name] = schema.initialValue;
      }
      return acc;
    }, {});

    return { ...defaults, ...initialValues };
  }, [initialValues, schemas]);

  const handleFinish = (values: Record<string, unknown>) => {
    onSubmit?.(values);
  };

  const handleReset = () => {
    form.resetFields();
    onReset?.();
  };

  return (
    <Form
      form={form}
      initialValues={formInitialValues}
      layout={layout}
      onFinish={handleFinish}
    >
      <Row gutter={16}>
        {schemas.map((schema) => (
          <SchemaFormItem key={schema.name} form={form} schema={schema} />
        ))}
        {showActionButtons ? (
          <Col span={24}>
            <Form.Item style={{ marginBottom: 0 }}>
              <Space>
                <Button htmlType="submit" type="primary">
                  {submitText}
                </Button>
                <Button onClick={handleReset}>{resetText}</Button>
              </Space>
            </Form.Item>
          </Col>
        ) : null}
      </Row>
    </Form>
  );
}
