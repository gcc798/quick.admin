import type { FormItemProps } from 'antd';
import type { ColProps } from 'antd/es/grid/col';

export type SchemaComponent =
  | 'Input'
  | 'Password'
  | 'InputNumber'
  | 'Select'
  | 'TreeSelect'
  | 'RadioGroup'
  | 'TextArea'
  | 'Switch'
  | 'MonacoEditor'
  | 'IconPicker';

export interface FormSchema {
  name: string;
  label: string;
  component: SchemaComponent;
  rules?: FormItemProps['rules'];
  props?: Record<string, unknown>;
  colProps?: ColProps;
  initialValue?: unknown;
  helpMessage?: string;
  hidden?: boolean | ((values: Record<string, unknown>) => boolean);
}
