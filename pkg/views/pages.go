package views

import (
	"bytes"
	"io"
	textTemplate "text/template"

	"vilmasoftware.com/colablists/pkg/community"
	"vilmasoftware.com/colablists/pkg/list"
	"vilmasoftware.com/colablists/pkg/user"
)

func (t *templates) RenderIndex(w io.Writer, args *IndexArgs) {
	t.Index.Execute(w, args)
}

type ListArgs struct {
	List     ListUi
	Editing  bool
	AllUsers []user.User
	IsDirty  bool
}

func (t *templates) RenderList(w io.Writer, args *ListArgs) {
	t.renderBase(w, &baseArgs{ExtraHead: t.ExecuteTemplateString(t.List, "extrahead", args), Body: t.ExecuteTemplateString(t.List, "body", args), Title: "!!!!" + args.List.Title, Description: GetDescription("")})
}

type ListsArgs struct {
	Lists []list.List
}

func (t *templates) RenderLists(w io.Writer, args *ListsArgs) {
	t.renderBase(w, &baseArgs{Body: t.ExecuteTemplateString(t.Lists, "body", args), Title: "your marketlists", Description: GetDescription("")})
}

func (t *templates) RenderLogin(w io.Writer, args *SignupArgs) {
	t.renderBase(w, &baseArgs{Body: t.ExecuteTemplateString(t.Auth, "bodylogin", args), Title: "Login", Description: GetDescription("")})
}

type SignupArgs struct {
	FormError string
}

func (t *templates) RenderSignup(w io.Writer, args *SignupArgs) {
	t.renderBase(w, &baseArgs{Body: t.ExecuteTemplateString(t.Auth, "bodysignup", args), Title: "Sign up", Description: GetDescription("Sign up here. Create an account. Fully manage your groceries lists with your family.")})
}

type baseArgs struct {
	ExtraHead   string
	Body        string
	Title       string
	Description string
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

func GetDescription(msg string) string {
	return `marketlist is an application to manage lists colaboratively.` + msg
}

type CommunitiesQuery struct {
	SelectedId int64
	New        bool
	EditingId  int64
}
type CommunitiesArgs struct {
	Query             CommunitiesQuery
	Communities       []*community.Community
	SelectedCommunity *community.Community
}

func (t *templates) RenderCommunities(w io.Writer, args *CommunitiesArgs) {
	t.renderBase(w, &baseArgs{
		Title:       "Communities",
		Description: GetDescription("Communities"),
		Body:        t.ExecuteTemplateString(t.Communities, "communitiesbody", args),
	})
}
