package user

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"github.com/minio/minio-go/v7"
	"golang.org/x/crypto/sha3"
	"gorm.io/gorm"
	g "main/app/global"
	"main/app/internal/model"
)

type SInfo struct{}

var insInfo = SInfo{}

func (s *SInfo) UpdateAvatar(ctx context.Context, userId int64, avatar []byte, contentType string) error {
	buffer := bytes.NewBuffer(avatar)

	d := sha3.Sum224(avatar)
	hash := hex.EncodeToString(d[:])

	userSubject := &model.UserSubject{}

	err := g.MysqlDB.WithContext(ctx).
		Table("`user_subject`").
		Where("`avatar` = ?", hash).
		Take(userSubject).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			_, err = g.MinioClient.PutObject(
				ctx,
				g.MinioBucket,
				fmt.Sprintf("avatar/%s", hash),
				buffer,
				int64(len(avatar)),
				minio.PutObjectOptions{ContentType: contentType},
			)
			if err != nil {
				return err
			}
		} else {
			return err
		}
	}

	err = g.MysqlDB.WithContext(ctx).
		Table("`user_subject`").
		Where("`id` = ?", userId).
		Update("`avatar`", hash).Error
	if err != nil {
		return err
	}

	return nil
}
