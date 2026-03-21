import { BarChartOutlined, DatabaseOutlined, SafetyOutlined, ThunderboltOutlined } from '@ant-design/icons';
import { Card, Col, List, Row, Space, Statistic, Typography } from 'antd';

const metrics = [
  {
    title: '系统模块',
    value: '11',
    description: '当前已确认需要迁移的核心后台模块数量',
    icon: <DatabaseOutlined />,
  },
  {
    title: '动态菜单',
    value: '已接入',
    description: '基于后端菜单树驱动 React 路由与侧边栏',
    icon: <SafetyOutlined />,
  },
  {
    title: '权限控制',
    value: '已接入',
    description: '按钮权限与菜单权限统一走权限仓库',
    icon: <ThunderboltOutlined />,
  },
  {
    title: '迁移节奏',
    value: '进行中',
    description: '当前优先完成登录、基础布局、用户、角色、菜单',
    icon: <BarChartOutlined />,
  },
];

const todoList = [
  '补齐组织、字典、配置、存储环境、文件、日志页面的完整业务逻辑',
  '补充 Monaco 编辑器、文件预览、复杂树结构页面的交互细节',
  '联调真实接口返回值，按后端实际结构收敛页面实体类型',
];

export default function DashboardPage() {
  return (
    <Space direction="vertical" size={20} style={{ display: 'flex' }}>
      <Card className="page-card" variant="borderless">
        <Space direction="vertical" size={6}>
          <Typography.Title level={3} style={{ margin: 0 }}>
            React 重写进度面板
          </Typography.Title>
          <Typography.Text type="secondary">
            这里不是最终业务仪表盘，而是当前重写工程的概览页，用来展示基础能力已经落地的部分。
          </Typography.Text>
        </Space>
      </Card>

      <Row gutter={[16, 16]}>
        {metrics.map((item) => (
          <Col key={item.title} lg={6} md={12} sm={24} xs={24}>
            <Card className="page-card" variant="borderless">
              <Statistic
                prefix={item.icon}
                title={item.title}
                value={item.value}
              />
              <Typography.Paragraph type="secondary" style={{ marginTop: 12, marginBottom: 0 }}>
                {item.description}
              </Typography.Paragraph>
            </Card>
          </Col>
        ))}
      </Row>

      <Card className="page-card" title="下一阶段工作" variant="borderless">
        <List
          dataSource={todoList}
          renderItem={(item) => <List.Item>{item}</List.Item>}
        />
      </Card>
    </Space>
  );
}
