package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"html/template"
	"net/http"
	"os"

	e "github.com/ferretcode-freelancing/blog/error"
	"github.com/ferretcode-freelancing/blog/session"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type User struct {
	Username string `json:"username" gorm:"primaryKey"`
	Password string `json:"password"`
	Admin    bool   `json:"admin"`
}

func ShowSignIn(w http.ResponseWriter, r *http.Request) error {
	tmpl, err := template.ParseFiles("templates/auth/sign_in.html")

	if err != nil {
		return err
	}

	tmpl.Execute(w, nil)

	return nil
}

func ShowCreateAccount(w http.ResponseWriter, r *http.Request) error {
	tmpl, err := template.ParseFiles("templates/auth/create_account.html")

	if err != nil {
		return err
	}

	tmpl.Execute(w, nil)

	return nil
}

func CreateAccount(w http.ResponseWriter, r *http.Request, db *gorm.DB) error {
	user := User{
		Username: r.URL.Query().Get("username"),
		Password: r.URL.Query().Get("password"),
	}

	if user.Username == "" || user.Password == "" {
		err := e.RenderError(w, r, 400, "Both username and password must be present.")

		if err != nil {
			return err
		}

		return nil
	}

	hash := sha256.New()
	hash.Write([]byte(user.Password))
	password := hash.Sum(nil)

	user.Password = base64.URLEncoding.EncodeToString(password)

	err := db.Create(&user).Error

	if err != nil {
		return err
	}

	http.Redirect(w, r, "/", http.StatusPermanentRedirect)

	return nil
}

func SignIn(w http.ResponseWriter, r *http.Request, db *gorm.DB) error {
	cookie, err := r.Cookie("blog")

	if err != nil && err != http.ErrNoCookie {
		return err
	} else if err == nil {
		_, err := session.GetSession(cookie.Value)

		if err != nil && err != session.ErrNotAuthenticated {
			return err
		} else if err == session.ErrNotAuthenticated {
			http.Redirect(w, r, "/auth/sign_in", http.StatusTemporaryRedirect)
		}

		http.Redirect(w, r, "/", http.StatusPermanentRedirect)

		return nil
	}

	user := User{
		Username: r.URL.Query().Get("username"),
		Password: r.URL.Query().Get("password"),
	}

	if user.Username == "" || user.Password == "" {
		err = e.RenderError(w, r, 400, "Both username and password must be present.")

		if err != nil {
			return err
		}

		return nil
	}

	dbUser := User{}

	hash := sha256.New()
	hash.Write([]byte(user.Password))
	password := hash.Sum(nil)

	user.Password = base64.URLEncoding.EncodeToString(password)

	err = db.Where("username = ?", user.Username).First(&dbUser).Error

	if err != nil {
		return err
	}

	match := subtle.ConstantTimeCompare([]byte(user.Password), []byte(dbUser.Password))

	if match != 1 {
		w.WriteHeader(403)
		w.Write([]byte("The password was incorrect."))

		return nil
	}

	sid := uuid.NewString()

	c := http.Cookie{
		Name:   "blog",
		Value:  sid,
		Domain: os.Getenv("COOKIE_DOMAIN"),
		Path:   "/",
	}

	http.SetCookie(w, &c)

	err = session.CreateSession(sid, user.Username)

	if err != nil {
		return err
	}

	w.WriteHeader(200)
	w.Write([]byte("You were signed in successfully."))

	return nil
}
