// Copyright 2019-2020 Axetroy. All rights reserved. MIT license.
package middleware

import (
	"github.com/axetroy/go-server/internal/library/exception"
	"github.com/axetroy/go-server/internal/schema"
	"github.com/axetroy/go-server/internal/service/authentication"
	"github.com/axetroy/go-server/internal/service/token"
	"github.com/kataras/iris/v12"
)

var (
	ContextUidField = "uid"
)

func getToken(c iris.Context) (*string, error) {

	authorization := c.Request().URL.Query().Get("Authorization")

	if len(authorization) > 0 {
		return &authorization, nil
	} else if len(c.GetHeader(token.AuthField)) > 0 {
		t := c.GetHeader(token.AuthField)
		return &t, nil
	} else {
		t := c.GetCookie(token.AuthField)

		if len(t) > 0 {
			return &t, nil
		}
	}

	return nil, nil
}

// Token 验证中间件
func AuthenticateNew(isAdmin bool) iris.Handler {
	return func(c iris.Context) {
		var (
			err    error
			status = schema.StatusFail
		)
		defer func() {
			if err != nil {
				_, _ = c.JSON(schema.Response{
					Status:  status,
					Message: err.Error(),
					Data:    nil,
				})
				return
			}

			c.Next()
		}()

		tokenString, err := getToken(c)

		if err != nil {
			return
		}

		if tokenString == nil {
			status = exception.InvalidToken.Code()
			err = exception.InvalidToken
			return
		}

		userId, err := authentication.Gateway(isAdmin).Parse(*tokenString)

		if err != nil {
			status = exception.InvalidToken.Code()
			return
		}

		c.Values().Set(ContextUidField, userId)
	}
}
