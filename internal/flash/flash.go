package flash

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"html/template"
	"log"
	"log/slog"
	"net/http"
	"time"
)

type Type string

var (
	Warning Type = "Warning"
	Success Type = "Success"
	Info    Type = "Info"
)

type Flash struct {
	Type    Type // Warning | Success | Info
	Title   string
	Content string
}

func init() {
	gob.Register(&Flash{})
}

// Done at init
var tpl = template.Must(template.New("").Parse(string(flashTemplate)))

// ReplaceTemplate allows you to replace the default, TailwindCSS flash
// with your own.
func ReplaceTemplate(newTemplate string) {
	tpl = template.Must(template.New("").Parse(newTemplate))
}

// Handler needs to be registered to a suitable file path such as "/notifications"
// You would need to use grab the contents of this page either using
// HTMX, iframe or XML Request. Returns a 201 No Content if no flash data
func Handler() http.Handler {
	return http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
		f, err := get(wr, req)
		if err != nil {
			wr.WriteHeader(http.StatusNoContent)
			return
		}
		tpl.Execute(wr, f)
	})
}

// HandlerWithLogger is slog compatible
// A handler needs to be registered to a suitable file path such as "/notifications"
// You would need to use grab the contents of this page either using
// HTMX, iframe or XML Request. Returns a 201 No Content if no flash data
func HandlerWithLogger(logger *slog.Logger) http.Handler {
	return http.HandlerFunc(func(wr http.ResponseWriter, req *http.Request) {
		f, err := get(wr, req)
		if err != nil {
			wr.WriteHeader(http.StatusNoContent)
			return
		}
		tpl.Execute(wr, f)
	})
}

// Set must be called before any other response is set as it writes out a cookie
//
//	flash.Set(wr, flash.Warning, "This is a warning", "This is the contents - i was set via Flash()")
func Set(w http.ResponseWriter, level Type, title string, content string) {

	// Initialize a buffer to hold the gob data.
	var buf bytes.Buffer
	f := Flash{
		Type:    level,
		Title:   title,
		Content: content,
	}

	err := gob.NewEncoder(&buf).Encode(&f)
	if err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	enc := base64.URLEncoding.EncodeToString(buf.Bytes())

	c := &http.Cookie{Name: "flash", Value: enc}
	http.SetCookie(w, c)
}

func get(w http.ResponseWriter, r *http.Request) (Flash, error) {
	var f Flash
	c, err := r.Cookie("flash")
	if err != nil {
		return f, err
	}

	// No point worrying about bad data. Just move on past.
	uDec, _ := base64.URLEncoding.DecodeString(c.Value)

	if err := gob.NewDecoder(bytes.NewReader(uDec)).Decode(&f); err != nil {
		log.Println(err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return Flash{}, err
	}

	// "Delete" by setting expired cookie
	http.SetCookie(w, &http.Cookie{Name: "flash", MaxAge: -1, Expires: time.Unix(1, 0)})
	return f, nil
}
