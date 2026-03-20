// Code scaffolded by goctl. Safe to edit.
// goctl 1.9.2

package menu

import (
	"net/http"
	"strings"

	"github.com/force-c/nai-tizi/application/sys-api/internal/logic/menu"
	"github.com/force-c/nai-tizi/application/sys-api/internal/svc"
	"github.com/zeromicro/go-zero/rest/httpx"
)

func MenuUserTreeHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get(svcCtx.Config.Auth.TokenHeader), "Bearer "))
		userId, err := menu.ResolveUserID(svcCtx, token)
		if err != nil {
			httpx.OkJsonCtx(r.Context(), w, map[string]interface{}{
				"code": 401,
				"msg":  err.Error(),
			})
			return
		}

		l := menu.NewMenuUserTreeLogic(r.Context(), svcCtx)
		resp, err := l.MenuUserTree(userId)
		if err != nil {
			httpx.ErrorCtx(r.Context(), w, err)
		} else {
			httpx.OkJsonCtx(r.Context(), w, resp)
		}
	}
}
