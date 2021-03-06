cmdnames = server client callee load
cmds = $(addprefix juggler-, $(cmdnames))

# run `make` to build all commands.
# run `make flags=-race` to build with race detector.
# assign any valid build flag to flags to build with that set of flags.
all: $(cmds)

$(cmds):
	go build -i $(flags) ./cmd/$@ 

.PHONY: all $(cmds) cluster

