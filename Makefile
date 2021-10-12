bin_dir = ./bin
binary = telegraf-input-lotus

all: clean build

build:
	mkdir -p $(bin_dir)
	go build -o $(bin_dir)/$(binary) cmd/main.go

clean:
	rm -rf $(bin_dir)

# A much better alternative to 'ls' https://the.exa.website/
overview:
	exa -lhTa --no-user --no-time -I ".git|.vagrant"

dev:
	cd env && vagrant up