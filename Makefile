all: linux darwin win32

linux:
	env GOOS=linux GOARCH=amd64 go build -v && tar zcvf git-stats-linux.tar.gz git-stats

darwin:
	env GOOS=darwin GOARCH=amd64 go build -v && tar zcvf git-stats-darwin.tar.gz git-stats

win32:
	env GOOS=windows GOARCH=386 go build -v && zip git-stats-win32.zip git-stats.exe

clean:
	rm -rf git-stats.exe git-stats git-stats-linux.tar.gz git-stats-darwin.tar.gz git-stats-win32.zip
