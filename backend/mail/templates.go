package mail

import (
	"fmt"
	"github.com/M1chaCH/deployment-controller/framework/logs"
	"html/template"
	"io"
	"os"
)

// InitTemplates reads all template files in
// IMPORTANT: Template files must end with an empty line
func InitTemplates() {
	var err error
	contactRequestTemplate, err = template.New(contactRequestTemplateName).Parse(mustReadTemplateFile(contactRequestFile))
	onboardingCompleteTemplate, err = template.New(onboardingCompleteTemplateName).Parse(mustReadTemplateFile(onboardingCompleteFile))
	mfaCodeTemplate, err = template.New(mfaCodeTemplateName).Parse(mustReadTemplateFile(mfaCodeFile))
	if err != nil {
		logs.Panic(fmt.Sprintf("Error loading contact request template: %v", err))
	}
}

const contactRequestFile = "mail/files/contact-request.html"
const contactRequestTemplateName = "contact-request"

var contactRequestTemplate *template.Template

type ContactRequestMailData struct {
	ClientId string
	DeviceId string
	Sender   string
	Message  string
}

func ParseContactRequestTemplate(writer io.WriteCloser, data ContactRequestMailData) error {
	return contactRequestTemplate.ExecuteTemplate(writer, contactRequestTemplateName, data)
}

const onboardingCompleteFile = "mail/files/onboarding-complete.html"
const onboardingCompleteTemplateName = "onboarding-complete"

var onboardingCompleteTemplate *template.Template

type OnboardingCompleteMailData struct {
	UserMail       string
	UserPageAccess string
}

func ParseOnboardingCompleteTemplate(writer io.WriteCloser, data OnboardingCompleteMailData) error {
	return onboardingCompleteTemplate.ExecuteTemplate(writer, onboardingCompleteTemplateName, data)
}

const mfaCodeFile = "mail/files/mfa-mail.html"
const mfaCodeTemplateName = "mfa-mail"

var mfaCodeTemplate *template.Template

type MfaCodeMailData struct {
	AdminMail string
	MfaCode   string
}

func ParseMfaCodeTemplate(writer io.WriteCloser, data MfaCodeMailData) error {
	return mfaCodeTemplate.ExecuteTemplate(writer, mfaCodeTemplateName, data)
}

func mustReadTemplateFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		logs.Panic(fmt.Sprintf("could not read file at: %s -- %v", path, err))
	}
	return string(data)
}
