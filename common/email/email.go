package email

import (
	"GopherAI/config"
	"fmt"
	"strings"

	"gopkg.in/gomail.v2"
)

const (
	CodeMsg     = "GopherAI验证码如下(验证码仅限于2分钟有效): "
	UserNameMsg = "GopherAI的账号如下，请保留好，后续可以用账号/邮箱登录 "
)

func SendCaptcha(email, code, msg string) error {
	conf := config.GetConfig().EmailConfig
	if conf.Email == "" || conf.Authcode == "" || conf.Email == "your qq email" || strings.Contains(conf.Authcode, "your authcode") {
		fmt.Printf("[dev email fallback] to=%s message=%s %s\n", email, msg, code)
		return nil
	}

	m := gomail.NewMessage()

	// 发件人
	m.SetHeader("From", conf.Email)
	// 收件人
	m.SetHeader("To", email)
	// 主题
	m.SetHeader("Subject", "来自GopherAI的信息")
	// 正文内容（纯文本形式，也可以用 text/html）
	m.SetBody("text/plain", msg+" "+code)

	// 配置 SMTP 服务器和授权码,587：是 SMTP 的明文/STARTTLS 端口号
	d := gomail.NewDialer("smtp.qq.com", 587, conf.Email, conf.Authcode)

	// 发送邮件
	if err := d.DialAndSend(m); err != nil {
		fmt.Printf("DialAndSend err %v:\n", err)
		return err
	}
	fmt.Printf("send mail success\n")
	return nil
}
