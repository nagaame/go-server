// Copyright 2019-2020 Axetroy. All rights reserved. MIT license.
package user

import (
	"errors"
	"github.com/axetroy/go-server/internal/library/exception"
	"github.com/axetroy/go-server/internal/library/helper"
	"github.com/axetroy/go-server/internal/library/router"
	"github.com/axetroy/go-server/internal/library/util"
	"github.com/axetroy/go-server/internal/library/validator"
	"github.com/axetroy/go-server/internal/model"
	"github.com/axetroy/go-server/internal/schema"
	"github.com/axetroy/go-server/internal/service/database"
	"github.com/jinzhu/gorm"
)

type UpdatePasswordByAdminParams struct {
	NewPassword string `json:"new_password" validate:"required" comment:"新密码"`
}

func UpdatePasswordByAdmin(c helper.Context, userId string, input UpdatePasswordByAdminParams) (res schema.Response) {
	var (
		err error
		tx  *gorm.DB
	)

	defer func() {
		if r := recover(); r != nil {
			switch t := r.(type) {
			case string:
				err = errors.New(t)
			case error:
				err = t
			default:
				err = exception.Unknown
			}
		}

		if tx != nil {
			if err != nil {
				_ = tx.Rollback().Error
			} else {
				err = tx.Commit().Error
			}
		}

		helper.Response(&res, nil, nil, err)
	}()

	// 参数校验
	if err = validator.ValidateStruct(input); err != nil {
		return
	}

	tx = database.Db.Begin()

	// 检查是否是管理员
	adminInfo := model.Admin{Id: c.Uid}

	if err = tx.First(&adminInfo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = exception.AdminNotExist
		}
		return
	}

	// 只有超级管理员才能操作
	if !adminInfo.IsSuper {
		err = exception.NoPermission
		return
	}

	userInfo := model.User{Id: userId}

	if err = tx.First(&userInfo).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			err = exception.UserNotExist
		}
		return
	}

	newPassword := util.GeneratePassword(input.NewPassword)

	if err = tx.Model(&userInfo).Update(model.User{Password: newPassword}).Error; err != nil {
		return
	}

	return
}

var UpdatePasswordByAdminRouter = router.Handler(func(c router.Context) {
	var (
		input UpdatePasswordByAdminParams
	)

	userId := c.Param("user_id")

	c.ResponseFunc(c.ShouldBindJSON(&input), func() schema.Response {
		return UpdatePasswordByAdmin(helper.NewContext(&c), userId, input)
	})
})
