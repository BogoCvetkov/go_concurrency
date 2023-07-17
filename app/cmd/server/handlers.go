package main

import (
	"fmt"
	"html/template"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/bogo/go-concurrency/data"
)

var partials = []string{
	"./templates/base.layout.gohtml",
	"./templates/header.partial.gohtml",
	"./templates/navbar.partial.gohtml",
	"./templates/footer.partial.gohtml",
	"./templates/alerts.partial.gohtml",
}

type TemplateData struct {
	Flash         string
	Warning       string
	Error         string
	Authenticated bool
	Now           time.Time
	User          *data.User
	Data          any
}

func (app *AppConfig) testing(w http.ResponseWriter, r *http.Request) {

	msg := MessageData{
		to:      "recipient@example.com",
		subject: "Testing mail concurrency",
		body:    "This should work bro",
	}

	app.Mailer.MailChan <- msg

	w.Write([]byte("Email Send"))
}

func (app *AppConfig) getHomePage(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.ParseFiles(append(partials, "./templates/home.page.gohtml")...)

	td := &TemplateData{
		Authenticated: false,
		Now:           time.Now(),
	}

	if err != nil {
		app.ErrLog.Println("Failed parsing homepage template")
		http.Error(w, "Internal Server Error", 500)
		return
	}

	if err = tmpl.ExecuteTemplate(w, "base", app.addDefaultData(td, r)); err != nil {
		app.ErrLog.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
	}
}

func (app *AppConfig) viewLogin(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.ParseFiles(append(partials, "./templates/login.page.gohtml")...)

	td := &TemplateData{
		Authenticated: false,
		Now:           time.Now(),
	}

	if err != nil {
		app.ErrLog.Println("Failed parsing homepage template")
		http.Error(w, "Internal Server Error", 500)
		return
	}

	if err = tmpl.ExecuteTemplate(w, "base", app.addDefaultData(td, r)); err != nil {
		app.ErrLog.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
	}
}

func (app *AppConfig) viewRegister(w http.ResponseWriter, r *http.Request) {

	tmpl, err := template.ParseFiles(append(partials, "./templates/register.page.gohtml")...)

	td := &TemplateData{
		Authenticated: false,
		Now:           time.Now(),
	}

	if err != nil {
		app.ErrLog.Println("Failed parsing homepage template")
		http.Error(w, "Internal Server Error", 500)
		return
	}

	if err = tmpl.ExecuteTemplate(w, "base", app.addDefaultData(td, r)); err != nil {
		app.ErrLog.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
	}
}

func (app *AppConfig) postRegister(w http.ResponseWriter, r *http.Request) {

	err := r.ParseForm()

	if err != nil {
		http.Error(w, "Failed to parse form data", 400)
		return
	}

	// Parse Data
	formData := r.PostForm

	user := data.User{
		Email:     formData.Get("email"),
		Password:  formData.Get("password"),
		FirstName: formData.Get("first-name"),
		LastName:  formData.Get("last-name"),
		Active:    0,
		IsAdmin:   0,
	}

	// Insert to DB
	if _, err = app.Models.User.Insert(user); err != nil {
		http.Error(w, "Failed insert user in DB", 400)
		return
	}

	// Sign url
	NewURLSigner()
	link := GenerateTokenFromString(fmt.Sprintf("http://localhost:3000/activate?email=%s", user.Email))

	email := MessageData{
		to:      user.Email,
		subject: "Registration",
		tmpl:    "./templates/confirmation.email.gohtml",
		dataMap: map[string]any{
			"name": fmt.Sprintf("%s %s", user.FirstName, user.LastName),
			"link": link,
		},
	}

	// Send to email workers
	app.Mailer.MailChan <- email

	app.Session.Put(r.Context(), "flash", "Register success")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *AppConfig) getActivate(w http.ResponseWriter, r *http.Request) {

	url := r.RequestURI
	urlParams := r.URL.Query()
	hash := fmt.Sprintf("http://localhost%s", url)

	isValid := VerifyToken(hash)
	expired := Expired(hash, 4)

	fmt.Println(isValid, expired, hash)

	if isValid && !expired {
		user, err := app.Models.User.GetByEmail(urlParams.Get("email"))
		if err != nil {
			app.Session.Put(r.Context(), "error", "user not found")
		} else {
			user.Active = 1
			user.Update()
			app.Session.Put(r.Context(), "flash", "Activation success")
		}
	} else {
		app.Session.Put(r.Context(), "error", "Invalid activation token")
	}
}

func (app *AppConfig) logout(w http.ResponseWriter, r *http.Request) {
	_ = app.Session.Destroy(r.Context())
	_ = app.Session.RenewToken(r.Context())
	app.Session.PopString(r.Context(), "userID")

	http.Redirect(w, r, "/login", 302)

}

func (app *AppConfig) login(w http.ResponseWriter, r *http.Request) {
	_ = app.Session.RenewToken(r.Context())
	err := r.ParseForm()

	if err != nil {
		app.ErrLog.Println(w, "Failed to parse form data", 400)
		return
	}

	data := r.PostForm

	user, err := app.Models.User.GetByEmail(data.Get("email"))
	if err != nil {
		app.Session.Pop(r.Context(), "error")
		app.Session.Put(r.Context(), "error", "Invalid email/password")
		notifyFailedLogin(data.Get("email"), app.Mailer.MailChan)
		http.Redirect(w, r, "/login", 302)
		return
	}

	if _, err := user.PasswordMatches(data.Get("password")); err != nil {
		app.Session.Pop(r.Context(), "error")
		app.Session.Put(r.Context(), "error", "Invalid email/password")
		// notifyFailedLogin(data.Get("email"), app.Mailer.MailChan)
		http.Redirect(w, r, "/login", 302)
		return
	}

	app.Session.Pop(r.Context(), "error")
	app.Session.Put(r.Context(), "userID", user.ID)
	app.Session.Put(r.Context(), "user", user)
	app.Session.Put(r.Context(), "flash", "Login success")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (app *AppConfig) chooseSubscription(w http.ResponseWriter, r *http.Request) {
	if !app.Session.Exists(r.Context(), "userID") {
		app.Session.Put(r.Context(), "warning", "You must log in to see this page!")
		http.Redirect(w, r, "/login", http.StatusTemporaryRedirect)
		return
	}

	plans, err := app.Models.Plan.GetAll()

	if err != nil {
		app.ErrLog.Println(err)
		return
	}

	tmpl, err := template.ParseFiles(append(partials, "./templates/plans.page.gohtml")...)

	td := &TemplateData{
		Authenticated: false,
		Now:           time.Now(),
		Data:          plans,
	}

	if err != nil {
		app.ErrLog.Println("Failed parsing plans template")
		http.Error(w, "Internal Server Error", 500)
		return
	}

	if err = tmpl.ExecuteTemplate(w, "base", app.addDefaultData(td, r)); err != nil {
		app.ErrLog.Println(err.Error())
		http.Error(w, "Internal Server Error", 500)
	}
}

func (app *AppConfig) addDefaultData(td *TemplateData, r *http.Request) *TemplateData {
	td.Flash = app.Session.PopString(r.Context(), "flash")
	td.Warning = app.Session.PopString(r.Context(), "warning")
	td.Error = app.Session.PopString(r.Context(), "error")
	if app.Session.Exists(r.Context(), "userID") {
		td.Authenticated = true
		user, ok := app.Session.Get(r.Context(), "user").(data.User)
		if !ok {
			app.ErrLog.Println("Can't get user from session")
		}

		td.User = &user
	}

	return td
}

func (app *AppConfig) subscribe(w http.ResponseWriter, r *http.Request) {

	q := r.URL.Query().Get("id")

	id, err := strconv.Atoi(q)

	if err != nil {
		app.ErrLog.Println("Failed to get plan id")
		http.Redirect(w, r, "/member/plans", 302)
		return
	}

	plan, err := app.Models.Plan.GetOne(id)

	if err != nil {
		app.ErrLog.Println("Failed to get plan")
		http.Redirect(w, r, "/member/plans", 302)
		return
	}

	user, ok := app.Session.Get(r.Context(), "user").(data.User)

	if !ok {
		app.ErrLog.Println("Failed to get user", err)
		http.Redirect(w, r, "/login", 302)
		return
	}

	err = plan.SubscribeUserToPlan(user, *plan)

	if err != nil {
		app.ErrLog.Println("Failed to subscribe to plan")
		http.Redirect(w, r, "/member/plans", 302)
		return
	}

	app.ShutDownWG.Add(2)
	go app.handleInvoice(plan, &user)
	go app.handleManual(plan, &user)

	w.Write([]byte("Email with PDF generated"))
}

func notifyFailedLogin(to string, ch chan<- MessageData) {
	msg := MessageData{
		to:      to,
		subject: fmt.Sprintf("%s failed to login", to),
		body:    fmt.Sprintf("%s failed to login", to),
		tmpl:    "./templates/failed_login.email.gohtml",
		dataMap: map[string]any{"email": to},
	}

	ch <- msg
}

func (app *AppConfig) handleInvoice(p *data.Plan, u *data.User) {
	iContent := fmt.Sprintf("You have bought a %s. You were charged %d.", p.PlanName, p.PlanAmount)
	file := fmt.Sprintf("../../subscriptions/subscription_%d.pdf", u.ID)
	generatePDf(iContent, file, app.ShutDownWG)

	msg := MessageData{
		to:         u.Email,
		subject:    "Invoice",
		dataMap:    map[string]any{"content": "Thank you for buying from us. You'll find the invoice attached"},
		tmpl:       "./templates/plan.email.gohtml",
		attachment: file,
	}

	app.Mailer.MailChan <- msg

}

func (app *AppConfig) handleManual(p *data.Plan, u *data.User) {
	iContent := strings.Repeat(fmt.Sprintf("Manual for using plan %s. \n", p.PlanName), 10)
	file := fmt.Sprintf("../../manuals/manual_%d.pdf", u.ID)

	generatePDf(iContent, file, app.ShutDownWG)

	msg := MessageData{
		to:         u.Email,
		subject:    "Manual",
		dataMap:    map[string]any{"content": "Thank you for buying from us. You'll find the manual attached"},
		tmpl:       "./templates/plan.email.gohtml",
		attachment: file,
	}

	app.Mailer.MailChan <- msg

}
