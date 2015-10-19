all: native win32

native:
	go build -v

win32:
	env GOOS=windows GOARCH=386 go build -v
