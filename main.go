package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	tmpl "text/template"

	lists "vilmasoftware.com/colablists/pkg"
)

var templatesMap map[string]*tmpl.Template

var listsRepository lists.ListsRepository;

type Args struct {
  Title string
  Description string
}
func indexHandler(w http.ResponseWriter, r *http.Request) {
  templatesMap["index"].Execute(w, &Args{
    Title:  "Lists app!!",
    Description: "Awesome lists app",
  })
}

type LoginArgs struct {
}
func loginHandler(w http.ResponseWriter, r *http.Request) {
  templatesMap["login"].Execute(w, &LoginArgs{})
}

type ListsArgs struct {
  Lists []lists.List
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
  List lists.List
  Editing bool
}
func listDetailHandler(w http.ResponseWriter, r *http.Request) {
  // Get id from path parameter
  id, err := strconv.Atoi(r.PathValue("listId"))
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
  list, err := listsRepository.Get(id)
  if err != nil {
    http.Error(w, err.Error(), http.StatusInternalServerError)
  }
  fmt.Printf("List: %v\n", list)
  templatesMap["list"].Execute(w, &ListArgs{
    List: list,
    Editing: r.URL.Query().Has("edit"),
  })
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
  listsRepository = lists.NewListsInMemoryRepository([]lists.List{
    lists.MockList(),
    lists.MockList(),
    lists.MockList(),
    lists.MockList(),
    lists.MockList(),
  })

  dir := http.Dir(".")
  fmt.Printf("Directory is %s\n", dir)
  http.Handle("GET /static/", http.FileServer(dir))
  http.HandleFunc("GET /login", loginHandler)
  http.HandleFunc("GET /", indexHandler)
  http.HandleFunc("GET /lists", listsHandler)
  http.HandleFunc("GET /lists/{listId}", listDetailHandler)

  fmt.Printf("Server started at http://localhost:8080\n")
  log.Fatal(http.ListenAndServe(":8080", nil))
}
