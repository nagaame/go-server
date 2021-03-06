// Copyright 2019-2020 Axetroy. All rights reserved. MIT license.
package config

import (
	"github.com/axetroy/go-server/internal/service/dotenv"
)

type database struct {
	Host         string `json:"host"`
	Port         string `json:"port"`
	Driver       string `json:"driver"`
	DatabaseName string `json:"database_name"`
	Username     string `json:"username"`
	Password     string `json:"password"`
}

var Database database

func init() {
	Database.Driver = dotenv.GetByDefault("DB_DRIVER", "postgres")
	Database.Host = dotenv.GetByDefault("DB_HOST", "localhost")
	Database.Port = dotenv.GetByDefault("DB_PORT", "65432")
	Database.DatabaseName = dotenv.GetByDefault("DB_NAME", "gotest")
	Database.Username = dotenv.GetByDefault("DB_USERNAME", "gotest")
	Database.Password = dotenv.GetByDefault("DB_PASSWORD", "gotest")
}
