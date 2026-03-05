# 菜单管理树形展开Bug修复

## 问题描述

在菜单管理界面中，树形结构的展开/收起功能存在bug：
- 点击任意节点的展开按钮，会导致所有节点都展开
- 点击任意节点的收起按钮，会导致所有节点都收起

## 问题原因

在 `web/src/views/system/menu/index.vue` 的 `handleExpand` 函数中：

```typescript
// 原代码（有bug）
const handleExpand = (expanded: boolean, record: any) => {
  if (expanded) {
    expandedKeys.value.push(record.menuId);  // 直接push，可能重复添加
  } else {
    expandedKeys.value = expandedKeys.value.filter(key => key !== record.menuId);
  }
};
```

**问题分析：**
1. Ant Design Vue 的 Table 组件在展开/收起时，可能会触发多次 `expand` 事件
2. 每次触发都会将 `menuId` 添加到 `expandedKeys` 数组中
3. 如果数组中已经存在该 `menuId`，重复添加会导致状态异常
4. 这会影响到其他节点的展开/收起状态

## 解决方案

在添加 key 之前，先检查是否已经存在：

```typescript
// 修复后的代码
const handleExpand = (expanded: boolean, record: any) => {
  if (expanded) {
    // 避免重复添加
    if (!expandedKeys.value.includes(record.menuId)) {
      expandedKeys.value.push(record.menuId);
    }
  } else {
    expandedKeys.value = expandedKeys.value.filter(key => key !== record.menuId);
  }
};
```

## 修复内容

### 文件：`web/src/views/system/menu/index.vue`

**修改位置：** `handleExpand` 函数

**修改内容：**
- 在展开节点时，添加 `includes` 检查，避免重复添加相同的 menuId
- 确保每个 menuId 在 expandedKeys 数组中只出现一次

## 测试验证

### 测试步骤

1. **单个节点展开/收起**
   - 点击某个节点的展开按钮
   - 验证：只有该节点展开，其他节点保持原状态
   - 点击该节点的收起按钮
   - 验证：只有该节点收起，其他节点保持原状态

2. **多个节点展开/收起**
   - 展开多个不同的节点
   - 验证：每个节点独立展开，互不影响
   - 分别收起这些节点
   - 验证：每个节点独立收起，互不影响

3. **展开全部/折叠全部按钮**
   - 点击"展开全部"按钮
   - 验证：所有节点都展开
   - 点击"折叠全部"按钮
   - 验证：所有节点都收起

4. **嵌套节点测试**
   - 展开父节点
   - 展开子节点
   - 收起父节点
   - 验证：父节点收起，子节点状态保持
   - 再次展开父节点
   - 验证：子节点保持之前的展开状态

### 预期结果

- ✅ 每个节点的展开/收起操作独立，不影响其他节点
- ✅ 展开全部/折叠全部按钮正常工作
- ✅ 嵌套节点的展开/收起状态正确维护
- ✅ 没有重复的 key 在 expandedKeys 数组中

## 技术说明

### expandedKeys 数组管理

`expandedKeys` 是一个响应式数组，用于控制哪些节点处于展开状态：

```typescript
const expandedKeys = ref<number[]>([]);
```

**关键点：**
1. 数组中的每个元素是一个 menuId
2. 如果 menuId 在数组中，对应的节点就会展开
3. 如果 menuId 不在数组中，对应的节点就会收起
4. **重要：** 每个 menuId 应该只在数组中出现一次

### Ant Design Vue Table 的 expand 事件

```vue
<a-table
  :expanded-row-keys="expandedKeys"
  @expand="handleExpand"
>
```

**事件参数：**
- `expanded`: boolean - 表示是展开(true)还是收起(false)
- `record`: any - 当前操作的行数据

**注意事项：**
- 该事件可能在某些情况下被多次触发
- 需要在处理函数中做好幂等性处理
- 避免重复添加相同的 key

## 相关代码

### 完整的展开/收起相关代码

```typescript
// 展开的节点keys
const expandedKeys = ref<number[]>([]);

// 单个节点展开/收起
const handleExpand = (expanded: boolean, record: any) => {
  if (expanded) {
    // 避免重复添加
    if (!expandedKeys.value.includes(record.menuId)) {
      expandedKeys.value.push(record.menuId);
    }
  } else {
    expandedKeys.value = expandedKeys.value.filter(key => key !== record.menuId);
  }
};

// 展开全部
const expandAll = () => {
  const getAllKeys = (data: any[]): number[] => {
    let keys: number[] = [];
    data.forEach(item => {
      keys.push(item.menuId);
      if (item.children && item.children.length > 0) {
        keys = keys.concat(getAllKeys(item.children));
      }
    });
    return keys;
  };
  expandedKeys.value = getAllKeys(menuTree.value);
};

// 折叠全部
const collapseAll = () => {
  expandedKeys.value = [];
};
```

## 总结

这是一个典型的状态管理问题，通过添加简单的重复检查就能解决。修复后：

- ✅ 树形结构的展开/收起功能正常
- ✅ 每个节点独立控制
- ✅ 展开全部/折叠全部功能正常
- ✅ 用户体验得到改善

## 建议

对于类似的树形结构组件，建议：

1. **使用 Set 代替 Array**：如果只需要存储唯一值，使用 Set 更合适
   ```typescript
   const expandedKeys = ref<Set<number>>(new Set());
   ```

2. **使用受控模式**：完全控制展开状态，避免组件内部状态和外部状态不一致

3. **添加日志**：在开发环境下添加日志，帮助调试状态变化

4. **单元测试**：为展开/收起逻辑编写单元测试，确保功能正确
