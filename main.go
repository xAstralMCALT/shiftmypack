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

	"github.com/xAstralMCALT/shiftmypack/shiftmypack"
	"github.com/xAstralMCALT/shiftmypack/shiftmypack/bedrock"
	"github.com/xAstralMCALT/shiftmypack/shiftmypack/java"
)

var indexTemplate = template.Must(template.ParseFiles("web/index.html"))

func main() {
	input := flag.String("i", "", "input path")
	output := flag.String("o", ".", "output path")
	addr := flag.String("addr", "0.0.0.0:80", "web server address")
	flag.Parse()

	if len(*input) > 0 {
		runCLI(*input, *output)
		return
	}

	fmt.Println("Shift Your Pack website available at http://shiftmypack.duckdns.org")
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

	mode := r.FormValue("mode")
	if mode == "" {
		mode = "java-to-bedrock"
	}

	version := r.FormValue("version")
	if version == "" {
		version = "1.20"
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

	outputFile, err := os.CreateTemp(tmpDir, "shiftmypack-*")
	if err != nil {
		http.Error(w, "unable to create output file: "+err.Error(), http.StatusInternalServerError)
		return
	}
	outputFile.Close()
	defer os.Remove(outputFile.Name())

	filename := strings.TrimSuffix(header.Filename, filepath.Ext(header.Filename))

	if mode == "bedrock-to-java" {
		bedrockpack, err := bedrock.NewResourcePack(uploaded.Name())
		if err != nil {
			http.Error(w, "invalid Bedrock pack: "+err.Error(), http.StatusBadRequest)
			return
		}

		outputPath := outputFile.Name() + ".zip"
		if err := shiftmypack.PortBedrockPackWithVersion(bedrockpack, outputPath, version); err != nil {
			http.Error(w, "conversion failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer os.Remove(outputPath)

		w.Header().Set("Content-Type", "application/zip")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+".zip\"")
		http.ServeFile(w, r, outputPath)
	} else {
		javapack, err := java.NewResourcePack(uploaded.Name())
		if err != nil {
			http.Error(w, "invalid Java pack: "+err.Error(), http.StatusBadRequest)
			return
		}

		outputPath := outputFile.Name() + ".mcpack"
		if err := shiftmypack.PortJavaEditionPack(javapack, outputPath); err != nil {
			http.Error(w, "conversion failed: "+err.Error(), http.StatusInternalServerError)
			return
		}
		defer os.Remove(outputPath)

		w.Header().Set("Content-Type", "application/vnd.minecraft.resource_pack")
		w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+".mcpack\"")
		http.ServeFile(w, r, outputPath)
	}
}
