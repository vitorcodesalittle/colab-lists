package views

import (
	"bytes"
	"io"
	textTemplate "text/template"

	"vilmasoftware.com/colablists/pkg/list"
	"vilmasoftware.com/colablists/pkg/user"
)

func (t *templates) RenderIndex(w io.Writer, args *IndexArgs) {
	t.Index.Execute(w, args)
}

type ListArgs struct {
	List     list.List
	Editing  bool
	AllUsers []user.User
	IsDirty  bool
}

func (t *templates) RenderList(w io.Writer, args *ListArgs) {
	t.renderBase(w, &baseArgs{ExtraHead: t.ExecuteTemplateString(t.List, "extrahead", args), Body: t.ExecuteTemplateString(t.List, "body", args)})
}

type ListsArgs struct {
	Lists []list.List
}

func (t *templates) RenderLists(w io.Writer, args *ListsArgs) {
	t.renderBase(w, &baseArgs{Body: t.ExecuteTemplateString(t.Lists, "body", args)})
}

func (t *templates) RenderLogin(w io.Writer) {
	t.renderBase(w, &baseArgs{Body: t.ExecuteTemplateString(t.Auth, "bodylogin", nil)})
}

func (t *templates) RenderSignup(w io.Writer) {
	t.renderBase(w, &baseArgs{Body: t.ExecuteTemplateString(t.Auth, "bodysignup", nil)})
}

type baseArgs struct {
	ExtraHead string
	Body      string
}

func (t *templates) renderBase(w io.Writer, args *baseArgs) {
	err := t.Base.ExecuteTemplate(w, "base", args)
	if err != nil {
		panic(err)
	}
}

func (t *templates) ExecuteTemplateString(template *textTemplate.Template, templateName string, args interface{}) string {
	b := bytes.NewBufferString("")
    err := template.ExecuteTemplate(b, templateName, args)
    if err != nil {
        panic(err)
    }
	return b.String()
}
