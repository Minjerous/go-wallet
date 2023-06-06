package user

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/matcornic/hermes/v2"
	"github.com/speps/go-hashids/v2"
	"github.com/spf13/cast"
	"gopkg.in/gomail.v2"
	"main/app/consts"
	g "main/app/global"
	"main/app/internal/model"
	"main/app/internal/service"
	"main/utils/cookie"
	"math/rand"
	"net/http"
	"time"
)

type SignApi struct{}

var insSign = SignApi{}

func (a *SignApi) SendRegisterEmail(c *gin.Context) {
	// 从表单获取信息
	email := c.PostForm("email")

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
	if isExist {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "duplicate email",
		})
		return
	}

	// 生成邮箱验证码
	code := fmt.Sprintf("%06v", rand.New(rand.NewSource(time.Now().UnixNano())).Int31n(1000000))

	// 从 redis 缓存中该邮箱验证码
	err = g.Rdb.Set(c,
		consts.RdbKeyRegisterEmailCode+email,
		code,
		5*time.Minute).Err()
	if err != nil {
		g.Logger.Errorf("set [register_email] cache failed, err: %v", err)
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
					Instructions: "验证码 (5分钟内有效):",
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
	m.SetHeader("Subject", "电子钱包-注册验证码")
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

func (a *SignApi) DoRegister(c *gin.Context) {
	// 从表单获取信息
	username := c.PostForm("username")
	phone := c.PostForm("phone")
	password := c.PostForm("password")
	email := c.PostForm("email")
	code := c.PostForm("code")

	if username == "" {
		// 生成默认昵称
		username = fmt.Sprintf("wallet_%s", newRandomString(phone, "username", 10))
	}
	// 判断手机号是否为空
	if phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "phone cannot be null",
		})
		return
	}
	// 判断邮箱是否为空
	if email == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "email cannot be null",
		})
		return
	}
	// 判断密码是否为空
	if password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "password cannot be null",
		})
		return
	}

	// 通过正则表达式判断密码是否符合格式
	if ok, err := g.Reg.PasswordReg.MatchString(password); !ok || err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "invalid password format",
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
		consts.RdbKeyRegisterEmailCode+email).Result()
	if err != nil {
		g.Logger.Errorf("get [register_email] cache failed, err: %v", err)
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

	// 检查手机号是否唯一
	err = service.User().Sign().CheckPhoneIsExist(c, phone)
	if err != nil {
		if err.Error() == "internal err" {
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  err.Error(),
			})
		} else if err.Error() == "phone already exist" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  err.Error(),
			})
		}

		return
	}

	// 添加注册缓存
	err = g.Rdb.SAdd(c,
		consts.RdbKeyRegisterEmail,
		email).Err()
	if err != nil {
		g.Logger.Errorf("update user register cache failed, err: %v", err)
	}

	// 拼装用户信息结构体
	userSubject := &model.UserSubject{}

	// 加密密码
	encryptedPassword := service.User().Sign().EncryptPassword(password)

	userSubject.Username = username
	userSubject.Phone = phone
	userSubject.Password = encryptedPassword
	userSubject.Email = email
	userSubject.Avatar = "avatar.png"

	// 插入数据库中
	service.User().Sign().CreateUser(c, userSubject)

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "register successfully",
	})
}

func (a *SignApi) ResetPassword(c *gin.Context) {
	// 从表单获取信息
	email := c.PostForm("email")

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
		consts.RdbKeyResetPasswordCode+email,
		code,
		5*time.Minute).Err()
	if err != nil {
		g.Logger.Errorf("set [register_email] cache failed, err: %v", err)
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
					Instructions: "重设密码验证码 (5分钟内有效):",
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
	m.SetHeader("Subject", "电子钱包-重设密码验证码")
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

func (a *SignApi) DoResetPassword(c *gin.Context) {
	// 从表单获取信息
	email := c.PostForm("email")
	password := c.PostForm("password")
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
	if password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "password cannot be null",
		})
		return
	}

	// 通过正则表达式判断密码是否符合格式
	if ok, err := g.Reg.PasswordReg.MatchString(password); !ok || err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "invalid password format",
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
		consts.RdbKeyResetPasswordCode+email).Result()
	if err != nil {
		g.Logger.Errorf("get [reset_password] cache failed, err: %v", err)
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

	// 在数据库中更新密码
	err = g.MysqlDB.WithContext(c).
		Table(model.TableNameUserSubject).
		Where("`email` = ?", email).
		Update("`password`", service.User().Sign().EncryptPassword(password)).Error
	if err != nil {
		g.Logger.Errorf("update user password failed, err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  "internal err",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "reset password successfully",
	})
}

func (a *SignApi) Login(c *gin.Context) {
	// 从表单获取信息
	phone := c.PostForm("phone")
	password := c.PostForm("password")
	currentIp := c.RemoteIP()

	// 判断手机号是否为空
	if phone == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "phone cannot be null",
		})
		return
	}
	// 判断密码是否为空
	if password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "password cannot be null",
		})
		return
	}

	userSubject := &model.UserSubject{
		Phone:    phone,
		Password: service.User().Sign().EncryptPassword(password),
	}

	// 检查密码是否正确
	err := service.User().Sign().CheckPassword(c, userSubject)
	if err != nil {
		switch err.Error() {
		case "internal err":
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  err.Error(),
			})
		case "invalid phone or password":
			c.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  err.Error(),
			})
		}

		return
	}

	// 生成 jwt
	tokenString, err := service.User().Sign().GenerateToken(c, userSubject)
	if err != nil {
		switch err.Error() {
		case "internal err":
			c.JSON(http.StatusInternalServerError, gin.H{
				"code": http.StatusInternalServerError,
				"msg":  err.Error(),
			})
		}

	}

	// 将 jwt 写入到 cookie 中
	cookieConfig := g.Config.Auth.Cookie
	cookieWriter := cookie.NewCookieWriter(&cookie.Config{
		Secret: cookieConfig.Secret,
		Ctx:    c,
		Cookie: http.Cookie{
			Path:     "/",
			Domain:   cookieConfig.Domain,
			MaxAge:   cookieConfig.MaxAge,
			Secure:   cookieConfig.Secure,
			HttpOnly: cookieConfig.HttpOnly,
			SameSite: cookieConfig.SameSite,
		},
	})

	cookieWriter.Set("x-token", tokenString)

	var lastIp string
	msg := "login successfully"

	err = g.MysqlDB.WithContext(c).
		Table(model.TableNameUserSubject).
		Where("`phone` = ?", phone).
		Select("`last_ip`").
		Take(&lastIp).Error
	if err != nil {
		g.Logger.Errorf("get mysql record failed, err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  err.Error(),
		})
	}

	if lastIp != currentIp {
		if lastIp != "" {
			msg = fmt.Sprintf("login successfully (attention: current ip is different from last login ip, %s - > %s)", lastIp, currentIp)
		}

		go func() {
			err = g.MysqlDB.WithContext(context.TODO()).
				Table(model.TableNameUserSubject).
				Where("`phone` = ?", phone).
				Update("`last_ip`", currentIp).
				Error
			if err != nil {
				g.Logger.Errorf("udpate mysql record failed, err: %v", err)
			}
		}()
	}

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  msg,
		"data": gin.H{
			"id":       userSubject.Id,
			"username": userSubject.Username,
			"phone":    userSubject.Phone,
			"email":    userSubject.Email,
			"avatar":   userSubject.Avatar,
		},
	})
}

func newRandomString(src string, salt string, length int) string {
	hd := hashids.NewData()
	hd.Salt = salt
	hd.MinLength = length
	h, _ := hashids.NewWithData(hd)
	randomString, _ := h.Encode(cast.ToIntSlice([]uint8(src)))
	return randomString
}
