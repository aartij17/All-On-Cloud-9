BASE_DIR=$(GOPATH)/src/All-On-Cloud-9
TARGET_DIR=$(BASE_DIR)/target
BIN_OUT=$(TARGET_DIR)/bin

GO_PATH=${GOPATH}

# mandatory env variables for us to proceed
variables := GOPATH
fatal_if_undefined = $(if $(findstring undefined,$(origin $1)),$(error Error: variable [$1] is undefined))
$(foreach 1,$(variables),$(fatal_if_undefined))

# make sure in OSX, you have bash version >= 4.0
#
# To update bash:
# wget http://ftp.gnu.org/gnu/bash/bash-5.0.tar.gz
# tar xzf bash-5.0.tar.gz
# cd bash-5.0
# ./configure && make && sudo make install
# sudo bash -c "echo /usr/local/bin/bash >> /private/etc/shells"
UNAME_S := $(shell uname -s)
ifeq ($(UNAME_S),Darwin)
  SHELL=/usr/local/bin/bash
else
  SHELL=/bin/bash
endif

.PHONY:  build local fmt test show format

pr-build: clean fmt build test

# Shows the variables we just set.
# By prepending `@` we prevent Make from printing the command before the
# stdout of the execution.
show:
	@echo "SRC  = $(GO_SRC_DIRS)"
	@echo "TEST = $(GO_TEST_DIRS)"

format:
	gofmt -w `find . -mindepth 1 -maxdepth 1 -type d | cut -c 3- | grep -vE '^\.'`

# Note that we're not using cd to get into the directories.
# That’s because not knowing how deep in the file structure we’ll go,
# just stacking the directory changes with  pushd is easier as to get
# back to the original place we just need to popd.
fmt: $(GO_SRC_DIRS)
	@for dir in $^; do \
		pushd ./$$dir > /dev/null ; \
		go fmt ; \
		popd > /dev/null ; \
	done;

test: $(GO_TEST_DIRS)
	@for dir in $^; do \
		pushd ./$$dir > /dev/null ; \
		go test -v ; \
		popd > /dev/null ; \
	done;

clean:
	rm -rf $(TARGET_DIR)

build:
	mkdir -p $(BIN_OUT)
	declare -A projects=(["orderers"]=consensus/orderers/main.go ["server"]=server/main.go )\
	;for i in "$${!projects[@]}"; do \
		echo $${i}; \
		if [[ ${UNAME_S} == "Darwin" ]]; then \
		  echo "Building on Mac. Target Platform Linux"; \
		  sudo env GOPATH=$(GO_PATH)  env GOOS=linux GOARCH=amd64 go build -o $(BIN_OUT)/$${i} -i $(BASE_DIR)/$${projects[$${i}]}; \
		else \
		  echo "Building on Linux."; \
		  env GOPATH=$(GO_PATH)  env GOOS=linux GOARCH=amd64 env CGO_ENABLED=0 go build -o $(BIN_OUT)/$${i} -i $(BASE_DIR)/$${projects[$${i}]}; \
		fi \
	done;


local: clean build
