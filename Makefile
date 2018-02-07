VERSION=0.0.1
PATH_BUILD=build
FILE_COMMAND=triton-kubernetes

OSX_BUILD_PATH=$(PATH_BUILD)/mac_osx
OSX_ARCHIVE_PATH=$(OSX_BUILD_PATH)/triton-kubernetes.zip
OSX_BINARY_PATH=$(OSX_BUILD_PATH)/triton-kubernetes
OSX_CHECKSUM_PATH=$(OSX_BUILD_PATH)/sha256_checksum

LINUX_BUILD_PATH=$(PATH_BUILD)/linux

clean:
	@rm -rf ./build

build: build-osx build-linux

build-osx: clean
	@echo "Building OSX..."
	@mkdir -p $(OSX_BUILD_PATH)
	@GOOS=darwin GOARCH=amd64 go build -o $(OSX_BUILD_PATH)/triton-kubernetes
	@zip --junk-paths $(OSX_ARCHIVE_PATH) $(OSX_BINARY_PATH)
	@shasum -a 256 $(OSX_ARCHIVE_PATH) > $(OSX_CHECKSUM_PATH)

build-linux: clean
	@echo "Building Linux..."
	@mkdir -p $(LINUX_BUILD_PATH)
	@GOOS=linux GOARCH=amd64 go build -o $(LINUX_BUILD_PATH)/triton-kubernetes
