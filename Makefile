build:
	go build -o bin/gotetris ./cmd/gotetris
run:
	./bin/gotetris

clean:
	rm -rf bin
	rm -f gotetris

	