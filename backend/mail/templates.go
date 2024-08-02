package mail

import (
	"fmt"
	"github.com/M1chaCH/deployment-controller/framework/logs"
	"html/template"
	"io"
	"os"
)

const contactRequestFile = "mail/files/contact-request.html"
const contactRequestTemplateName = "contact-request"

var contactRequestTemplate *template.Template

type ContactRequestMailData struct {
	ClientId string
	DeviceId string
	Sender   string
	Message  string
}

// InitTemplates reads all template files in
// IMPORTANT: Template files must end with an empty line
func InitTemplates() {
	var err error
	contactRequestTemplate, err = template.New(contactRequestTemplateName).Parse(mustReadTemplateFile(contactRequestFile))
	if err != nil {
		logs.Panic(fmt.Sprintf("Error loading contact request template: %v", err))
	}
}

func ParseContactRequestTemplate(writer io.WriteCloser, data ContactRequestMailData) error {
	return contactRequestTemplate.ExecuteTemplate(writer, contactRequestTemplateName, data)
}

func mustReadTemplateFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		logs.Panic(fmt.Sprintf("could not read file at: %s -- %v", path, err))
	}
	return string(data)
}
