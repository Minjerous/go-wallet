package boot

import (
	"github.com/dlclark/regexp2"
	g "main/app/global"
)

func RegexpSetup() {
	// 密码强度必须包含大写字母, 小写字母, 数字, 特殊字符中的两种或两种以上, 8-20位
	passwordReg := regexp2.MustCompile(`^(?![\d]+$)(?![a-z]+$)(?![A-Z]+$)(?![~!@#$%^&*.]+$)[\da-zA-z~!@#$%^&*.]{8,20}$`, regexp2.None)
	// 支付密码6位数字
	paymentCodeReg := regexp2.MustCompile("^[0-9]{6}$", regexp2.None)
	// 名称允许汉字、字母、数字，域名只允许英文域名
	emailReg := regexp2.MustCompile(`^[A-Za-z0-9\u4e00-\u9fa5]+@[a-zA-Z0-9_-]+(\.[a-zA-Z_-]+)+$`, regexp2.None)
	// 6位纯数字
	codeReg := regexp2.MustCompile(`^[0-9]{6}$`, regexp2.None)

	g.Reg = &g.Regexp{
		PasswordReg:    passwordReg,
		PaymentCodeReg: paymentCodeReg,
		EmailReg:       emailReg,
		CodeReg:        codeReg,
	}
}
