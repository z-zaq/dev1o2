package views

import (
	"html/template"
	"net/http"
)
func RenderTemplate(w http.ResponseWriter, file string, data interface{}){
	tmpl, err := template.ParseFiles(
		"templates/base.html", "templates/"+file)
	if err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	err = tmpl.Execute(w, data)
	if err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}