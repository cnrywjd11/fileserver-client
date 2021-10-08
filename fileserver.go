package main

import (
	"log"
	"net/http"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("파일서버 포트를 입력하세요.")
	}
	_, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatal("포트 번호 확인하세요.")
	}
	log.Fatal(http.ListenAndServe(":"+os.Args[1], http.FileServer(http.Dir("/"))))
}
