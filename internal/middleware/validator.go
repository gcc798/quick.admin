package middleware

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/force-c/nai-tizi/internal/domain/response"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

// ValidationError 单个字段验证错误
type ValidationError struct {
	Field   string `json:"field"`   // 字段名（英文）
	Message string `json:"message"` // 错误信息（中文）
}

// ToResponseFieldError 转换为响应层的字段错误类型
func (e ValidationError) ToResponseFieldError() response.ValidationFieldError {
	return response.ValidationFieldError{
		Field:   e.Field,
		Message: e.Message,
	}
}

// ValidationErrorsResponse 验证错误响应
type ValidationErrorsResponse struct {
	Errors []ValidationError `json:"errors"` // 错误列表
}

// fieldNameMap 字段名称映射（英文 -> 中文）
var fieldNameMap = map[string]string{
	"UserName":    "用户名",
	"NickName":    "昵称",
	"Password":    "密码",
	"Email":       "邮箱",
	"Phonenumber": "手机号",
	"Avatar":      "头像",
	"Status":      "状态",
	"Remark":      "备注",
	"RoleKey":     "角色标识",
	"RoleName":    "角色名称",
	"Sort":        "排序",
	"UserType":    "用户类型",
	"Sex":         "性别",
	"OrgName":     "组织名称",
	"OrgCode":     "组织编码",
	"OrgType":     "组织类型",
	"Leader":      "负责人",
	"Phone":       "联系电话",
	"MenuName":    "菜单名称",
	"Path":        "路由地址",
	"Component":   "组件路径",
	"Permission":  "权限标识",
	"Title":       "标题",
	"Content":     "内容",
	"Code":        "编码",
	"Name":        "名称",
	"Type":        "类型",
	"Value":       "值",
	"ParentId":    "父级ID",
	"SortOrder":   "排序",
	"IsSystem":    "是否系统内置",
	"DataScope":   "数据范围",
	"OldPassword": "旧密码",
	"NewPassword": "新密码",
	"ClientId":    "客户端ID",
	"DeviceType":  "设备类型",
}

// tagMessageMap 验证标签错误消息映射
type tagMessageMap map[string]string

// fieldTagMessages 特定字段的验证错误消息（优先级高于通用消息）
var fieldTagMessages = map[string]tagMessageMap{
	"Email": {
		"email": "邮箱格式不正确，请输入有效的邮箱地址",
	},
	"Phonenumber": {
		"len": "手机号必须为11位数字",
	},
	"Password": {
		"min": "密码长度不能少于{0}个字符",
	},
	"UserName": {
		"min": "用户名长度不能少于{0}个字符",
		"max": "用户名长度不能超过{0}个字符",
	},
}

// tagMessage 通用验证标签错误消息
var tagMessage = map[string]string{
	"required": "{0}不能为空",
	"min":      "{0}长度不能少于{1}个字符",
	"max":      "{0}长度不能超过{1}个字符",
	"len":      "{0}长度必须为{1}位",
	"email":    "{0}格式不正确",
	"oneof":    "{0}的值必须是以下之一: {1}",
	"numeric":  "{0}必须是数字",
	"alphanum": "{0}只能包含字母和数字",
	"gt":       "{0}必须大于{1}",
	"gte":      "{0}必须大于或等于{1}",
	"lt":       "{0}必须小于{1}",
	"lte":      "{0}必须小于或等于{1}",
	"url":      "{0}必须是有效的URL地址",
	"ip":       "{0}必须是有效的IP地址",
	"datetime": "{0}必须是有效的日期时间格式",
	"json":     "{0}必须是有效的JSON格式",
}

// getFieldName 获取字段的中文名称
func getFieldName(field string) string {
	if name, ok := fieldNameMap[field]; ok {
		return name
	}
	return field
}

// getErrorMessage 获取字段的错误消息
func getErrorMessage(field, tag, param string) string {
	fieldName := getFieldName(field)

	// 1. 优先检查特定字段+标签的自定义消息
	if fieldMsgs, ok := fieldTagMessages[field]; ok {
		if msg, ok := fieldMsgs[tag]; ok {
			return strings.ReplaceAll(msg, "{0}", param)
		}
	}

	// 2. 检查通用标签消息
	if msgTemplate, ok := tagMessage[tag]; ok {
		msg := strings.ReplaceAll(msgTemplate, "{0}", fieldName)
		if param != "" {
			msg = strings.ReplaceAll(msg, "{1}", param)
		}
		return msg
	}

	// 3. 默认错误消息
	return fmt.Sprintf("%s验证失败(%s)", fieldName, tag)
}

// ParseValidationErrors 解析验证错误为友好的中文提示
func ParseValidationErrors(err error) []ValidationError {
	var errors []ValidationError

	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		for _, e := range validationErrors {
			errors = append(errors, ValidationError{
				Field:   e.Field(),
				Message: getErrorMessage(e.Field(), e.Tag(), e.Param()),
			})
		}
	}

	return errors
}

// ValidationErrorHandler 验证错误处理中间件
// 统一处理请求参数验证错误，返回友好的中文提示
func ValidationErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 检查是否有验证错误
		if len(c.Errors) > 0 {
			for _, err := range c.Errors {
				// 检查是否是验证错误
				if validationErrors, ok := err.Err.(validator.ValidationErrors); ok {
					resp := ValidationErrorsResponse{
						Errors: ParseValidationErrors(validationErrors),
					}
					response.FailCode(c, response.CodeInvalidParam, formatValidationErrors(resp.Errors))
					return
				}
			}
		}
	}
}

// formatValidationErrors 格式化验证错误列表为字符串
func formatValidationErrors(errors []ValidationError) string {
	if len(errors) == 0 {
		return "参数错误"
	}

	if len(errors) == 1 {
		return errors[0].Message
	}

	// 多个错误时，拼接成一条消息
	var msgs []string
	for _, e := range errors {
		msgs = append(msgs, e.Message)
	}
	return strings.Join(msgs, "；")
}

// ShouldBindJSON 绑定 JSON 并验证，返回友好的错误信息
func ShouldBindJSON(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindJSON(obj); err != nil {
		// 如果是验证错误，返回格式化的错误
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors := ParseValidationErrors(validationErrors)
			return fmt.Errorf("%s", formatValidationErrors(errors))
		}
		// 其他绑定错误（如 JSON 格式错误）
		return err
	}
	return nil
}

// ShouldBindQuery 绑定 Query 参数并验证，返回友好的错误信息
func ShouldBindQuery(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindQuery(obj); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors := ParseValidationErrors(validationErrors)
			return fmt.Errorf("%s", formatValidationErrors(errors))
		}
		return err
	}
	return nil
}

// ShouldBindUri 绑定 URI 参数并验证，返回友好的错误信息
func ShouldBindUri(c *gin.Context, obj interface{}) error {
	if err := c.ShouldBindUri(obj); err != nil {
		if validationErrors, ok := err.(validator.ValidationErrors); ok {
			errors := ParseValidationErrors(validationErrors)
			return fmt.Errorf("%s", formatValidationErrors(errors))
		}
		return err
	}
	return nil
}

// RegisterCustomValidators 注册自定义验证器
func RegisterCustomValidators(v *validator.Validate) {
	// 注册自定义字段名获取函数
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		// 优先使用 json tag 中的名称
		name := strings.SplitN(fld.Tag.Get("json"), ",", 1)[0]
		if name == "-" {
			return fld.Name
		}
		if name != "" {
			return name
		}
		return fld.Name
	})
}
