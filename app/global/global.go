package global

import (
	"github.com/dlclark/regexp2"
	"github.com/minio/minio-go/v7"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"gorm.io/gorm"
	"main/app/internal/model/config"
)

const (
	MinioBucket = "wallet"
)

type Regexp struct {
	PasswordReg    *regexp2.Regexp
	PaymentCodeReg *regexp2.Regexp
	EmailReg       *regexp2.Regexp
	CodeReg        *regexp2.Regexp
}

var (
	Config      *config.Config
	Logger      *zap.SugaredLogger
	MysqlDB     *gorm.DB
	Rdb         *redis.Client
	MinioClient *minio.Client

	Reg *Regexp
)
