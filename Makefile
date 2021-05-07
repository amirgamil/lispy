CMD= ./cmd/lispy.go
RUN = go run ${CMD}

all: repl build

#repl is default rule for now, when we add reading from file, will add that as first
repl:
	${RUN} -repl


build:
	go build -o lispy ${CMD}