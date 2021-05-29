CMD= ./cmd/lispy.go
RUN = go run ${CMD}

all: repl build

#repl is default rule for now, when we add reading from file, will add that as first
repl:
	${RUN} -repl

build:
	go build -o lispy ${CMD}


test:
	go build -o lispy ${CMD}
	./lispy tests/test1.lpy
	./lispy tests/test2.lpy
	./lispy tests/test3.lpy
	./lispy tests/test4.lpy
	./lispy tests/test5.lpy
	./lispy tests/test6.lpy
	./lispy tests/test7.lpy

