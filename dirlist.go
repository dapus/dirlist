package dirlist

import (
	"net/http"
	"io"
	"log"
	"html/template"
	"os"
	"net/url"
	"sort"
)

type DirList struct {
	Dir http.Dir
	UrlPrefix string
	Tpl *template.Template
	IndexFileName string
}

func (d *DirList) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	urlPath := r.URL.Path[len(d.UrlPrefix):]
	method := r.Method

	log.Printf("%s %s", method, urlPath)

	if method != "GET" {
		http.NotFound(w, r)
		return
	}

	file, err := d.Dir.Open(urlPath)

	if err != nil {
		http.NotFound(w, r)
		return
	}

	st, err := file.Stat()

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if !st.IsDir() {
		d.ServeFile(w, r, file)
		return
	}

	if urlPath[len(urlPath)-1:] != "/" {
		http.Redirect(w, r, urlPath + "/", http.StatusMovedPermanently)
		return
	}

	files, err := file.Readdir(0)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	sort.Sort(FileSorter(files))

	var indexFile *os.File
	{
		f, err := d.Dir.Open(urlPath + d.IndexFileName)

		if err == nil {
			indexFile = f.(*os.File)
		}

	}

	d.ServeDir(w, r, files, indexFile)
}

func (d *DirList) ServeFile(w http.ResponseWriter, r *http.Request, file http.File) {
	_, err := io.Copy(w, file)

	if err != nil {
		log.Print(err)
	}
}

func (d *DirList) ServeDir(w http.ResponseWriter, r *http.Request, files []os.FileInfo, index *os.File) {
	context := &TplContext{
		Files: files,
		Index: index,
		Url: r.URL,
		Host: r.Host,
	}

	w.Header().Add("Content-Type", "text/html; charset=UTF-8")

	err := d.Tpl.Execute(w, context)

	if err != nil {
		log.Printf("Failed to execure template: %s", err)
	}
}

type TplContext struct {
	Files []os.FileInfo
	Index *os.File
	Url *url.URL
	Host string
}

// For sorting files list
type FileSorter []os.FileInfo

func (f FileSorter) Len() int {
	return len(f)
}

func (f FileSorter) Swap(i, j int) {
    f[i], f[j] = f[j], f[i]
}

func (f FileSorter) Less(i, j int) bool {
    return f[i].Name() < f[j].Name()
}
