package mail

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/M1chaCH/deployment-controller/framework/config"
	"github.com/M1chaCH/deployment-controller/framework/logs"
	"github.com/gin-gonic/gin"
	"io"
	"net/smtp"
	"sync"
	"time"
)

// SendMailToAdmin sends an E-Mail to the configured admin address
// the context should remain optional
func SendMailToAdmin(c *gin.Context, throttleId string, subject string, messageWriter func(writer io.WriteCloser) error) error {
	cnf := config.Config()
	return SendMail(c, throttleId, cnf.Mail.Receiver, subject, messageWriter)
}

var sendMailMutex sync.Mutex

// SendMail sends an E-Mail via SMTP
// the context should remain optional
func SendMail(c *gin.Context, throttleId string, receiver string, subject string, renderTemplate func(writer io.WriteCloser) error) error {
	sendMailMutex.Lock()
	defer sendMailMutex.Unlock()

	err := throttleSentMails(throttleId)
	if err != nil {
		return err
	}

	cnf := config.Config()

	client, err := smtp.Dial(fmt.Sprintf("%s:%d", cnf.Mail.SMTP.Host, cnf.Mail.SMTP.Port))
	if err != nil {
		logs.Warn(c, "SMTP Dial Error: %v", err)
		tryResetConnection(c, client)
		return err
	}

	err = client.Hello("deployment-controller.michu-tech.com")
	if err != nil {
		logs.Warn(c, "SMTP Hello Error: %v", err)
		tryResetConnection(c, client)
		return err
	}

	err = client.StartTLS(&tls.Config{
		ServerName: cnf.Mail.SMTP.Host,
	})
	if err != nil {
		logs.Warn(c, "SMTP Start TLS Error: %v", err)
		tryResetConnection(c, client)
		return err
	}

	auth := smtp.PlainAuth("deployment-controller.michu-tech.com", cnf.Mail.SMTP.User, cnf.Mail.SMTP.Password, cnf.Mail.SMTP.Host)
	err = client.Auth(auth)
	if err != nil {
		logs.Warn(c, "SMTP Auth Error: %v", err)
		tryResetConnection(c, client)
		return err
	}

	err = client.Mail(cnf.Mail.Sender)
	if err != nil {
		logs.Warn(c, "SMTP Mail Error: %v", err)
		tryResetConnection(c, client)
		return err
	}

	// more rcpt -> call this command more
	err = client.Rcpt(receiver)
	if err != nil {
		logs.Warn(c, "SMTP Rcpt Error: %v", err)
		tryResetConnection(c, client)
		return err
	}

	bodyWriter, err := client.Data()
	if err != nil {
		logs.Warn(c, "SMTP Data Error: %v", err)
		tryResetConnection(c, client)
		return err
	}
	mimeHeaders := "MIME-version: 1.0;\nContent-Type: text/html; charset=\"UTF-8\";\n\n"
	_, err = bodyWriter.Write([]byte(fmt.Sprintf("Subject: %s \n%s\n\n", subject, mimeHeaders)))
	err = renderTemplate(bodyWriter)
	if err != nil {
		logs.Warn(c, "SendMail failed to render body: %v", err)
		tryResetConnection(c, client)
		return err
	}

	err = bodyWriter.Close()
	if err != nil {
		logs.Warn(c, "SendMail failed to close body: %v", err)
		tryResetConnection(c, client)
		return err
	}

	err = client.Quit()
	if err != nil {
		logs.Warn(c, "SMTP Quit Error: %v", err)
		return err
	}

	return err
}

func tryResetConnection(c *gin.Context, client *smtp.Client) {
	err := client.Reset()
	if err != nil {
		logs.Warn(c, "failed to reset smtp connection: %v", err)
	}
}

const TechnicalNoThrottle = "technical"

var TooManyMailsError = errors.New("too many mails")

var throttleMap map[string][]time.Time

// throttleSendMails verifies that a user did not send more than x mails in the last x minutes
// NOTE: needs to be thread safe (is currently guaranteed by sendMailMutex)
func throttleSentMails(throttleId string) error {
	if throttleId == TechnicalNoThrottle {
		return nil
	}
	if throttleMap == nil {
		throttleMap = make(map[string][]time.Time)
	}

	maxCount := config.Config().Mail.MaxCount

	mailSendTimes, found := throttleMap[throttleId]
	if !found {
		newThrottles := make([]time.Time, 0)
		throttleMap[throttleId] = append(newThrottles, time.Now())
		return nil
	}

	mailSendTimes = removeOldThrottledMails(mailSendTimes)
	currentThrottledMails := len(mailSendTimes)
	if currentThrottledMails < maxCount {
		throttleMap[throttleId] = append(mailSendTimes, time.Now())
		return nil
	}

	// this should never happen
	if currentThrottledMails > maxCount {
		logs.Error(nil, "More mails throttled than allowed (to many) will block all, throttleId: %s count: %d", throttleId, currentThrottledMails)
		return TooManyMailsError
	}

	if currentThrottledMails == maxCount {
		// remove the oldest time and add current
		mailSendTimes = append(mailSendTimes[:0], mailSendTimes[1:]...)
		throttleMap[throttleId] = append(mailSendTimes, time.Now())
		return TooManyMailsError
	}

	// this should never be reached
	return nil
}

func removeOldThrottledMails(times []time.Time) []time.Time {
	if len(times) == 0 {
		return times
	}

	duration := config.Config().Mail.CountDuration
	removeFromNowInPast := time.Now().Add(-time.Duration(duration) * time.Minute)
	for i, t := range times {
		// times is expected to be sorted -> we can return the content of times from where the first time is in range
		if removeFromNowInPast.Before(t) {
			return times[i:]
		}
	}

	return make([]time.Time, 0)
}
