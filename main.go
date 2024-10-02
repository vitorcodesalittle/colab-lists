package main

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/mail"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	migrate "vilmasoftware.com/colablists/cmd"
	recovery "vilmasoftware.com/colablists/pkg"
	"vilmasoftware.com/colablists/pkg/community"
	"vilmasoftware.com/colablists/pkg/config"
	"vilmasoftware.com/colablists/pkg/list"
	"vilmasoftware.com/colablists/pkg/realtime"
	"vilmasoftware.com/colablists/pkg/session"
	"vilmasoftware.com/colablists/pkg/user"
	"vilmasoftware.com/colablists/pkg/views"
)

var (
	listsRepository     list.ListsRepository       = &list.SqlListRepository{}
	usersRepository     user.UsersRepository       = &user.SqlUsersRepository{}
	communityRepository *community.HouseRepository = &community.HouseRepository{}
)

var (
	liveEditor *realtime.LiveEditor = realtime.NewLiveEditor(listsRepository)
	upgrader                        = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
	}
	recoveryService *recovery.Recovery = &recovery.Recovery{UserRepository: usersRepository}
)

func getIndexHandler(w http.ResponseWriter, r *http.Request) {
	dir := http.Dir("./static")
	handler := http.FileServer(dir)
	accept := r.Header.Get("Accept")
	if strings.HasPrefix(accept, "text/html") {
		if redirectIfNotLoggedIn(w, r) {
			return
		}
		http.Redirect(w, r, "/lists", http.StatusSeeOther)
	} else {
		handler.ServeHTTP(w, r)
	}
}

func getLoginHandler(w http.ResponseWriter, r *http.Request) {
	views.Templates.RenderLogin(w, &views.SignupArgs{FormError: r.URL.Query().Get("formError")})
}

type Session struct {
	user.User
	SessionId string
	LastUsed  time.Time
}

func postLoginHandler(w http.ResponseWriter, r *http.Request) {
	user, err := usersRepository.UnsafeGetByUsername(r.FormValue("username"))
	if err == sql.ErrNoRows {
		http.Redirect(w, r, "/login?formError="+"No user named "+r.FormValue("username"), http.StatusSeeOther)
		return
	}
	if err != nil {
		http.Redirect(w, r, "/login?formError="+err.Error(), http.StatusSeeOther)
		return
	}
	if usersRepository.ComparePassword([]byte(r.FormValue("password")), []byte(user.PasswordHash)) {
		sessionId := session.GetSessionId()
		user.PasswordHash = ""
		if err := session.SaveSessionInDb(&session.Session{
			User:      user,
			SessionId: sessionId,
			LastUsed:  time.Now(),
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}

		w.Header().Add("Set-Cookie", "SESSION="+sessionId)
		http.Redirect(w, r, "/lists", http.StatusSeeOther)
	} else {
		http.Redirect(w, r, "/login?formError="+"Unauthorized", http.StatusSeeOther)
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
	_, err := usersRepository.CreateUser(r.FormValue("username"), r.FormValue("password"), r.FormValue("email"))
	if err != nil {
		http.Redirect(w, r, "/signup?formError="+err.Error(), http.StatusSeeOther)
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

	communities, err := communityRepository.FindMyHouses(user.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	views.Templates.RenderLists(w, &views.ListsArgs{
		Lists: lists,
		Form: views.ListCreationForm{
			Communities:      communities,
			DefaultCommunity: community.GetDefault(communities),
		},
		New: r.URL.Query().Has("new"),
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

	user, err := session.GetUserFromSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	listArgs := &views.ListArgs{
		List:    *views.NewListUi(&list, user),
		Editing: r.URL.Query().Has("edit"),
		IsDirty: false,
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

func getUsersHandler(w http.ResponseWriter, r *http.Request) {
	// Get id from path parameter
	query := r.URL.Query().Get("q")
	users, err := usersRepository.Search(query)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	err = json.NewEncoder(w).Encode(users)
	if err != nil {
		log.Println("Error encoding users")
		log.Println(err)
	}
}

func postListsHandler(w http.ResponseWriter, r *http.Request) {
	title := r.FormValue("title")
	description := r.FormValue("description")
	communityIdString := r.FormValue("communityId")
	var communityId *int64
	if len(communityIdString) > 0 {
		communityIdInt, err := strconv.ParseInt(communityIdString, 10, 64)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		communityId = &communityIdInt
	}
	user, err := session.GetUserFromSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
	}
	_, err = listsRepository.Create(&list.ListCreationParams{
		Title:       title,
		Description: description,
		CreatorId:   user.Id,
		CommunityId: communityId,
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

type UpdateListParams struct {
	Title       *string   `json:"title"`
	Description *string   `json:"description"`
	Members     *[]string `json:"members"`
}

func putListHandler(w http.ResponseWriter, r *http.Request) {
	listId, err := strconv.ParseInt(r.PathValue("listId"), 10, 64)
	if err != nil {
		http.Error(w, "listId path value should be integer", http.StatusBadRequest)
	}
	list, err := listsRepository.Get(listId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}

	var params UpdateListParams
	json.NewDecoder(r.Body).Decode(&params)
	if params.Title != nil {
		list.Title = *params.Title
	}
	if params.Description != nil {
		list.Description = *params.Description
	}
	if params.Members != nil {
		list.Colaborators = []user.User{}
		for _, colaborator := range *params.Members {
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
	}
	listv, err := listsRepository.Update(&list)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	listState := liveEditor.GetCurrentListState(listId)
	if listState != nil {
		listState.Ui.List = listv
	}

	w.Header().Add("HX-Redirect", fmt.Sprintf("/lists/%d", listId))
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
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
	liveEditor.SetupList(int64(listId), user, conn)
}

func getSignupHandler(w http.ResponseWriter, r *http.Request) {
	views.Templates.RenderSignup(w, &views.SignupArgs{FormError: r.URL.Query().Get("formError")})
}

func getCommunitiesHandler(w http.ResponseWriter, r *http.Request) {
	if redirectIfNotLoggedIn(w, r) {
		return
	}
	comunityPageArgs := &views.CommunitiesArgs{}
	query, err := getCommunitiesQuery(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	comunityPageArgs.Query = *query
	if query.SelectedId > 0 && query.EditingId > 0 {
		http.Error(w, "Cannot select and edit at the same time", http.StatusBadRequest)
	} else if query.SelectedId > 0 || query.EditingId > 0 {
		comunityPageArgs.SelectedCommunity, err = communityRepository.Get(query.SelectedId + query.EditingId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
	}
	user, err := session.GetUserFromSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	comunityPageArgs.Communities, err = communityRepository.FindMyHouses(user.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	views.Templates.RenderCommunities(w, comunityPageArgs)
}

func getCommunitiesQuery(r *http.Request) (*views.CommunitiesQuery, error) {
	var err error
	result := &views.CommunitiesQuery{}

	selectedIdString := r.URL.Query().Get("selectedId")
	if selectedIdString != "" {
		result.SelectedId, err = strconv.ParseInt(selectedIdString, 10, 64)
		if err != nil {
			return nil, err
		}
	}
	newString := r.URL.Query().Get("new")
	if newString != "" {
		result.New, err = strconv.ParseBool(newString)
		if err != nil {
			return nil, err
		}
	}

	editingIdString := r.URL.Query().Get("editingId")
	if editingIdString != "" {
		result.EditingId, err = strconv.ParseInt(editingIdString, 10, 64)
		if err != nil {
			return nil, err
		}
	}
	return result, nil
}

var allconns []*websocket.Conn = make([]*websocket.Conn, 0)

func getHotReloadHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// Return Internal Error
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	allconns = append(allconns, conn)
	go handleHotReload(conn)
}

type HotReloadMessage struct {
	ServerRunId string `json:"serverRunId"`
}

var ServerRunID string

func removeHotReloadCon(conn *websocket.Conn) {
	for i, c := range allconns {
		if c == conn {
			allconns = append(allconns[:i], allconns[i+1:]...)
			return
		}
	}
}

func handleHotReload(conn *websocket.Conn) {
	defer conn.Close()
	for {
		hotReloadMessage := &HotReloadMessage{}
		err := conn.ReadJSON(hotReloadMessage)
		if err != nil {
			removeHotReloadCon(conn)
			return
		}
		conn.WriteJSON(&HotReloadMessage{ServerRunId: ServerRunID})
	}
}

func postCommunitiesHandler(w http.ResponseWriter, r *http.Request) {
	user, err := session.GetUserFromSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	community := &community.Community{}
	mountCommunityFromRequest(r, user, community)
	community.CreatedBy = user
	_, err = communityRepository.Save(community, nil)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Add("HX-Redirect", "/communities")
}

type CommunityCreationParams struct {
	CommunityName string   `json:"communityName"`
	MemberIds     []string `json:"members"`
}

func mountCommunityFromRequest(r *http.Request, u *user.User, c *community.Community) error {
	params := &CommunityCreationParams{}
	err := json.NewDecoder(r.Body).Decode(params)
	if err != nil {
		return err
	}
	c.CommunityName = params.CommunityName
	c.CreatedBy = u
	c.Members = make([]*community.Member, 0)
	c.Default = false
	for _, memberId := range params.MemberIds {
		intMemberId, err := strconv.Atoi(memberId)
		if err != nil {
			return err
		}
		member, err := usersRepository.Get(int64(intMemberId))
		if err != nil {
			return err
		}
		c.Members = append(c.Members, &community.Member{User: member})
	}
	return nil
}

func putCommunitiesHandler(w http.ResponseWriter, r *http.Request) {
	user, err := session.GetUserFromSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	community := &community.Community{}
	mountCommunityFromRequest(r, user, community)
	community.UpdatedAt = time.Now()
	communityId, err := strconv.Atoi(r.PathValue("communityId"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	community.CommunityId = int64(communityId)
	_, err = communityRepository.Save(community, &user.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Add("HX-Redirect", fmt.Sprintf("/communities?selectedId=%d", community.CommunityId))
}

func deleteCommunitiesHandler(w http.ResponseWriter, r *http.Request) {
	user, err := session.GetUserFromSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	communityId, err := strconv.ParseInt(r.PathValue("communityId"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = communityRepository.Delete(communityId, user.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Add("HX-Redirect", "/communities")
}

func getPasswordRecoveryHandler(w http.ResponseWriter, r *http.Request) {
	views.Templates.RenderPasswordRecovery(w, &views.PasswordRecoveryArgs{})
}

func postPasswordRecoveryHandler(w http.ResponseWriter, r *http.Request) {
	password := r.FormValue("password")
	token := r.FormValue("token")
	if err := recoveryService.RecoverPassword(password, token); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func getPasswordRecoveryRequestHandler(w http.ResponseWriter, r *http.Request) {
	views.Templates.RenderPasswordRecoveryRequest(w, &views.PasswordRecoveryRequestArgs{})
}

func postPasswordRecoveryRequestHandler(w http.ResponseWriter, r *http.Request) {
	email := r.FormValue("email")
	if _, err := mail.ParseAddress(email); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := recoveryService.CreatePasswordRecoveryRequest(email); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/login", http.StatusSeeOther)
}

func deleteListHandler(w http.ResponseWriter, r *http.Request) {
	if redirectIfNotLoggedIn(w, r) {
		return
	}
	lisetId := r.PathValue("listId")
	id, err := strconv.ParseInt(lisetId, 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := session.GetUserFromSession(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	err = listsRepository.Delete(id, user.Id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Add("HX-Redirect", "/lists")
}

func main() {
	ServerRunID = fmt.Sprintf("%x", sha256.New().Sum([]byte(time.Now().String())))
	config := config.GetConfig()

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
	//
	http.HandleFunc("GET /login", getLoginHandler)
	http.HandleFunc("POST /login", postLoginHandler)
	http.HandleFunc("GET /logout", getLogoutHandler)
	http.HandleFunc("GET /signup", getSignupHandler)
	http.HandleFunc("POST /sign-up", postSignupHandler)
	http.HandleFunc("GET /", getIndexHandler)
	http.HandleFunc("GET /lists", getListsHandler)
	http.HandleFunc("POST /lists", postListsHandler)
	http.HandleFunc("GET /lists/{listId}", getListDetailHandler)
	http.HandleFunc("DELETE /lists/{listId}", deleteListHandler)
	http.HandleFunc("GET /api/users/{userId}", getUserHandler)
	http.HandleFunc("GET /api/users", getUsersHandler)
	http.HandleFunc("GET /ws/list-editor", getListEditorHandler)
	http.HandleFunc("PUT /lists/{listId}/save", putListSaveHandler)
	http.HandleFunc("PUT /lists/{listId}", putListHandler)
	http.HandleFunc("GET /communities", getCommunitiesHandler)
	http.HandleFunc("POST /communities", postCommunitiesHandler)
	http.HandleFunc("PUT /communities/{communityId}", putCommunitiesHandler)
	http.HandleFunc("DELETE /communities/{communityId}", deleteCommunitiesHandler)
	http.HandleFunc("GET /password-recovery", getPasswordRecoveryHandler)
	http.HandleFunc("POST /password-recovery", postPasswordRecoveryHandler)
	http.HandleFunc("GET /password-recovery-request", getPasswordRecoveryRequestHandler)
	http.HandleFunc("POST /password-recovery-request", postPasswordRecoveryRequestHandler)

	// For development purposes only:
	if config.HotReload {
		http.HandleFunc("GET /ws/hot-reload", getHotReloadHandler)
	}

	log.Printf("Server started at %s\n", config.Listen)
	httpServer := http.Server{
		Addr:              config.Listen,
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

	if config.UseTls {
		if err := httpServer.ListenAndServeTLS(config.Certificate, config.PrivateKey); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	} else {
		if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}

	err = session.SaveSessionsInDb()
	if err != nil {
		log.Println("Failed to save current sessions map to DB")
		log.Fatal(err)
	}
	log.Println("Server stopped!")
}
