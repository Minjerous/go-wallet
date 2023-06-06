package router

import (
	"github.com/gin-gonic/gin"
	"main/app/api"
)

type UserRouter struct{}

func (r *UserRouter) InitUserSignRouter(router *gin.RouterGroup) (R gin.IRoutes) {
	userApi := api.User()
	{
		// 注册账号
		router.POST("/register/email", userApi.Sign().SendRegisterEmail)
		router.POST("/register", userApi.Sign().DoRegister)

		// 登录
		router.POST("/login", userApi.Sign().Login)

		// 重设密码
		router.POST("/password/reset/email", userApi.Sign().ResetPassword)
		router.POST("/password/reset", userApi.Sign().DoResetPassword)
	}

	return router
}

func (r *UserRouter) InitUserInfoRouter(router *gin.RouterGroup) (R gin.IRoutes) {
	userApi := api.User()
	{
		router.POST("/avatar", userApi.Info().UpdateAvatar)
	}

	return router
}
