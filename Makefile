bin_dir = ./bin
binary = telegraf-input-lotus

all: clean build

build:
	mkdir -p $(bin_dir)
	go build -o $(bin_dir)/$(binary) cmd/main.go

clean:
	rm -rf $(bin_dir)