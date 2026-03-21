import Editor from '@monaco-editor/react';

interface MonacoEditorProps {
  value?: string;
  onChange?: (value: string) => void;
  language?: string;
  height?: string | number;
  theme?: 'vs-dark' | 'vs' | 'hc-black';
  readonly?: boolean;
}

export function MonacoEditor({
  value,
  onChange,
  language = 'json',
  height = 320,
  theme = 'vs-dark',
  readonly = false,
}: MonacoEditorProps) {
  return (
    <div style={{ border: '1px solid #d9d9d9', borderRadius: 8, overflow: 'hidden' }}>
      <Editor
        height={height}
        language={language}
        theme={theme}
        value={value}
        onChange={(nextValue) => onChange?.(nextValue ?? '')}
        options={{
          minimap: { enabled: false },
          formatOnPaste: true,
          formatOnType: true,
          scrollBeyondLastLine: false,
          automaticLayout: true,
          readOnly: readonly,
          tabSize: 2,
          fontSize: 14,
        }}
      />
    </div>
  );
}
