package response

import "github.com/force-c/nai-tizi/internal/utils"

// UserResponse 用户响应
type UserResponse struct {
	UserId      int64           `json:"userId"`
	UserName    string          `json:"userName"`
	NickName    string          `json:"nickName"`
	UserType    int32           `json:"userType"` // 用户类型：0系统用户 1微信用户 2APP用户
	Email       string          `json:"email"`
	Phonenumber string          `json:"phonenumber"`
	Sex         int32           `json:"sex"` // 性别：0男 1女 2未知
	Avatar      string          `json:"avatar"`
	Status      int32           `json:"status"` // 状态：0正常 1停用
	Sort        int64           `json:"sort"`
	LoginIp     string          `json:"loginIp"`
	LoginDate   int64           `json:"loginDate"`
	OpenId      string          `json:"openId"`
	UnionId     string          `json:"unionId"`
	Remark      string          `json:"remark"`
	CreateBy    int64           `json:"createBy"`
	UpdateBy    int64           `json:"updateBy"`
	CreatedAt   utils.LocalTime `json:"createdAt"`
	UpdatedAt   utils.LocalTime `json:"updatedAt"`
}
