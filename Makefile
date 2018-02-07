VERSION=0.0.1
BUILD_PATH=build
FILE_COMMAND=triton-kubernetes

OSX_ARCHIVE_PATH=$(BUILD_PATH)/triton-kubernetes_osx-amd64.zip
OSX_BINARY_PATH=$(BUILD_PATH)/triton-kubernetes_osx-amd64

LINUX_BINARY_PATH=$(BUILD_PATH)/triton-kubernetes_linux-amd64

clean:
	@rm -rf ./build

build: build-osx build-linux
	@echo "Generating checksums..."
	@cd build; shasum -a 256 * > sha256-checksums.txt

build-osx: clean
	@echo "Building OSX..."
	@mkdir -p $(BUILD_PATH)
	@GOOS=darwin GOARCH=amd64 go build -o $(OSX_BINARY_PATH)
	@zip --junk-paths $(OSX_ARCHIVE_PATH) $(OSX_BINARY_PATH)

build-linux: clean
	@echo "Building Linux..."
	@mkdir -p $(BUILD_PATH)
	@GOOS=linux GOARCH=amd64 go build -o $(LINUX_BINARY_PATH)
