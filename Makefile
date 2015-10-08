all: test deploy 

deploy:
	godep go build -v .

test:
	godep go test -v ./nethooks ./labels

clean:
	rm deploy
