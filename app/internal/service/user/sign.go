package user

import (
	"context"
	"encoding/hex"
	"fmt"
	"golang.org/x/crypto/sha3"
	"gorm.io/gorm"
	g "main/app/global"
	"main/app/internal/model"
	"main/utils/jwt"
	"time"
)

type SSign struct{}

var insSign = SSign{}

func (s *SSign) CheckPhoneIsExist(ctx context.Context, phone string) error {
	userSubject := &model.UserSubject{}
	err := g.MysqlDB.WithContext(ctx).
		Table(model.TableNameUserSubject).
		Select("phone").
		Where("phone = ?", phone).
		First(userSubject).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			g.Logger.Errorf("query [user_subject] record failed, err: %v", err)
			return fmt.Errorf("internal err")
		}
	} else {
		return fmt.Errorf("phone already exist")
	}

	return nil
}

func (s *SSign) EncryptPassword(password string) string {
	d := sha3.Sum224([]byte(password))
	return hex.EncodeToString(d[:])
}

func (s *SSign) CreateUser(ctx context.Context, userSubject *model.UserSubject) {
	g.MysqlDB.WithContext(ctx).
		Table(model.TableNameUserSubject).
		Create(userSubject)
}

func (s *SSign) CheckPassword(ctx context.Context, userSubject *model.UserSubject) error {
	err := g.MysqlDB.WithContext(ctx).
		Table(model.TableNameUserSubject).
		Where(&model.UserSubject{
			Username: userSubject.Username,
			Password: userSubject.Password,
		}).
		First(userSubject).Error
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			g.Logger.Errorf("query [user_subject] record failed, err: %v", err)
			return fmt.Errorf("internal err")
		} else {
			return fmt.Errorf("invalid phone or password")
		}
	}

	return nil
}

func (s *SSign) GenerateToken(ctx context.Context, userSubject *model.UserSubject) (string, error) {
	jwtConfig := g.Config.Auth.Jwt

	j := jwt.NewJWT(&jwt.Config{
		SecretKey:   jwtConfig.SecretKey,
		ExpiresTime: jwtConfig.ExpiresTime,
		BufferTime:  jwtConfig.BufferTime,
		Issuer:      jwtConfig.Issuer})
	claims := j.CreateClaims(&jwt.BaseClaims{
		Id:         userSubject.Id,
		Email:      userSubject.Email,
		Phone:      userSubject.Phone,
		Username:   userSubject.Username,
		CreateTime: userSubject.CreateTime,
		UpdateTime: userSubject.UpdateTime,
	})

	tokenString, err := j.GenerateToken(&claims)
	if err != nil {
		g.Logger.Errorf("generate token failed, %v", err)
		return "", fmt.Errorf("internal err")
	}

	err = g.Rdb.Set(ctx,
		fmt.Sprintf("jwt_%d", userSubject.Id),
		tokenString,
		time.Duration(jwtConfig.ExpiresTime)*time.Second).Err()
	if err != nil {
		g.Logger.Errorf("set [jwt] cache failed, %v", err)
		return "", fmt.Errorf("internal err")
	}

	return tokenString, nil
}
