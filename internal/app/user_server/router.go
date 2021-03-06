// Copyright 2019-2020 Axetroy. All rights reserved. MIT license.
package user_server

import (
	"fmt"
	"github.com/axetroy/go-server/internal/app/user_server/controller/address"
	"github.com/axetroy/go-server/internal/app/user_server/controller/area"
	"github.com/axetroy/go-server/internal/app/user_server/controller/auth"
	"github.com/axetroy/go-server/internal/app/user_server/controller/banner"
	"github.com/axetroy/go-server/internal/app/user_server/controller/email"
	"github.com/axetroy/go-server/internal/app/user_server/controller/finance"
	"github.com/axetroy/go-server/internal/app/user_server/controller/help"
	"github.com/axetroy/go-server/internal/app/user_server/controller/invite"
	"github.com/axetroy/go-server/internal/app/user_server/controller/message"
	"github.com/axetroy/go-server/internal/app/user_server/controller/news"
	"github.com/axetroy/go-server/internal/app/user_server/controller/notification"
	"github.com/axetroy/go-server/internal/app/user_server/controller/oauth2"
	"github.com/axetroy/go-server/internal/app/user_server/controller/report"
	"github.com/axetroy/go-server/internal/app/user_server/controller/signature"
	"github.com/axetroy/go-server/internal/app/user_server/controller/transfer"
	"github.com/axetroy/go-server/internal/app/user_server/controller/user"
	"github.com/axetroy/go-server/internal/app/user_server/controller/wallet"
	"github.com/axetroy/go-server/internal/library/config"
	"github.com/axetroy/go-server/internal/library/router"
	"github.com/axetroy/go-server/internal/middleware"
	"github.com/axetroy/go-server/internal/rbac/accession"
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/middleware/logger"
	"github.com/kataras/iris/v12/middleware/recover"
	"net/http"
)

var UserRouter *iris.Application

func init() {
	app := iris.New()

	app.OnAnyErrorCode(router.Handler(func(c router.Context) {
		code := c.GetStatusCode()

		c.StatusCode(code)

		c.JSON(fmt.Errorf("%d %s", code, http.StatusText(code)), nil, nil)
	}))

	v1 := app.Party("/v1").AllowMethods(iris.MethodOptions)

	{
		v1.Use(recover.New())
		v1.Use(middleware.Common())
		v1.Use(middleware.CORS())
		v1.Use(middleware.RateLimit(30))

		if config.Common.Mode != "production" {
			v1.Use(logger.New())
			v1.Use(middleware.Ip())
		}

		{
			v1.Get("", router.Handler(func(c router.Context) {
				c.JSON(nil, map[string]string{"ping": "tong"}, nil)
			}))
		}

		userAuthMiddleware := middleware.AuthenticateNew(false) // 用户Token的中间件

		// 认证类
		{
			authRouter := v1.Party("/auth")
			authRouter.Post("/signup/email", auth.SignUpWithEmailRouter)   // 注册账号，通过邮箱+验证码
			authRouter.Post("/signup/phone", auth.SignUpWithPhoneRouter)   // 注册账号，通过手机+验证码
			authRouter.Post("/signup", auth.SignUpWithUsernameRouter)      // 注册账号, 通过用户名+密码
			authRouter.Post("/signin/email", auth.SignInWithEmailRouter)   // 邮箱+验证码 登陆
			authRouter.Post("/signin/phone", auth.SignInWithPhoneRouter)   // 手机+验证码 登陆
			authRouter.Post("/signin/wechat", auth.SignInWithWechatRouter) // 微信帐号登陆
			authRouter.Post("/signin/oauth2", auth.SignInWithOAuthRouter)  // oAuth 码登陆
			authRouter.Post("/signin", auth.SignInRouter)                  // 登陆账号
			authRouter.Put("/password/reset", auth.ResetPasswordRouter)    // 密码重置
			authRouter.Post("/code/email", auth.SendEmailAuthCodeRouter)   // 发送邮箱验证码，验证邮箱是否为用户所有 TODO: 缺少测试用例
			authRouter.Post("/code/phone", auth.SendPhoneAuthCodeRouter)   // 发送手机验证码，验证手机是否为用户所有 TODO: 缺少测试用例
		}

		// oAuth2 认证
		{
			oAuthRouter := v1.Party("/oauth2")
			oAuthRouter.Get("/{provider}", oauth2.AuthRouter)                  // 前去进行 oAuth 认证
			oAuthRouter.Get("/{provider}/callback", oauth2.AuthCallbackRouter) // 认证成功后，跳转回来的回调地址
		}

		// 用户类
		{
			userRouter := v1.Party("/user")
			userRouter.Use(userAuthMiddleware)
			userRouter.Get("/signout", user.SignOut)                                                                              // 用户登出
			userRouter.Get("/profile", user.GetProfileRouter)                                                                     // 获取用户详细信息
			userRouter.Put("/profile", middleware.Permission(*accession.ProfileUpdate), user.UpdateProfileRouter)                 // 更新用户资料
			userRouter.Put("/password", middleware.Permission(*accession.PasswordUpdate), user.UpdatePasswordRouter)              // 更新登陆密码
			userRouter.Post("/password2", middleware.Permission(*accession.Password2Set), user.SetPayPasswordRouter)              // 设置交易密码
			userRouter.Put("/password2", middleware.Permission(*accession.Password2Update), user.UpdatePayPasswordRouter)         // 更新交易密码
			userRouter.Put("/password2/reset", middleware.Permission(*accession.Password2Reset), user.ResetPayPasswordRouter)     // 重置交易密码
			userRouter.Get("/password2/reset", middleware.Permission(*accession.Password2Reset), user.SendResetPayPasswordRouter) // 发送重置交易密码的邮件/短信 			// 上传用户头像

			// 验证码类
			{
				authRouter := userRouter.Party("/auth")

				authRouter.Post("/email", user.SendAuthEmailRouter) // 发送邮箱验证码到用户绑定的邮箱 TODO: 缺少测试用例
				authRouter.Post("/phone", user.SendAuthPhoneRouter) // 发送手机验证码 TODO: 缺少测试用例
			}

			// 绑定类
			{
				bindRouter := userRouter.Party("/bind")
				bindRouter.Post("/email", auth.BindingEmailRouter)   // 绑定邮箱 TODO: 缺少测试用例
				bindRouter.Post("/phone", auth.BindingPhoneRouter)   // 绑定手机号 TODO: 缺少测试用例
				bindRouter.Post("/wechat", auth.BindingWechatRouter) // 绑定微信小程序 TODO: 缺少测试用例

				unbindRouter := userRouter.Party("/unbind")
				unbindRouter.Delete("/email", auth.UnbindingEmailRouter)   // 解除邮箱绑定 TODO: 缺少测试用例
				unbindRouter.Delete("/phone", auth.UnbindingPhoneRouter)   // 解除手机号绑定 TODO: 缺少测试用例
				unbindRouter.Delete("/wechat", auth.UnbindingWechatRouter) // 解除微信小程序绑定 TODO: 缺少测试用例
			}

			// 邀请人列表
			{
				inviteRouter := userRouter.Party("/invite")
				inviteRouter.Get("", invite.GetInviteListByUserRouter) // 获取我已邀请的列表
				inviteRouter.Get("/{invite_id}", invite.GetRouter)     // 获取单条邀请记录详情
			}

			// 收货地址
			{
				addressRouter := userRouter.Party("/address")
				addressRouter.Get("", address.GetAddressListByUserRouter)   // 获取地址列表
				addressRouter.Post("", address.CreateRouter)                // 添加收货地址
				addressRouter.Get("/default", address.GetDefaultRouter)     // 获取默认地址
				addressRouter.Put("/{address_id}", address.UpdateRouter)    // 更新收货地址
				addressRouter.Delete("/{address_id}", address.DeleteRouter) // 删除收货地址
				addressRouter.Get("/{address_id}", address.GetDetailRouter) // 获取地址详情
			}
		}

		// 钱包类
		{
			walletRouter := v1.Party("/wallet")
			walletRouter.Use(userAuthMiddleware)
			walletRouter.Get("", wallet.GetWalletsRouter)           // 获取所有钱包列表
			walletRouter.Get("/{currency}", wallet.GetWalletRouter) // 获取单个钱包的详细信息
		}

		// 转账类
		{
			transferRouter := v1.Party("/transfer")
			transferRouter.Use(userAuthMiddleware)
			transferRouter.Get("", transfer.GetHistoryRouter)                                                                       // 获取我的转账记录
			transferRouter.Post("", middleware.Permission(*accession.DoTransfer), middleware.AuthPayPasswordNew, transfer.ToRouter) // 转账给某人
			transferRouter.Get("/{transfer_id}", transfer.GetDetailRouter)                                                          // 获取单条转账详情
		}

		// 财务日志
		{
			financeRouter := v1.Party("/finance")
			financeRouter.Use(userAuthMiddleware)
			financeRouter.Get("/history", finance.GetHistory) // TODO: 获取我的财务日志
		}

		// 新闻咨询类
		{
			newsRouter := v1.Party("/news")
			newsRouter.Get("", news.GetNewsListRouter)  // 获取新闻公告列表
			newsRouter.Get("/{id}", news.GetNewsRouter) // 获取单个新闻公告详情
		}

		// 系统通知
		{
			notificationRouter := v1.Party("/notification")
			notificationRouter.Use(userAuthMiddleware)
			notificationRouter.Get("", notification.GetNotificationListByUserRouter) // 获取系统通知列表
			notificationRouter.Put("/read/batch", notification.MarkReadBatchRouter)  // 标记多个通知为已读
			notificationRouter.Put("/{id}/read", notification.ReadRouter)            // 标记通知为已读
			notificationRouter.Get("/{id}", notification.GetRouter)                  // 获取某一条系统通知详情
		}

		// 用户的个人消息, 个人消息是可以删除的
		{
			messageRouter := v1.Party("/message")
			messageRouter.Use(userAuthMiddleware)
			messageRouter.Get("/status", message.GetStatusRouter)            // 获取单个消息详情
			messageRouter.Put("/read/all", message.ReadAllRouter)            // 全部已读
			messageRouter.Put("/read/batch", message.ReadBatchRouter)        // 多个已读
			messageRouter.Get("/:message_id", message.GetRouter)             // 获取单个消息详情
			messageRouter.Put("/:message_id/read", message.ReadRouter)       // 标记消息为已读
			messageRouter.Delete("/:message_id", message.DeleteByUserRouter) // 删除消息
			messageRouter.Get("", message.GetMessageListByUserRouter)        // 获取我的消息列表
		}

		// 用户反馈
		{
			reportRouter := v1.Party("/report")
			reportRouter.Use(userAuthMiddleware)
			reportRouter.Get("/type", report.GetTypesRouter)         // 获取反馈类型列表
			reportRouter.Get("", report.GetListRouter)               // 获取我的反馈列表
			reportRouter.Post("", report.CreateRouter)               // 添加一条反馈
			reportRouter.Get("/{report_id}", report.GetReportRouter) // 获取反馈详情
			reportRouter.Put("/{report_id}", report.UpdateRouter)    // 更新这条反馈信息
		}

		// 帮助中心
		{
			helpRouter := v1.Party("/help")
			helpRouter.Get("", help.GetHelpListRouter)       // 创建帮助列表
			helpRouter.Get("/{help_id}", help.GetHelpRouter) // 获取帮助详情
		}

		// Banner
		{
			bannerRouter := v1.Party("/banner")
			bannerRouter.Get("", banner.GetBannerListRouter)         // 获取 banner 列表
			bannerRouter.Get("/{banner_id}", banner.GetBannerRouter) // 获取 banner 详情
		}

		// 地区接口
		{
			areaRouter := v1.Party("/area")
			areaRouter.Get("/provinces", area.GetProvinces)      // 获取省份
			areaRouter.Get("/{code}", area.GetDetail)            // 获取地区码的详情
			areaRouter.Get("/{code}/children", area.GetChildren) // 获取地区下的子地区
			areaRouter.Get("", area.GetArea)                     // 获取所有地区
		}

		// 通用类
		{
			// 邮件服务
			v1.Post("/email/send/register", auth.SignUpWithEmailActionRouter)         // 发送注册邮件
			v1.Post("/email/send/password/reset", email.SendResetPasswordEmailRouter) // 发送密码重置邮件

			// 数据签名
			v1.Post("/signature", signature.EncryptionRouter)
		}
	}

	_ = app.Build()

	UserRouter = app
}
