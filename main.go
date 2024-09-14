package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"flag"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"vilmasoftware.com/colablists/pkg/list"
	"vilmasoftware.com/colablists/pkg/user"
	"vilmasoftware.com/colablists/pkg/views"
)

var (
	listsRepository list.ListsRepository = &list.SqlListRepository{}
	usersRepository user.UsersRepository = &user.SqlUsersRepository{}
)

var (
	liveEditor *list.LiveEditor = list.NewLiveEditor(listsRepository)
	upgrader                    = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)


func indexHandler(w http.ResponseWriter, r *http.Request) {
	views.Templates.RenderIndex(w, &views.IndexArgs{
		Title:       "Lists app!!",
		Description: "Awesome lists app",
	})
}


func loginHandler(w http.ResponseWriter, r *http.Request) {
	views.Templates.RenderLogin(w, &views.LoginArgs{})
}

type Session struct {
	user.User
	SessionId string
	LastUsed  time.Time
}

var sessionsMap map[string]Session = make(map[string]Session)

func GenerateRandomBytes(n int) []byte {
	if n == 0 {
		panic("n must be greater than 0")
	}
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		panic(err)
	}
	return b
}

func getSessionId() string {
	for true {
		sessionIdBytes := base64.RawStdEncoding.EncodeToString(GenerateRandomBytes(128))
		sessionId := string(sessionIdBytes)
		if _, ok := sessionsMap[sessionId]; !ok {
			return sessionId
		}
	}
	return ""
}

func getUserFromSession(r *http.Request) (user.User, error) {
	sessionId, err := r.Cookie("SESSION")
	if err != nil {
		return user.User{}, err
	}
	session, ok := sessionsMap[sessionId.Value]
	if !ok {
		return user.User{}, errors.New("Session not found")
	}
	return session.User, nil
}

func postLoginHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("%v:%v", r.FormValue("username"), r.FormValue("password"))
	user, err := usersRepository.GetByUsername(r.FormValue("username"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if usersRepository.ComparePassword([]byte(r.FormValue("password")), []byte(user.PasswordHash)) {
		sessionId := getSessionId()
		sessionsMap[sessionId] = Session{
			User:      user,
			SessionId: sessionId,
			LastUsed:  time.Now(),
		}
		w.Header().Add("Set-Cookie", "SESSION="+sessionId)
		http.Redirect(w, r, "/lists", http.StatusSeeOther)
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
		log.Println("Error creating user")
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
	if redirectIfNotLoggedIn(w, r) {
		return
	}
	lists, err := listsRepository.GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	views.Templates.RenderLists(w, &ListsArgs{
		Lists: lists,
	})
}

type ListArgs struct {
	List         list.List
	Editing      bool
	AllUsers []user.User
}

func redirectIfNotLoggedIn(w http.ResponseWriter, r *http.Request) bool {
	_, err := getUserFromSession(r)
	if err != nil {
		// Redirect to login
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return true
	}
	return false
}

func listDetailHandler(w http.ResponseWriter, r *http.Request) {
	if redirectIfNotLoggedIn(w, r) {
		return
	}
	// Get id from path parameter
	id, err := strconv.Atoi(r.PathValue("listId"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	list, err := listsRepository.Get(int64(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	allUsers, err := usersRepository.GetAll()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
    listArgs := &ListArgs{
		List:         list,
		Editing:      r.URL.Query().Has("edit"),
		AllUsers: allUsers,
	}
	views.Templates.RenderList(w, listArgs)
}

func getUserHandler(w http.ResponseWriter, r *http.Request) {
	// Get id from path parameter
	id, err := strconv.Atoi(r.PathValue("userId"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	user, err := usersRepository.Get(int64(id))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	w.Write([]byte(user.Username))
}

func postListsHandler(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	description := r.FormValue("description")
	user, err := getUserFromSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
	}
	_, err = listsRepository.Create(&list.ListCreationParams{
		Title:       title,
		Description: description,
		CreatorId:   user.Id,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	http.Redirect(w, r, "/lists", http.StatusSeeOther)
}

func listEditorHandler(w http.ResponseWriter, r *http.Request) {
	user, err := getUserFromSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		// Return Internal Error
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	listId, err := strconv.Atoi(r.URL.Query().Get("listId"))
	liveEditor.SetupList(int64(listId), user, conn)
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
	_, err := listsRepository.GetAll()
	if err != nil {
		panic(err)
	}

	dir := http.Dir(".")
	http.Handle("GET /static/", http.FileServer(dir))
	http.HandleFunc("GET /login", loginHandler)
	http.HandleFunc("POST /login", postLoginHandler)
	http.HandleFunc("POST /logout", postLogoutHandler)
	http.HandleFunc("POST /sign-up", postSignupHandler)
	http.HandleFunc("GET /", indexHandler)
	http.HandleFunc("GET /lists", listsHandler)
	http.HandleFunc("POST /lists", postListsHandler)
	http.HandleFunc("GET /lists/{listId}", listDetailHandler)
	http.HandleFunc("GET /api/users/{userId}", getUserHandler)
	http.HandleFunc("GET /ws/list-editor", listEditorHandler)

	log.Printf("Server started at http://localhost:8080\n")


	var listenAddress = flag.String("listen", ":8080", "Listen address.")
	httpServer := http.Server{
		Addr: *listenAddress,
	}
	go func() {
		sigint := make(chan os.Signal, 1)
		signal.Notify(sigint, os.Interrupt)
		<-sigint
		if err := httpServer.Shutdown(context.Background()); err != nil {
			log.Printf("HTTP Server Shutdown Error: %v", err)
		}
	}()

    if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
        log.Println("Error")
        log.Fatal(err)
    }

    log.Println("Bye")
}
