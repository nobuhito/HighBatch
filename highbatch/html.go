package highbatch

import (
	"strings"
	"reflect"
	"html/template"
	"bytes"
)

type SpecInfo struct {
	Name string
	Elm string
	Desc string
	Url string
	Key string
}

func getSpecInfo() []SpecInfo {
	spec := &Spec{}
	rt := reflect.TypeOf(*spec)

	var ret []SpecInfo
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		elm := field.Tag.Get("elm")
		if elm == "" {
			continue
		}
		si := SpecInfo{
			Elm: elm,
			Name: field.Tag.Get("json"),
			Desc: field.Tag.Get("desc"),
      Url: field.Tag.Get("url"),
      Key: field.Tag.Get("key"),
		}
		ret = append(ret, si)
	}
	return ret
}

func isExists(slice []string, key string) bool {
	list := map[string]bool{}
	for i := range slice {
		list[slice[i]] = true
	}
	if list[key] {
		return true
	}
	return false
}

func reverseSlice(slice []string) []string {
	ret := []string{}
	for i := range slice {
		ret = append(ret, slice[len(slice) - i - 1])
	}
	return ret
}

func getAddTaskPage() (html string) {
	specInfo := getSpecInfo()

	b := new(bytes.Buffer)
	tmpl := template.Must(template.ParseFiles("highbatch/html/AddTaskPage.html"))
	tmpl.Execute(b, specInfo)
	return b.String()
}

func getMainPage() (html string) {
	b := new(bytes.Buffer)
	tmpl := template.Must(template.ParseFiles("highbatch/html/MainPage.html"))
	tmpl.Execute(b, nil)
	return b.String()
}

func getHtml(page, js string) string {

	pageHtml := ""
	switch page {
	case "AddTaskPage":
		pageHtml = getAddTaskPage()
	case "MainPage":
		pageHtml = getMainPage()
	}

	base := new(bytes.Buffer)
	baseTmpl := template.Must(template.ParseFiles("highbatch/html/Base.html"))
	baseTmpl.Execute(base, nil)
	s := base.String()
	s = strings.Replace(s, "%%HTML%%", pageHtml, -1)
	s = strings.Replace(s, "%%JS%%", js, -1)
	return s
}
