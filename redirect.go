package main

import (
	"bytes"
	"html/template"
	"net/http"
	"strings"
)

const (
	hostname = "go.kendal.io"
)

var (
	tmpl = template.Must(template.New("github").Parse(`<!DOCTYPE html>
<html>
<head>
<meta http-equiv="Content-Type" content="text/html; charset=utf-8"/>
<meta name="go-import" content="{{.Package}} git {{.Remote}}"/>
</head>
</html>
`))
)

// PkgDescriptor is used to fill out the HTTP redirect template above.
type PkgDescriptor struct {
	// Package is the path of the requested Go package.
	Package string

	// Remote is the remote repository containing Package.
	Remote string
}

func Redirect(w http.ResponseWriter, r *http.Request) {
	_ = requireHTTPS(w, r) ||
		requireGET(w, r) ||
		redirectBrowserToGodoc(w, r) ||
		redirectToGithub(w, r)
}

func requireHTTPS(w http.ResponseWriter, r *http.Request) bool {
	if r.URL.Scheme == "http" {
		r.URL.Scheme = "https"
		http.Redirect(w, r, r.URL.String(), http.StatusMovedPermanently)
		return true
	}
	return false
}

func requireGET(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != http.MethodGet {
		status := http.StatusMethodNotAllowed
		http.Error(w, http.StatusText(status), status)
		return true
	}
	return false
}

func redirectBrowserToGodoc(w http.ResponseWriter, r *http.Request) bool {
	if r.FormValue("go-get") != "1" {
		var newURL = "https://godoc.org/" + r.Host + r.URL.Path
		if strings.TrimPrefix(r.URL.Path, "/") == "" {
			newURL = "https://godoc.org/?q=" + hostname
		}
		http.Redirect(w, r, newURL, http.StatusTemporaryRedirect)
		return true
	}
	return false
}

func redirectToGithub(w http.ResponseWriter, r *http.Request) bool {
	pkg := strings.Trim(r.URL.Path, "/")

	// Redirect go.kendal.io/foo/cmd/bar -> go.kendal.io/foo
	parts := strings.Split(pkg, "/")
	if len(parts) > 1 {
		pkg = parts[0]
	}
	descriptor := &PkgDescriptor{
		Package: hostname + "/" + pkg,
		Remote:  "https://github.com/kharland/" + pkg,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, descriptor); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return true
	}

	w.Header().Set("Cache-Control", "public, max-age=600")
	w.Write(buf.Bytes())
	return true
}
