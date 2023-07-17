package main

import (
	"bytes"
	"fmt"
	"html/template"
	"sync"
	"time"

	"gopkg.in/gomail.v2"
)

type Mailer struct {
	Host     string
	Port     int
	Sender   string
	Dialer   *gomail.Dialer
	MailChan chan MessageData
	DoneChan chan string
	ErrChan  chan error
}

type MessageData struct {
	to         string
	subject    string
	body       string
	tmpl       string
	dataMap    map[string]any
	attachment string
}

type Message struct {
	Message string
}

func (app *AppConfig) listenForEmails() {

	for {
		select {
		case email := <-app.Mailer.MailChan:
			app.ShutDownWG.Add(1)
			go app.Mailer.sendMessage(email, app.ShutDownWG)
		case val := <-app.Mailer.DoneChan:
			app.InfoLog.Printf(fmt.Sprintf("Email send to --> %s \n", val))
		case err := <-app.Mailer.ErrChan:
			app.ErrLog.Printf(fmt.Sprintf("Failed sending email to --> %s \n", err))

		}
	}

}

func (m *Mailer) initiateDialer() *Mailer {
	m.Dialer = &gomail.Dialer{
		Host: m.Host,
		Port: m.Port,
	}

	return m
}

func (m *Mailer) sendMessage(data MessageData, wg *sync.WaitGroup) {
	defer wg.Done()

	time.Sleep(time.Second * 1)
	msg := gomail.NewMessage()
	msg.SetHeader("From", m.Sender)
	msg.SetHeader("To", data.to)
	msg.SetHeader("Subject", data.subject)

	if data.attachment != "" {
		msg.Attach(data.attachment)
	}

	body, err := m.compileHTML(data)
	if err != nil {
		m.ErrChan <- err
		return
	}
	msg.SetBody("text/html", body)

	if err := m.Dialer.DialAndSend(msg); err != nil {
		m.ErrChan <- err
		return
	}

	m.DoneChan <- data.to
}

func (m *Mailer) compileHTML(data MessageData) (string, error) {

	tmpl, err := template.ParseFiles(append(partials, data.tmpl)...)

	var buffer bytes.Buffer

	if err = tmpl.ExecuteTemplate(&buffer, "body", data.dataMap); err != nil {

		return "", err
	}

	return buffer.String(), nil
}
