package response

// RoleResponse 角色响应
type RoleResponse struct {
	RoleId     int64  `json:"roleId" example:"1"`                       // 角色ID
	RoleKey    string `json:"roleKey" example:"admin"`                  // 角色标识
	RoleName   string `json:"roleName" example:"超级管理员"`                 // 角色名称
	Sort       int64  `json:"sort" example:"1"`                         // 显示顺序
	Status     int32  `json:"status" example:"0"`                       // 状态：0正常 1停用
	DataScope  int32  `json:"dataScope" example:"1"`                    // 数据范围
	IsSystem   bool   `json:"isSystem" example:"true"`                  // 是否系统内置
	Remark     string `json:"remark" example:"超级管理员，拥有所有权限"`            // 备注
	CreateBy   int64  `json:"createBy" example:"1"`                     // 创建人
	CreateTime string `json:"createTime" example:"2024-01-01 12:00:00"` // 创建时间
}

// RolePermissionResponse 角色权限响应
type RolePermissionResponse struct {
	RoleKey  string `json:"roleKey" example:"admin"` // 角色标识
	Resource string `json:"resource" example:"*"`    // 资源路径
	Action   string `json:"action" example:"*"`      // 操作类型
}

// UserRoleResponse 用户角色响应
type UserRoleResponse struct {
	UserId   int64    `json:"userId" example:"1001"`              // 用户ID
	UserName string   `json:"userName" example:"zhangsan"`        // 用户名
	Roles    []string `json:"roles" example:"admin,user_manager"` // 角色列表
}
