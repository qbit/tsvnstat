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

//go:embed style.css
var style string

func runCmd(cmd string, dir string, args ...string) {
	genCmd := exec.Command(cmd, args...)
	genCmd.Dir = dir
	out, err := genCmd.Output()
	if err != nil {
		log.Println(string(out), err)
	}
}

func main() {
	tmpDir, err := os.MkdirTemp("", "tsvnstat")
	if err != nil {
		log.Fatal(err)
	}

	defer os.RemoveAll(tmpDir)

	name := flag.String("name", "", "name of service")
	dir := flag.String("dir", tmpDir, "directory containing vnstat images")
	key := flag.String("key", "", "path to file containing the api key")
	vnstati := flag.String("vnstati", "/bin/vnstati", "path to vnstati")
	flag.Parse()

	s := &tsnet.Server{
		Hostname: *name,
	}

	if *key != "" {
		keyData, err := os.ReadFile(*key)
		if err != nil {
			log.Fatal(err)
		}
		s.AuthKey = string(keyData)
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
			ifaces, err := net.Interfaces()
			if err != nil {
				log.Fatal("can't get interfaces...", err)
			}

			for _, iface := range ifaces {
				if iface.Flags&net.FlagUp == 0 {
					continue
				}
				runCmd(*vnstati, *dir, "--style", "1", "-L", "-s", "-o", fmt.Sprintf("%s-s.png", iface.Name), iface.Name)
				runCmd(*vnstati, *dir, "--style", "1", "-L", "--fivegraph", "576", "218", "-o", fmt.Sprintf("%s-5g.png", iface.Name), iface.Name)
				runCmd(*vnstati, *dir, "--style", "1", "-L", "-hg", "-o", fmt.Sprintf("%s-hg.png", iface.Name), iface.Name)
				runCmd(*vnstati, *dir, "--style", "1", "-L", "-h", "24", "-o", fmt.Sprintf("%s-h.png", iface.Name), iface.Name)
				runCmd(*vnstati, *dir, "--style", "1", "-L", "-d", "30", "-o", fmt.Sprintf("%s-d.png", iface.Name), iface.Name)
				runCmd(*vnstati, *dir, "--style", "1", "-L", "-t", "10", "-o", fmt.Sprintf("%s-t.png", iface.Name), iface.Name)
				runCmd(*vnstati, *dir, "--style", "1", "-L", "-m", "12", "-o", fmt.Sprintf("%s-m.png", iface.Name), iface.Name)
				runCmd(*vnstati, *dir, "--style", "1", "-L", "-y", "5", "-o", fmt.Sprintf("%s-y.png", iface.Name), iface.Name)
			}

			time.Sleep(5 * time.Minute)
		}
	}()

	mux := http.NewServeMux()
	fileServer := http.FileServer(http.Dir(*dir))

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		fileServer.ServeHTTP(w, r)
	})
	mux.HandleFunc("/index.html", func(w http.ResponseWriter, r *http.Request) {
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
	})

	hs := &http.Server{
		Handler: mux,
	}

	log.Panic(hs.Serve(ln))
}
