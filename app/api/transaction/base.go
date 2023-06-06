package transaction

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/xuri/excelize/v2"
	"gorm.io/gorm"
	g "main/app/global"
	"main/app/internal/model"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type BaseApi struct{}

var insBase = BaseApi{}

func (a *BaseApi) Export(c *gin.Context) {
	userId := c.GetInt64("id")
	rawStartDate := c.PostForm("start_date")
	rawEndDate := c.PostForm("end_date")

	startDate, err := time.Parse("2006-01-02", rawStartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "invalid start date",
		})
		return
	}
	endDate, err := time.Parse("2006-01-02", rawEndDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "invalid start date",
		})
		return
	}

	var transactions []*model.Transaction

	err := g.MysqlDB.WithContext(c).
		Table(model.TableNameTransaction).
		Where("`user_id` = ? AND `create_time` >= ? AND `create_time` <= ?", userId, startDate, endDate).
		Find(&transactions).Error
	if err != nil {
		g.Logger.Errorf("get mysql record failed, err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  "internal err",
		})
		return
	}

	excelFile := excelize.NewFile()
	defer func() {
		if err := excelFile.Close(); err != nil {
			fmt.Println(err)
		}
	}()
	// Create a new sheet.
	_, _ = excelFile.NewSheet("Sheet1")

	_ = excelFile.SetDefaultFont("宋体")

	headStyle, _ := excelFile.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "left", Color: "000000", Style: 2},
			{Type: "top", Color: "000000", Style: 2},
			{Type: "bottom", Color: "000000", Style: 2},
			{Type: "right", Color: "000000", Style: 2},
		},
		Fill: excelize.Fill{},
		Font: &excelize.Font{
			Bold:         true,
			Italic:       false,
			Underline:    "",
			Family:       "",
			Size:         12,
			Strike:       false,
			Color:        "000000",
			ColorIndexed: 0,
			ColorTheme:   nil,
			ColorTint:    0,
			VertAlign:    "",
		},
		Alignment: &excelize.Alignment{
			Horizontal:      "center",
			Indent:          0,
			JustifyLastLine: false,
			ReadingOrder:    0,
			RelativeIndent:  0,
			ShrinkToFit:     false,
			TextRotation:    0,
			Vertical:        "center",
			WrapText:        false,
		},
		Protection:    nil,
		NumFmt:        0,
		DecimalPlaces: 0,
		CustomNumFmt:  nil,
		Lang:          "",
		NegRed:        false,
	})

	_ = excelFile.SetCellStyle("Sheet1", "A1", "I1", headStyle)

	// Set value of a cell.
	_ = excelFile.SetRowHeight("Sheet1", 1, 18)
	_ = excelFile.SetColWidth("Sheet1", "A", "A", 30)
	_ = excelFile.SetColWidth("Sheet1", "B", "B", 15)
	_ = excelFile.SetColWidth("Sheet1", "I", "I", 60)
	_ = excelFile.SetCellStr("Sheet1", "A1", "交易记录id")
	_ = excelFile.SetCellStr("Sheet1", "B1", "交易金额")
	_ = excelFile.SetCellStr("Sheet1", "C1", "交易详细")

	if len(transactions) != 0 {
		bodyStyle, _ := excelFile.NewStyle(&excelize.Style{
			Border: []excelize.Border{
				{Type: "left", Color: "000000", Style: 1},
				{Type: "top", Color: "000000", Style: 1},
				{Type: "bottom", Color: "000000", Style: 1},
				{Type: "right", Color: "000000", Style: 1},
			},
			Fill: excelize.Fill{},
			Font: &excelize.Font{
				Bold:         false,
				Italic:       false,
				Underline:    "",
				Family:       "",
				Size:         12,
				Strike:       false,
				Color:        "000000",
				ColorIndexed: 0,
				ColorTheme:   nil,
				ColorTint:    0,
				VertAlign:    "",
			},
			Alignment: &excelize.Alignment{
				Horizontal:      "center",
				Indent:          0,
				JustifyLastLine: false,
				ReadingOrder:    0,
				RelativeIndent:  0,
				ShrinkToFit:     false,
				TextRotation:    0,
				Vertical:        "center",
				WrapText:        false,
			},
			Protection:    nil,
			NumFmt:        0,
			DecimalPlaces: 0,
			CustomNumFmt:  nil,
			Lang:          "",
			NegRed:        false,
		})

		_ = excelFile.SetCellStyle("Sheet1", "A2", fmt.Sprintf("I%d", len(transactions)+1), bodyStyle)

		for i, info := range transactions {
			_ = excelFile.SetCellStr("Sheet1", fmt.Sprintf("A%d", i+2), strconv.FormatInt(info.Id, 10))
			_ = excelFile.SetCellStr("Sheet1", fmt.Sprintf("B%d", i+2), cast.ToString(info.Amount))
			_ = excelFile.SetCellStr("Sheet1", fmt.Sprintf("C%d", i+2), info.Description)
		}
	}

	buf, _ := excelFile.WriteToBuffer()

	c.Writer.Header().Set("content-type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	c.Writer.Header().Set("Content-Disposition",
		fmt.Sprintf("attachment; filename=\"%s-%s.xlsx\"", rawStartDate, rawEndDate))

	c.Writer.Header().Set("Content-length", strconv.Itoa(buf.Len()))
	_, _ = c.Writer.Write(buf.Bytes())

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "export successfully",
	})
}

func (a *BaseApi) TopUp(c *gin.Context) {
	userId := c.GetInt64("id")
	amount, err := cast.ToFloat64E(c.PostForm("amount"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "invalid amount",
		})
		return
	}

	if amount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "invalid amount",
		})
		return
	}

	if GetDecimalDigitLen(c.PostForm("amount")) > 2 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "invalid amount",
		})
		return
	}

	err = g.MysqlDB.Transaction(func(tx *gorm.DB) error {
		err = tx.WithContext(c).
			Table(model.TableNameTransaction).
			Create(&model.Transaction{
				UserId:      userId,
				Amount:      amount,
				Description: fmt.Sprintf("充值 %f 元", amount),
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

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "TopUp successfully",
	})
}

func (a *BaseApi) Withdraw(c *gin.Context) {
	userId := c.GetInt64("id")
	paymentCode := c.PostForm("payment_code")
	amount, err := cast.ToFloat64E(c.PostForm("amount"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "invalid amount",
		})
		return
	}

	if amount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "invalid amount",
		})
		return
	}

	if GetDecimalDigitLen(c.PostForm("amount")) > 2 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "invalid amount",
		})
		return
	}

	var storedPaymentCode string

	err = g.MysqlDB.WithContext(c).
		Table(model.TableNameUserSubject).
		Select("`payment_code`").
		Where("`id` = ?", userId).
		Take(&storedPaymentCode).
		Error

	if storedPaymentCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "payment code not set, please set first",
		})
		return
	}

	if paymentCode != storedPaymentCode {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "wrong payment code",
		})
		return
	}

	err = g.MysqlDB.Transaction(func(tx *gorm.DB) error {
		err = tx.WithContext(c).
			Table(model.TableNameTransaction).
			Create(&model.Transaction{
				UserId:      userId,
				Amount:      amount,
				Description: fmt.Sprintf("提现 %f 元", amount),
			}).Error

		err = tx.WithContext(c).
			Table(model.TableNameUserSubject).
			Where("`id` = ?", userId).
			Update("`balance` = `balance` - ?", amount).
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

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "withdraw successfully",
	})
}

func (a *BaseApi) Transfer(c *gin.Context) {
	srcUserId := c.GetInt64("id")
	srcPhone := c.GetString("phone")
	dstPhone := c.PostForm("phone")
	paymentCode := c.PostForm("payment_code")
	amount, err := cast.ToFloat64E(c.PostForm("amount"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "invalid amount",
		})
		return
	}

	if amount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "invalid amount",
		})
		return
	}

	if GetDecimalDigitLen(c.PostForm("amount")) > 2 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "invalid amount",
		})
		return
	}

	var storedPaymentCode string

	err = g.MysqlDB.WithContext(c).
		Table(model.TableNameUserSubject).
		Select("`payment_code`").
		Where("`id` = ?", srcUserId).
		Take(&storedPaymentCode).
		Error

	if storedPaymentCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "payment code not set, please set first",
		})
		return
	}

	if paymentCode != storedPaymentCode {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "wrong payment code",
		})
		return
	}

	var dstUserId int64
	err = g.MysqlDB.WithContext(c).
		Table(model.TableNameUserSubject).
		Select("`id`").
		Where("`phone` = ?", dstPhone).
		Take(&dstUserId).
		Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusBadRequest, gin.H{
				"code": http.StatusBadRequest,
				"msg":  "user not found",
			})
			return
		}

		g.Logger.Errorf("query mysql record failed, err: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"code": http.StatusInternalServerError,
			"msg":  "internal err",
		})
		return
	}

	err = g.MysqlDB.Transaction(func(tx *gorm.DB) error {
		err = tx.WithContext(c).
			Table(model.TableNameTransaction).
			Create(&model.Transaction{
				UserId:      srcUserId,
				Amount:      amount,
				Description: fmt.Sprintf("向用户 %s 转账 %f 元", dstPhone, amount),
			}).Error

		err = tx.WithContext(c).
			Table(model.TableNameTransaction).
			Create(&model.Transaction{
				UserId:      dstUserId,
				Amount:      amount,
				Description: fmt.Sprintf("用户 %s 向你转账 %f 元", srcPhone, amount),
			}).Error

		err = tx.WithContext(c).
			Table(model.TableNameUserSubject).
			Where("`id` = ?", srcUserId).
			Update("`balance` = `balance` - ?", amount).
			Error
		if err != nil {
			return err
		}
		err = tx.WithContext(c).
			Table(model.TableNameUserSubject).
			Where("`id` = ?", dstUserId).
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

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "withdraw successfully",
	})
}

func (a *BaseApi) Consume(c *gin.Context) {
	userId := c.GetInt64("id")
	paymentCode := c.PostForm("payment_code")
	amount, err := cast.ToFloat64E(c.PostForm("amount"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "invalid amount",
		})
		return
	}

	if amount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "invalid amount",
		})
		return
	}

	if GetDecimalDigitLen(c.PostForm("amount")) > 2 {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "invalid amount",
		})
		return
	}

	var storedPaymentCode string

	err = g.MysqlDB.WithContext(c).
		Table(model.TableNameUserSubject).
		Select("`payment_code`").
		Where("`id` = ?", userId).
		Take(&storedPaymentCode).
		Error

	if storedPaymentCode == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "payment code not set, please set first",
		})
		return
	}

	if paymentCode != storedPaymentCode {
		c.JSON(http.StatusBadRequest, gin.H{
			"code": http.StatusBadRequest,
			"msg":  "wrong payment code",
		})
		return
	}

	err = g.MysqlDB.Transaction(func(tx *gorm.DB) error {
		err = tx.WithContext(c).
			Table(model.TableNameTransaction).
			Create(&model.Transaction{
				UserId:      userId,
				Amount:      amount,
				Description: fmt.Sprintf("消费 %f 元", amount),
			}).Error

		err = tx.WithContext(c).
			Table(model.TableNameUserSubject).
			Where("`id` = ?", userId).
			Update("`balance` = `balance` - ?", amount).
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

	c.JSON(http.StatusOK, gin.H{
		"code": 0,
		"msg":  "withdraw successfully",
	})
}

func GetDecimalDigitLen(numStr string) int {
	tmp := strings.Split(numStr, ".")
	if len(tmp) <= 1 {
		return 0
	}
	return len(tmp[1])
}
