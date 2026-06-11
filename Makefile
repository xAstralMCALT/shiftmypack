install:
	mkdir -p bin
	go build -o bin/shiftmypack .
	sudo cp bin/shiftmypack /usr/bin/shiftmypack
