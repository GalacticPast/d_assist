build:
	go build  -gcflags="all=-N -l" -o app cmd/server/main.go

