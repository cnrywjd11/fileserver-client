## fileserver-client
제한된 환경에서 다른 서버 간에 파일 또는 디렉토리를 복사하기 위한 파일 서버,클라이언트 

### build
직접 빌드:

    $ go build fileserver.go
    $ go build fileserver-client.go

linux binary:

    $ GOOS=linux GOARCH=amd64 go build fileserver.go
    $ GOOS=linux GOARCH=amd64 go build fileserver-client.go
### usage
run:

    $ 파일서버: ./fileserver {port}
    $ 파일 받을 클라이언트: ./fileserver-client {host:port} {srcDir} {dstDir} {제한 cpu core 개수}
