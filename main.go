package main

import (
	"crypto/rand"
	"log"
	"net/http"
	"strconv"
	tmpl "text/template"
	"time"

	"vilmasoftware.com/colablists/pkg/list"
	"vilmasoftware.com/colablists/pkg/user"
)

var templatesMap map[string]*tmpl.Template

var (
	listsRepository list.ListsRepository
	usersRepository user.UsersRepository
)

type Args struct {
	Title       string
	Description string
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	templatesMap["index"].Execute(w, &Args{
		Title:       "Lists app!!",
		Description: "Awesome lists app",
	})
}

type LoginArgs struct{}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	templatesMap["login"].Execute(w, &LoginArgs{})
}

type Session struct {
  user.User
  SessionId string
  LastUsed time.Time
}

var sessionsMap map[string]Session;

func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	// Note that err == nil only if we read len(b) bytes.
	if err != nil {
		return nil, err
	}

	return b, nil
}

func getSessionId() string {
  for true {
    sessionIdBytes, err := GenerateRandomBytes(12)
    if err != nil { panic(err) }
    sessionId := string(sessionIdBytes)
    if _, ok := sessionsMap[sessionId]; !ok {
      return sessionId
    }
  }
  return ""
}

func getUserFromSession(r *http.Request) (user.User, error) {
  sessionId := r.Header.Get("Cookie")
  session, ok := sessionsMap[sessionId]
  if !ok {
    return user.User{}, nil
  }
  return session.User, nil
}

func postLoginHandler(w http.ResponseWriter, r *http.Request) { user, err := usersRepository.GetByUsername(r.FormValue("username"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
  if usersRepository.ComparePassword([]byte(r.FormValue("password")), []byte(user.PasswordHash)) {
    sessionId := getSessionId()
    sessionsMap[sessionId] = Session{
      User: user,
      SessionId: sessionId,
      LastUsed: time.Now(),
    }
    w.Header().Add("Set-Cookie", "SESSION="+sessionId)
  } else {
		http.Error(w, "", http.StatusUnauthorized)
  }
}

func postLogoutHandler(w http.ResponseWriter, r *http.Request) {
  cookie, err := r.Cookie("SESSION")
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
  delete(sessionsMap, cookie.Value)
  w.Header().Add("Set-Cookie", "SESSION=; expires=Thu, 01 Jan 1970 00:00:00 GMT")
}

func postSignupHandler(w http.ResponseWriter, r *http.Request) {
  _, err := usersRepository.CreateUser(r.FormValue("username"), r.FormValue("password"))
  if err != nil {
    log.Println(err)
    http.Error(w, err.Error(), http.StatusInternalServerError)
  } else {
    http.Redirect(w, r, "/login", http.StatusSeeOther)
  }
}


type ListsArgs struct {
	Lists []list.List
}

func listsHandler(w http.ResponseWriter, r *http.Request) {
	lists, err := listsRepository.GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	templatesMap["lists"].Execute(w, &ListsArgs{
		Lists: lists,
	})
}

type ListArgs struct {
	List         list.List
	Editing      bool
	Colaborators []user.User
}

func listDetailHandler(w http.ResponseWriter, r *http.Request) {
	// Get id from path parameter
	id, err := strconv.Atoi(r.PathValue("listId"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	list, err := listsRepository.Get(int64(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	colaborators, err := usersRepository.GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	templatesMap["list"].Execute(w, &ListArgs{
		List:         list,
		Editing:      r.URL.Query().Has("edit"),
		Colaborators: colaborators,
	})
}

func collectUsers(lists []list.List) []user.User {
	var users []user.User
	for _, list := range lists {
		for _, colaborator := range list.Colaborators {
			users = append(users, colaborator)
		}
	}
	return users
}

func main() {
	// Fill templatesMap with all templates
	templatesMap = make(map[string]*tmpl.Template)
	templatesMap["index"] = tmpl.Must(
		tmpl.ParseFiles("./templates/pages/index.html"),
	)
	templatesMap["login"] = tmpl.Must(
		tmpl.ParseFiles("./templates/pages/login.html"),
	)
	templatesMap["lists"] = tmpl.Must(
		tmpl.ParseFiles("./templates/pages/lists.html"),
	)
	templatesMap["list"] = tmpl.Must(
		tmpl.ParseFiles("./templates/pages/list.html"),
	)
	// Create a new in-memory repository
	listsRepository = &list.SqlListRepository{}

	_, err := listsRepository.GetAll()
	if err != nil {
		panic(err)
	}
	usersRepository = &user.SqlUsersRepository{}

	dir := http.Dir(".")
	http.Handle("GET /static/", http.FileServer(dir))
	http.HandleFunc("GET /login", loginHandler)
	http.HandleFunc("POST /login", postLoginHandler)
	http.HandleFunc("POST /logout", postLogoutHandler)
	http.HandleFunc("POST /sign-up", postSignupHandler)
	http.HandleFunc("GET /", indexHandler)
	http.HandleFunc("GET /lists", listsHandler)
	http.HandleFunc("GET /lists/{listId}", listDetailHandler)

	log.Printf("Server started at http://localhost:8080\n")
	log.Fatal(http.ListenAndServe(":8080", nil))

	log.Println("Shutting server down...")
}
