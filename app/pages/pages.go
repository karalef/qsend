package pages

import (
	_ "embed"
	"html/template"
	"net/http"
)

//go:embed done.html
var doneFileData string

//go:embed error.html
var errorFileData string

//go:embed upload.html
var uploadFileData string

var (
	done          = template.Must(template.New("done").Parse(doneFileData))
	errorTemplate = template.Must(template.New("error").Parse(errorFileData))
	upload        = template.Must(template.New("upload").Parse(uploadFileData))
)

func RenderDone(w http.ResponseWriter, output string) {
	render(w, done, map[string]string{"Output": output})
}

func RenderError(w http.ResponseWriter, err string) {
	render(w, errorTemplate, map[string]string{"Error": err})
}

func RenderUpload(w http.ResponseWriter, route string) {
	render(w, upload, map[string]string{"Route": route})
}

func render(w http.ResponseWriter, tpl *template.Template, data any) {
	err := tpl.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
