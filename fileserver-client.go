package main

import (
	"bufio"
	"encoding/xml"
	"errors"
	"io"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

func main() {
	if len(os.Args) < 5 {
		log.Fatal("host, src, dst, 제한 cpu core 개수를 입력하세요")
	}
	host := os.Args[1]
	src := os.Args[2]
	dst := os.Args[3]
	strThreadCounts := os.Args[4]

	threadCounts, err := strconv.Atoi(strThreadCounts)
	if err != nil {
		log.Fatal(err)
	}
	runtime.GOMAXPROCS(threadCounts)

	src, dst = modifyPath(host, src, dst)
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		if !errors.Is(err, fs.ErrExist) {
			log.Fatal(err)
		}
	}

	fileCounts := checkFileCount(host, src)
	dones := make(chan bool, fileCounts)

	recursiveDownload(host, src, dst, dones)
	for {
		if fileCounts == 0 {
			break
		}
		<-dones
		fileCounts--
	}
}

func modifyPath(host, src, dst string) (string, string) {
	if strings.HasSuffix(src, "/") &&
		strings.HasSuffix(dst, "/") {
		return src, dst
	}

	src = strings.TrimSuffix(src, "/")
	dst = strings.TrimSuffix(dst, "/")

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Head("http://" + host + src)
	if err != nil {
		log.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode == http.StatusMovedPermanently {
		src += "/"
		dst += "/"
	}

	return src, dst
}

func checkFileCount(host, src string) int {
	if !strings.HasSuffix(src, "/") {
		return 1
	}

	resp, err := http.Get("http://" + host + src)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("http://%s%s , status code: %d\n", host, src, resp.StatusCode)
	}

	bbuf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	sbuf := string(bbuf)
	splitArr := strings.Split(sbuf, "\n")

	fileCounts := 0
	for _, su := range splitArr {
		if !strings.Contains(su, "<a href=") {
			continue
		}
		var A string
		xml.Unmarshal([]byte(su), &A)
		fileCounts += checkFileCount(host, src+A)
	}

	return fileCounts
}

func recursiveDownload(host, src, dst string, done chan bool) {
	srcPath := "http://" + host + src

	if !strings.HasSuffix(src, "/") {
		go func() {
			resp, err := http.Get(srcPath)
			if err != nil {
				log.Fatal(err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				log.Fatalf("http://%s%s , status code: %d\n", host, src, resp.StatusCode)
			}

			log.Printf("%s: start\n", srcPath)

			f, err := os.Create(dst)
			if err != nil {
				log.Fatal(err)
			}
			defer f.Close()

			br := bufio.NewReader(resp.Body)
			bw := bufio.NewWriter(f)
			buf := make([]byte, 2048)

			for {
				rbytes, err := br.Read(buf)
				if err != nil {
					if errors.Is(err, io.EOF) {
						break
					} else {
						log.Fatal(err)
					}
				}
				if _, err := bw.Write(buf[:rbytes]); err != nil {
					log.Fatal(err)
				}
				bw.Flush()
			}

			log.Printf("%s: end\n", srcPath)
			done <- true
		}()
		time.Sleep(100 * time.Millisecond)

		return
	}

	err := os.Mkdir(dst, 0755)
	if err != nil {
		if !errors.Is(err, fs.ErrExist) {
			log.Fatal(err)
		}
	}

	resp, err := http.Get(srcPath)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Fatalf("http://%s%s , status code: %d\n", host, src, resp.StatusCode)
	}

	bbuf, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	sbuf := string(bbuf)
	splitArr := strings.Split(sbuf, "\n")

	for _, su := range splitArr {
		if !strings.Contains(su, "<a href=") {
			continue
		}
		var A string
		xml.Unmarshal([]byte(su), &A)
		recursiveDownload(host, src+A, dst+A, done)
	}
}
