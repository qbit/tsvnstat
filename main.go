package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"sort"
	"strings"
	"time"

	"tailscale.com/tsnet"
)

//go:embed generate_images.sh
var genScript []byte

//go:embed style.css
var style string

func main() {
	tmpDir, err := os.MkdirTemp("", "tsvnstat")
	if err != nil {
		log.Fatal(err)
	}

	defer os.RemoveAll(tmpDir)

	tmpFile, err := os.CreateTemp("", "generate_images.sh")
	if err != nil {
		log.Fatal(err)
	}

	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(genScript); err != nil {
		log.Fatal(err)
	}
	if err := tmpFile.Close(); err != nil {
		log.Fatal(err)
	}

	err = os.Chmod(tmpFile.Name(), 0700)
	if err != nil {
		log.Fatal(err)
	}

	name := flag.String("name", "", "name of service")
	dir := flag.String("dir", tmpDir, "directory containing vnstat images")
	flag.Parse()

	s := &tsnet.Server{
		Hostname: *name,
	}

	ln, err := s.Listen("tcp", ":80")
	if err != nil {
		log.Fatal(err)
	}
	defer ln.Close()

	host, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		for {
			log.Printf("running %q in %q", tmpFile.Name(), tmpDir)

			ifaces, err := net.Interfaces()
			if err != nil {
				log.Fatal(err)
			}

			var ifNames []string
			for _, intf := range ifaces {
				ifNames = append(ifNames, intf.Name)
			}

			genCmd := exec.Command(tmpFile.Name(), ifNames...)
			genCmd.Dir = *dir
			_, err = genCmd.Output()
			if err != nil {
				log.Fatal(err)
			}
			time.Sleep(5 * time.Minute)
		}
	}()

	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir(*dir))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fileServer.ServeHTTP(w, r)
	})
	mux.HandleFunc("/index.html", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		images, err := os.ReadDir(*dir)
		if err != nil {
			http.Error(w, err.Error(), 500)
			return
		}

		fmt.Fprintf(w, `<html>
		<head>
		<title>%s</title>
		<style>%s</style>
		</head>
		<body>
		<h1>vnstat for %s</h1>`, host, style, host)

		sort.Slice(images, func(i, j int) bool {
			return images[i].Name() < images[j].Name()
		})

		prefix := ""
		oldPrefix := ""
		for _, img := range images {
			in := img.Name()
			imgPrefix := strings.Split(in, "-")[0]
			headImg := fmt.Sprintf("%s-s.png", imgPrefix)

			if prefix != imgPrefix {
				if oldPrefix != prefix {
					fmt.Fprintf(w, "</p></details>")
				}
				fmt.Fprintf(w, "<details><summary><img src=%q /></summary><p>", headImg)
				prefix = imgPrefix
			}

			if in != headImg {
				fmt.Fprintf(w, "<img src=%q /><br />", in)
			}
		}
	}))

	hs := &http.Server{
		Handler: mux,
	}

	log.Panic(hs.Serve(ln))
}
