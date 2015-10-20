all: linux darwin win32

linux:
	env GOOS=linux GOARCH=amd64 go build -v -o git-stats-linux

darwin:
	env GOOS=darwin GOARCH=amd64 go build -v -o git-stats-mac

win32:
	env GOOS=windows GOARCH=386 go build -v -o git-stats-win32.exe

clean:
	rm -rf git-stats-win32.exe git-stats-mac git-stats-linux
