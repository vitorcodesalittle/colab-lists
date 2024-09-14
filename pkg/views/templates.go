package views

import (
	"io"
	textTemplate "text/template"
)

type templates struct {
	Index *textTemplate.Template
	List  *textTemplate.Template
	Lists *textTemplate.Template
	Login *textTemplate.Template
}

type IndexArgs struct {
	Title       string
	Description string
}

func (t *templates) RenderIndex(w io.Writer, args any) {
	t.Index.Execute(w, args)
}

func (t *templates) RenderList(w io.Writer, args any) {
	t.List.Execute(w, args)
}

func (t *templates) RenderLists(w io.Writer, args any) {
	t.Lists.Execute(w, args)
}

type LoginArgs struct{}

func (t *templates) RenderLogin(w io.Writer, args any) {
	t.Login.Execute(w, args)
}

func (t *templates) RenderCollaboratorsList(w io.Writer, args any) {
	t.List.ExecuteTemplate(w, "colaborators", args)
}

func newTemplates() *templates {
	templates := &templates{}
	templates.Index = textTemplate.Must(textTemplate.ParseFiles("./templates/pages/index.html"))
	templates.Login = textTemplate.Must(textTemplate.ParseFiles("./templates/pages/login.html"))
	templates.Lists = textTemplate.Must(textTemplate.ParseFiles("./templates/pages/lists.html"))
	templates.List = textTemplate.Must(textTemplate.ParseFiles("./templates/pages/list.html"))
	return templates
}

var Templates *templates = newTemplates()
