package mvp

import (
	"strings"

	"github.com/andreyvit/mvp/postmark"
)

type Email struct {
	From    string
	To      string
	ReplyTo string
	Cc      string
	Bcc     string
	Subject string

	View     string
	Layout   string
	Data     any
	TextBody string
	HtmlBody string

	ServerTemplateIntID int64
	ServerTemplateData  map[string]any

	MessageStream string
	Category      string
}

func (app *App) SendEmail(rc *RC, msg *Email) {
	if msg.From == "" {
		msg.From = app.Settings.EmailDefaultFrom
	}
	if msg.TextBody == "" && msg.HtmlBody == "" && msg.View != "" {
		if msg.Layout == "" {
			msg.Layout = app.Settings.EmailDefaultLayout
			if msg.Layout == "" {
				msg.Layout = "none"
			}
		}
		if !strings.HasPrefix(msg.View, "emails/") {
			panic("email view must start with emails/")
		}
		msg.HtmlBody = string(must(app.Render(rc, &ViewData{
			View:   msg.View,
			Layout: msg.Layout,
			Data:   msg.Data,
		})))
	}
	if msg.MessageStream == "" {
		msg.MessageStream = app.Settings.PostmarkDefaultMessageStream
	}
	pmsg := &postmark.Message{
		From:          msg.From,
		To:            msg.To,
		ReplyTo:       msg.ReplyTo,
		Cc:            msg.Cc,
		Bcc:           msg.Bcc,
		Subject:       msg.Subject,
		Tag:           msg.Category,
		TextBody:      msg.TextBody,
		HtmlBody:      msg.HtmlBody,
		MessageStream: msg.MessageStream,
		TemplateId:    msg.ServerTemplateIntID,
		TemplateModel: msg.ServerTemplateData,
	}
	app.postmrk.Send(rc, pmsg)
}
