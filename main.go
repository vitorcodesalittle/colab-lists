package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	migrate "vilmasoftware.com/colablists/cmd"
	"vilmasoftware.com/colablists/pkg/list"
	"vilmasoftware.com/colablists/pkg/realtime"
	"vilmasoftware.com/colablists/pkg/session"
	"vilmasoftware.com/colablists/pkg/user"
	"vilmasoftware.com/colablists/pkg/views"
)

var (
	listsRepository list.ListsRepository = &list.SqlListRepository{}
	usersRepository user.UsersRepository = &user.SqlUsersRepository{}
)

var (
	liveEditor *realtime.LiveEditor = realtime.NewLiveEditor(listsRepository)
	upgrader                        = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
)

func getIndexHandler(w http.ResponseWriter, r *http.Request) {
    log.Printf("Headers: %v\n", r.Header)
    //if redirectIfNotLoggedIn(w, r) {
    //    return
    //}
    // http.Redirect(w, r, "/lists", http.StatusSeeOther)
    views.Templates.RenderIndex(w, &views.IndexArgs{
        Title: "Lists!",
        Description: "Website for lists",
    })
}

func getLoginHandler(w http.ResponseWriter, r *http.Request) {
	views.Templates.RenderLogin(w)
}

type Session struct {
	user.User
	SessionId string
	LastUsed  time.Time
}

func postLoginHandler(w http.ResponseWriter, r *http.Request) {
	user, err := usersRepository.GetByUsername(r.FormValue("username"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	if usersRepository.ComparePassword([]byte(r.FormValue("password")), []byte(user.PasswordHash)) {
		sessionId := session.GetSessionId()
		session.SessionsMap[sessionId] = &session.Session{
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

func getLogoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("SESSION")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	delete(session.SessionsMap, cookie.Value)
	w.Header().Add("Set-Cookie", "SESSION=; expires=Thu, 01 Jan 1970 00:00:00 GMT")
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func postSignupHandler(w http.ResponseWriter, r *http.Request) {
	_, err := usersRepository.CreateUser(r.FormValue("username"), r.FormValue("password"))
	if err != nil {
		http.Redirect(w, r, "/signup?error=" + err.Error(), http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
	}
}

func getListsHandler(w http.ResponseWriter, r *http.Request) {
	if redirectIfNotLoggedIn(w, r) {
		return
	}
    user, err := session.GetUserFromSession(r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusUnauthorized)
        return
    }
	lists, err := listsRepository.GetAll(user.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	views.Templates.RenderLists(w, &views.ListsArgs{
		Lists: lists,
	})
}

func redirectIfNotLoggedIn(w http.ResponseWriter, r *http.Request) bool {
	_, err := session.GetUserFromSession(r)
	if err != nil {
		// Redirect to login
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return true
	}
	return false
}

func getListDetailHandler(w http.ResponseWriter, r *http.Request) {
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

    user, err := session.GetUserFromSession(r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusUnauthorized)
        return
    }

	listArgs := &views.ListArgs{
		List:     *views.NewListUi(&list, user ),
		Editing:  r.URL.Query().Has("edit"),
		AllUsers: allUsers,
		IsDirty:  false,
	}
	list2 := liveEditor.GetCurrentListState(int64(id))
	if list2 != nil {
		listArgs.List = *list2.Ui
		listArgs.IsDirty = list2.Dirty
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
	user, err := session.GetUserFromSession(r)
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

func putListSaveHandler(w http.ResponseWriter, r *http.Request) {
	listId, err := strconv.ParseInt(r.PathValue("listId"), 10, 64)
	if err != nil {
		http.Error(w, "listId path value should be integer", http.StatusBadRequest)
	}
	found := liveEditor.GetCurrentListUi(listId)
	if found == nil {
		http.Error(w, "List not found", http.StatusNotFound)
		return
	}
	list, err := listsRepository.Update(found.List)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
    user, err := session.GetUserFromSession(r)
    if err != nil {
        http.Error(w, err.Error(), http.StatusUnauthorized)
        return
    }
	// TODO: send changes to all users
	views.Templates.RenderSaveList(w, &views.ListArgs{List: *views.NewListUi(list, user), IsDirty: false})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func putListHandler(w http.ResponseWriter, r *http.Request) {
	listId, err := strconv.ParseInt(r.PathValue("listId"), 10, 64)
	if err != nil {
		http.Error(w, "listId path value should be integer", http.StatusBadRequest)
	}
	list, err := listsRepository.Get(listId)
	formBody, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	urlParsed, err := url.Parse("http://localhost:8080?" + string(formBody))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	colaborators := urlParsed.Query()["colaborators"]
	list.Title = urlParsed.Query().Get("title")
	list.Description = urlParsed.Query().Get("description")
    list.Colaborators = []user.User{}
    for _, colaborator := range colaborators {
        colaboratorId, err := strconv.Atoi(colaborator)
        if err != nil {
            http.Error(w, err.Error(), http.StatusBadRequest)
        }
        user, err := usersRepository.Get(int64(colaboratorId))
        if err != nil {
            http.Error(w, err.Error(), http.StatusInternalServerError)
            return
        }
        list.Colaborators = append(list.Colaborators, user)
    }
    fmt.Printf("Salvando lista %v\n", list)
	listv, err := listsRepository.Update(&list)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	liveEditor.GetCurrentListState(listId).Ui.List = listv

	w.Header().Add("HX-Redirect", fmt.Sprintf("/lists/%d", listId))
	// http.Redirect(w, r, "/lists", http.StatusSeeOther)
}

func getListEditorHandler(w http.ResponseWriter, r *http.Request) {
	user, err := session.GetUserFromSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
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

func getSignupHandler(w http.ResponseWriter, r *http.Request) {
    views.Templates.RenderSignup(w, &views.SignupArgs{Error: r.URL.Query().Get("error")})
}

func main() {
    log.Println("Starting migrations")
    result := migrate.MigrateDb()

    if result.Error != nil {
        log.Println("Failed to migrate DB")
        if result.MigrationError != "" {
            log.Println("Migration error: ", result.MigrationError)
        }
        log.Fatal(result.Error)
    }
    log.Printf("Migrations done: %v\n", result.RanMigrations)
    err := session.RestoreSessionsFromDb()
	if err != nil {
		log.Fatal(err)
	}

	dir := http.Dir(".")
	http.Handle("GET /static/", http.FileServer(dir))
	http.HandleFunc("GET /login", getLoginHandler)
	http.HandleFunc("POST /login", postLoginHandler)
	http.HandleFunc("GET /logout", getLogoutHandler)
	http.HandleFunc("GET /signup", getSignupHandler)
	http.HandleFunc("POST /sign-up", postSignupHandler)
	http.HandleFunc("GET /", getIndexHandler)
	http.HandleFunc("GET /lists", getListsHandler)
	http.HandleFunc("POST /lists", postListsHandler)
	http.HandleFunc("GET /lists/{listId}", getListDetailHandler)
	http.HandleFunc("GET /api/users/{userId}", getUserHandler)
	http.HandleFunc("GET /ws/list-editor", getListEditorHandler)
	http.HandleFunc("PUT /lists/{listId}/save", putListSaveHandler)
	http.HandleFunc("PUT /lists/{listId}", putListHandler)


	listenAddress := flag.String("listen", ":8080", "Listen address.")
    flag.Parse()
	log.Printf("Server started at %s\n", *listenAddress)
	httpServer := http.Server{
		Addr:              *listenAddress,
		ReadHeaderTimeout: 3 * time.Second,
		ReadTimeout:       5 * time.Second,
		IdleTimeout:       10 * time.Second,
		WriteTimeout:      10 * time.Second,
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
		log.Fatal(err)
	}

	err = session.SaveSessionsInDb()
	if err != nil {
		log.Println("Failed to save current sessions map to DB")
        log.Fatal(err)
	}
}
