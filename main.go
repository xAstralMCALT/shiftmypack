package main

import (
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/restartfu/shiftmypack/shiftmypack"
	"github.com/restartfu/shiftmypack/shiftmypack/java"
)

var indexTemplate = template.Must(template.ParseFiles("web/index.html"))

func main() {
	input := flag.String("i", "", "input path")
	output := flag.String("o", ".", "output path")
	addr := flag.String("addr", ":8080", "web server address")
	flag.Parse()

	if len(*input) > 0 {
		runCLI(*input, *output)
		return
	}

	fmt.Println("Shift My Pack website available at http://localhost" + *addr)
	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("web"))))
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/convert", convertHandler)
	log.Fatal(http.ListenAndServe(*addr, nil))
}

func runCLI(input, output string) {
	javapack, err := java.NewResourcePack(input)
	if err != nil {
		log.Fatal(err)
	}
	if err := shiftmypack.PortJavaEditionPackAndExtract(javapack, output); err != nil {
		log.Fatal(err)
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if err := indexTemplate.Execute(w, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func convertHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	if err := r.ParseMultipartForm(100 << 20); err != nil {
		http.Error(w, "invalid upload: "+err.Error(), http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("pack")
	if err != nil {
		http.Error(w, "upload required: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	tmpDir := "tmp"
	if err := os.MkdirAll(tmpDir, 0755); err != nil {
		http.Error(w, "failed to create temp directory: "+err.Error(), http.StatusInternalServerError)
		return
	}

	uploaded, err := os.CreateTemp(tmpDir, "shiftmypack-*.zip")
	if err != nil {
		http.Error(w, "unable to save upload: "+err.Error(), http.StatusInternalServerError)
		return
	}
	defer func() {
		uploaded.Close()
		os.Remove(uploaded.Name())
	}()

	if _, err := io.Copy(uploaded, file); err != nil {
		http.Error(w, "unable to save upload: "+err.Error(), http.StatusInternalServerError)
		return
	}

	javapack, err := java.NewResourcePack(uploaded.Name())
	if err != nil {
		http.Error(w, "invalid Java pack: "+err.Error(), http.StatusBadRequest)
		return
	}

	outputFile, err := os.CreateTemp(tmpDir, "shiftmypack-*.mcpack")
	if err != nil {
		http.Error(w, "unable to create output file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	outputFile.Close()
	defer os.Remove(outputFile.Name())

	if err := shiftmypack.PortJavaEditionPack(javapack, outputFile.Name()); err != nil {
		http.Error(w, "conversion failed: "+err.Error(), http.StatusInternalServerError)
		return
	}

	filename := strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename)) + ".mcpack"
	w.Header().Set("Content-Type", "application/vnd.minecraft.resource_pack")
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	http.ServeFile(w, r, outputFile.Name())
}
