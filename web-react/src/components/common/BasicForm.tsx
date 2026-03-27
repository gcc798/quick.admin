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
  variant?: 'default' | 'search';
  showActionButtons?: boolean;
  submitText?: string;
  resetText?: string;
  onSubmit?: (values: Record<string, unknown>) => void;
  onReset?: () => void;
}

function getSearchPlaceholder(schema: FormSchema) {
  if (schema.component === 'Select' || schema.component === 'TreeSelect') {
    return `请选择${schema.label}`;
  }

  return `请输入${schema.label}`;
}

function getSearchFieldStyle(schema: FormSchema, style?: Record<string, unknown>) {
  const defaultWidth =
    schema.component === 'Select' || schema.component === 'TreeSelect' ? 148 : 176;

  return {
    width: defaultWidth,
    ...style,
  };
}

function renderField(schema: FormSchema, variant: 'default' | 'search') {
  const nextProps = { ...(schema.props ?? {}) };

  if (variant === 'search') {
    if (nextProps.placeholder === undefined) {
      nextProps.placeholder = getSearchPlaceholder(schema);
    }

    if (nextProps.allowClear === undefined) {
      nextProps.allowClear = true;
    }

    if (
      schema.component === 'Input'
      || schema.component === 'Password'
      || schema.component === 'InputNumber'
      || schema.component === 'Select'
      || schema.component === 'TreeSelect'
    ) {
      nextProps.style = getSearchFieldStyle(schema, nextProps.style as Record<string, unknown> | undefined);
    }
  }

  switch (schema.component) {
    case 'Password':
      return <Input.Password {...nextProps} />;
    case 'InputNumber':
      return <InputNumber style={{ width: '100%' }} {...nextProps} />;
    case 'Select':
      return <Select {...nextProps} />;
    case 'TreeSelect':
      return <TreeSelect {...nextProps} />;
    case 'RadioGroup':
      return <Radio.Group {...nextProps} />;
    case 'TextArea':
      return <Input.TextArea {...nextProps} />;
    case 'Switch':
      return <Switch {...nextProps} />;
    case 'MonacoEditor':
      return <MonacoEditor {...nextProps} />;
    case 'IconPicker':
      return <IconPicker {...nextProps} />;
    default:
      return <Input {...nextProps} />;
  }
}

function SchemaFormItem({
  schema,
  form,
  variant,
}: {
  schema: FormSchema;
  form: FormInstance;
  variant: 'default' | 'search';
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
  const content = (
    <Form.Item
      className={variant === 'search' ? 'search-form-item' : undefined}
      extra={schema.helpMessage}
      initialValue={schema.initialValue}
      label={variant === 'search' ? undefined : schema.label}
      name={schema.name}
      rules={schema.rules}
      valuePropName={valuePropName}
    >
      {renderField(schema, variant)}
    </Form.Item>
  );

  if (variant === 'search') {
    return (
      <div className="search-form-field">
        <div className="search-form-inline">
          <span className="search-form-label">{schema.label}</span>
          <div className="search-form-control">{content}</div>
        </div>
      </div>
    );
  }

  return (
    <Col {...schema.colProps} span={schema.colProps?.span ?? 24}>
      {content}
    </Col>
  );
}

export function BasicForm({
  form,
  schemas,
  initialValues,
  layout = 'vertical',
  variant = 'default',
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
      className={variant === 'search' ? 'search-form' : undefined}
      form={form}
      initialValues={formInitialValues}
      layout={layout}
      onFinish={handleFinish}
    >
      {variant === 'search' ? (
        <div className="search-form-grid">
          {schemas.map((schema) => (
            <SchemaFormItem key={schema.name} form={form} schema={schema} variant={variant} />
          ))}
          {showActionButtons ? (
            <div className="search-form-actions">
              <Form.Item style={{ marginBottom: 0 }}>
                <Space size={8}>
                  <Button htmlType="submit" type="primary">
                    {submitText}
                  </Button>
                  <Button onClick={handleReset}>{resetText}</Button>
                </Space>
              </Form.Item>
            </div>
          ) : null}
        </div>
      ) : (
        <Row gutter={16}>
          {schemas.map((schema) => (
            <SchemaFormItem key={schema.name} form={form} schema={schema} variant={variant} />
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
      )}
    </Form>
  );
}
