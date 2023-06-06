package boot

import (
	"context"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	g "main/app/global"
	"time"
)

func MysqlDBSetup() {
	config := g.Config.DataBase.Mysql

	db, err := gorm.Open(mysql.Open(config.GetDsn()))
	if err != nil {
		g.Logger.Fatalf("initialize mysql db failed, err: %v", err)
	}

	sqlDB, _ := db.DB()
	sqlDB.SetConnMaxIdleTime(10 * time.Second)
	sqlDB.SetConnMaxLifetime(100 * time.Second)
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	err = sqlDB.Ping()
	if err != nil {
		g.Logger.Fatalf("connect to mysql db failed, err: %v", err)
	}

	g.Logger.Infof("initialize mysql db successfully")
	g.MysqlDB = db
}

func RedisSetup() {
	config := g.Config.DataBase.Redis

	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.Addr, config.Port),
		Username: "",
		Password: config.Password,
		DB:       config.Db,
		PoolSize: 10000,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		g.Logger.Fatalf("connect to redis instance failed, err: %v", err)
	}

	g.Rdb = rdb

	g.Logger.Info("initialize redis client successfully")
}

func MinioSetup() {
	config := g.Config.DataBase

	endpoint := config.Minio.Endpoint
	username := config.Minio.Username
	password := config.Minio.Password

	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:           credentials.NewStaticV4(username, password, ""),
		Secure:          true,
		Transport:       nil,
		Region:          "",
		BucketLookup:    0,
		TrailingHeaders: false,
		CustomMD5:       nil,
		CustomSHA256:    nil,
	})
	if err != nil {
		g.Logger.Fatalf("connect to minio instance failed, err: %v", err)
	}

	g.MinioClient = minioClient

	g.Logger.Info("initialize minio client successfully")
}
