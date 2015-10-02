
all: test deploy 

deploy:
	godep go build .

test:
	godep go test ./nethooks ./labels

clean:
	rm deploy
