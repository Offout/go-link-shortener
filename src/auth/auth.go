package auth

import (
	"encoding/json"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"net/http"
	"strings"
)

type registerForm struct {
	Username string
	Password string
}

type loginForm struct {
	Username string
	Password string
}

type loginResponse struct {
	AccessToken string `json:"accessToken"`
}

// userName => passwordhash
var users = make(map[string]string)

// sessionToken => userName
var sessions = make(map[string]string)

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	return string(bytes), err
}

func checkPasswordHash(password, hash string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}

func Register(w http.ResponseWriter, r *http.Request) {
	var form registerForm
	err := json.NewDecoder(r.Body).Decode(&form)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var passwordHash string
	passwordHash, err = hashPassword(form.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	_, ok := users[form.Username]
	if ok {
		http.Error(w, "User already registered", http.StatusConflict)
		return
	}
	users[form.Username] = passwordHash
}

func Login(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var loginForm = loginForm{r.PostFormValue("username"), r.PostFormValue("password")}
	passwordHash, ok := users[loginForm.Username]
	if !ok {
		http.Error(w, "No such user", http.StatusBadRequest)
		return
	}
	if !checkPasswordHash(loginForm.Password, passwordHash) {
		http.Error(w, "Wrong password", http.StatusBadRequest)
		return
	}
	sessionId := uuid.New()

	sessions[sessionId.String()] = loginForm.Username
	err = json.NewEncoder(w).Encode(loginResponse{sessionId.String()})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// CheckSession Returns UserName or empty string
func CheckSession(r *http.Request) string {
	var session = r.Header.Get("authorization")
	if !strings.HasPrefix(session, "Bearer ") {
		return ""
	}
	var userName, ok = sessions[session[7:]]
	if ok {
		return userName
	}
	return ""
}
