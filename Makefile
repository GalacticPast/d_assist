ifeq ($(OS),Windows_NT)
	SHELL := cmd.exe

	ASSEMBLY  := app
	EXTENSION := .exe

else
	ASSEMBLY  := app
	EXTENSION := 
endif 

COMPILER_FLAGS := -gcflags="all=-N -l"

build:
	go build  $(COMPILER_FLAGS) -o $(ASSEMBLY)$(EXTENSION) cmd/server/main.go

