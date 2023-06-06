package user

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/matcornic/hermes/v2"
	"github.com/spf13/cast"
	"gopkg.in/gomail.v2"
	"gorm.io/gorm"
	"io"
	"main/app/consts"
	g "main/app/global"
	"main/app/internal/model"
	"main/app/internal/service"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type InfoApi struct{}

var insInfo = InfoApi{}

func (a *InfoApi) UpdateAvatar(c *gin.Context) {
	fileHeader, err := c.FormFile("avatar")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  "internal err",
		})
		return
	}

	f, err := fileHeader.Open()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  "internal err",
		})
		return
	}

	contentBytes, err := io.ReadAll(f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  "internal err",
		})
		return
	}

	if len(contentBytes) > (1 << 19) {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "image is too large (please upload less than 512K)",
		})
		return
	}

	contentType := http.DetectContentType(contentBytes)
	if !strings.HasPrefix(contentType, "image") {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "invalid image type",
		})
		return
	}

	err = service.User().Info().UpdateAvatar(c, c.GetInt64("id"), contentBytes, contentType)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  "internal err",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "update avatar successfully",
	})
}

func (a *InfoApi) SetPaymentCode(c *gin.Context) {
	email := c.GetString("email")

	// 判断邮箱是否为空
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "email cannot be null",
		})
		return
	}
	// 通过正则表达式判断邮箱是否符合格式
	if ok, err := g.Reg.EmailReg.MatchString(email); !ok || err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "invalid email format",
		})
		return
	}

	// 判断邮箱是否重复
	// 从缓存中查找邮箱是否有重复
	isExist, err := g.Rdb.SIsMember(c,
		consts.RdbKeyRegisterEmail,
		email).Result()
	if err != nil {
		g.Logger.Errorf("get [register_email_set] cache failed, err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  "internal err",
		})
		return
	}
	if !isExist {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "invalid email",
		})
		return
	}

	// 生成邮箱验证码
	code := fmt.Sprintf("%06v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(1000000))

	// 从 redis 缓存中该邮箱验证码
	err = g.Rdb.Set(c,
		consts.RdbKeySetPaymentCode+email,
		code,
		5*time.Minute).Err()
	if err != nil {
		g.Logger.Errorf("set [set_payment_code] cache failed, err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  "internal err",
		})
		return
	}

	// 生成邮件内容
	h := hermes.Hermes{
		Theme: new(hermes.Flat),
		Product: hermes.Product{
			Name:      "电子钱包",
			Link:      fmt.Sprintf("https://%s", g.Config.App.Domain),
			Logo:      fmt.Sprintf("https://%s/favicon.ico", g.Config.App.Domain),
			Copyright: "Copyright © 2023 电子钱包. All rights reserved.",
		},
		DisableCSSInlining: false,
	}

	msg := hermes.Email{
		Body: hermes.Body{
			Intros: []string{
				"欢迎来到\"电子钱包\"网站!",
			},
			Actions: []hermes.Action{
				{
					Instructions: "重设支付密码验证码 (5分钟内有效):",
					InviteCode:   code,
				},
			},
			Outros: []string{
				"如果有什么问题请直接回复这封邮件",
			},
			Greeting:  "你好",
			Signature: "来自",
		},
	}

	// 渲染邮件内容
	emailBody, err := h.GenerateHTML(msg)
	if err != nil {
		g.Logger.Errorf("generate email failed, err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  "internal err",
		})
		return
	}

	// 发送邮件
	m := gomail.NewMessage()
	m.SetHeader("From", g.Config.EMail.Account)
	m.SetHeader("To", email)
	m.SetHeader("Subject", "电子钱包-重设支付密码验证码")
	m.SetBody("text/html", emailBody)
	d := gomail.NewDialer(g.Config.App.PrimaryDomain, 587, g.Config.EMail.Account, g.Config.EMail.Password)
	err = d.DialAndSend(m)
	if err != nil {
		g.Logger.Errorf("send email failed, err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  "internal err",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "send register email successfully",
	})
}

func (a *InfoApi) DoSetPaymentCode(c *gin.Context) {
	// 从表单获取信息
	userId := c.GetInt64("id")
	email := c.GetString("email")
	paymentCode := c.PostForm("paymentCode")
	code := c.PostForm("code")

	// 判断邮箱是否为空
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "email cannot be null",
		})
		return
	}
	// 判断密码是否为空
	if paymentCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "paymentCode cannot be null",
		})
		return
	}

	// 通过正则表达式判断支付密码是否符合格式
	if ok, err := g.Reg.PaymentCodeReg.MatchString(paymentCode); !ok || err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "invalid paymentCode format",
		})
		return
	}
	// 通过正则表达式判断邮箱是否符合格式
	if ok, err := g.Reg.EmailReg.MatchString(email); !ok || err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "invalid email format",
		})
		return
	}
	// 通过正则表达式判断验证码是否符合格式
	if ok, err := g.Reg.CodeReg.MatchString(code); !ok || err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "invalid code",
		})
		return
	}

	// 校验验证码
	storedCode, err := g.Rdb.Get(c,
		consts.RdbKeySetPaymentCode+email).Result()
	if err != nil {
		g.Logger.Errorf("get [set_payment_code] cache failed, err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  "internal err",
		})
		return
	}
	if storedCode != code {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "invalid code",
		})
	}

	var storedPassword string

	err = g.MysqlDB.WithContext(c).
		Table(model.TableNameUserSubject).
		Select("`password`").
		Where("`id` = ?", userId).
		Take(&storedPassword).Error
	if err != nil {
		g.Logger.Errorf("get mysql record failed, err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  "internal err",
		})
		return
	}

	if service.User().Sign().EncryptPassword(paymentCode) == storedPassword {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "payment code is same with password",
		})
		return
	}

	// 在数据库中更新密码
	err = g.MysqlDB.WithContext(c).
		Table(model.TableNameUserSubject).
		Where("`email` = ?", email).
		Update("`payment_code`", service.User().Sign().EncryptPassword(paymentCode)).Error
	if err != nil {
		g.Logger.Errorf("update user paymentCode failed, err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  "internal err",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "reset paymentCode successfully",
	})
}

func (a *InfoApi) SignIn(c *gin.Context) {
	userId := c.GetInt64("id")

	res, err := g.Rdb.GetBit(c,
		fmt.Sprintf("%s%d:%d:%d",
			consts.RdbKeySignIn,
			userId,
			time.Now().Year(),
			time.Now().Month(),
		),
		int64(time.Now().Day()-1),
	).Result()
	if err != nil {
		g.Logger.Errorf("get redis bit failed, err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  "internal err",
		})
		return
	}

	if res == 1 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "already signed in",
		})
		return
	}

	err = g.Rdb.SetBit(c,
		fmt.Sprintf("%s%d:%d:%d",
			consts.RdbKeySignIn,
			userId,
			time.Now().Year(),
			time.Now().Month(),
		),
		int64(time.Now().Day()-1),
		1,
	).Err()
	if err != nil {
		g.Logger.Errorf("set redis bit failed, err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  "internal err",
		})
		return
	}

	amountInt := rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(10000)
	amount := cast.ToFloat64(amountInt) / 100

	err = g.MysqlDB.Transaction(func(tx *gorm.DB) error {
		err = tx.WithContext(c).
			Table(model.TableNameTransaction).
			Create(&model.Transaction{
				UserId:      userId,
				Amount:      amount,
				Description: fmt.Sprintf("获得签到红包 %f 元", amount),
			}).Error
		if err != nil {
			return err
		}

		err = tx.WithContext(c).
			Table(model.TableNameUserSubject).
			Where("`id` = ?", userId).
			Update("`balance` = `balance` + ?", amount).
			Error
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		g.Logger.Errorf("update mysql record failed, err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  "internal err",
		})
		return
	}
}
