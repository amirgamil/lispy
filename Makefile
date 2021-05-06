CMD= ./cmd/lispy.go
RUN = go run ${CMD}

all: repl build

repl:
	${RUN} -repl


build:
	go build -o lispy ${CMD}