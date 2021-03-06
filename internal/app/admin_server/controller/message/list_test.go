// Copyright 2019-2020 Axetroy. All rights reserved. MIT license.
package message_test

import (
	"encoding/json"
	"fmt"
	"github.com/axetroy/go-server/internal/app/admin_server/controller/message"
	"github.com/axetroy/go-server/internal/library/helper"
	"github.com/axetroy/go-server/internal/schema"
	"github.com/axetroy/go-server/internal/service/token"
	"github.com/axetroy/go-server/tester"
	"github.com/axetroy/mocker"
	"github.com/stretchr/testify/assert"
	"net/http"
	"testing"
)

func TestGetMessageListByAdmin(t *testing.T) {

	{
		var (
			data = make([]schema.Message, 0)
		)
		query := schema.Query{
			Limit: 20,
		}
		r := message.GetMessageListByAdmin(helper.Context{
			Uid: "123123",
		}, message.Query{
			Query: query,
		})

		fmt.Printf("%+v\n", r)

		assert.Equal(t, schema.StatusSuccess, r.Status)
		assert.Equal(t, "", r.Message)

		assert.Nil(t, r.Decode(&data))
		assert.NotNil(t, r.Meta.Limit)
		assert.Equal(t, query.Limit, r.Meta.Limit)
		assert.Equal(t, schema.DefaultPage, r.Meta.Page)
	}

	adminInfo, _ := tester.LoginAdmin()

	userInfo, _ := tester.CreateUser()

	defer tester.DeleteUserByUserName(userInfo.Username)

	{
		var (
			title   = "test"
			content = "test"
		)

		r := message.Create(helper.Context{
			Uid: adminInfo.Id,
		}, message.CreateMessageParams{
			Uid:     userInfo.Id,
			Title:   title,
			Content: content,
		})

		assert.Equal(t, schema.StatusSuccess, r.Status)
		assert.Equal(t, "", r.Message)

		n := schema.Message{}

		assert.Nil(t, r.Decode(&n))

		defer message.DeleteMessageById(n.Id)
	}

	// 3. 获取列表
	{
		data := make([]schema.Message, 0)

		query := message.Query{
			Query: schema.Query{
				Limit: 20,
			},
		}
		r := message.GetMessageListByAdmin(helper.Context{
			Uid: adminInfo.Id,
		}, query)

		assert.Equal(t, schema.StatusSuccess, r.Status)
		assert.Equal(t, "", r.Message)

		assert.Nil(t, r.Decode(&data))

		assert.Equal(t, query.Limit, r.Meta.Limit)
		assert.Equal(t, schema.DefaultPage, r.Meta.Page)
		assert.True(t, r.Meta.Num >= 1)
		assert.True(t, r.Meta.Total >= 1)
		assert.True(t, len(data) >= 1)
	}
}

func TestGetMessageListByAdminRouter(t *testing.T) {
	adminInfo, _ := tester.LoginAdmin()

	userInfo, _ := tester.CreateUser()

	defer tester.DeleteUserByUserName(userInfo.Username)

	{
		var (
			title   = "test"
			content = "test"
		)

		r := message.Create(helper.Context{
			Uid: adminInfo.Id,
		}, message.CreateMessageParams{
			Uid:     userInfo.Id,
			Title:   title,
			Content: content,
		})

		assert.Equal(t, schema.StatusSuccess, r.Status)
		assert.Equal(t, "", r.Message)

		n := schema.Message{}

		assert.Nil(t, r.Decode(&n))

		//defer message.DeleteMessageById(n.Id)
	}

	header := mocker.Header{
		"Authorization": token.Prefix + " " + adminInfo.Token,
	}

	{
		r := tester.HttpAdmin.Get("/v1/message", nil, &header)

		res := schema.Response{}

		assert.Equal(t, http.StatusOK, r.Code)

		if !assert.Nil(t, json.Unmarshal(r.Body.Bytes(), &res)) {
			return
		}

		if !assert.Equal(t, schema.StatusSuccess, res.Status) {
			return
		}

		if !assert.Equal(t, "", res.Message) {
			return
		}

		messages := make([]schema.Message, 0)

		assert.Nil(t, res.Decode(&messages))

		assert.True(t, len(messages) > 0)

		for _, b := range messages {
			assert.IsType(t, "string", b.Title)
			assert.IsType(t, "string", b.Content)
			assert.IsType(t, "string", b.CreatedAt)
			assert.IsType(t, "string", b.UpdatedAt)
		}
	}
}
