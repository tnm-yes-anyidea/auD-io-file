.PHONY: all build run clean deps zip

APP_NAME = auD-io-file
BUILD_DIR = bin

deps:
	go mod tidy

build: deps
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(APP_NAME) ./cmd/app

run: build
	./$(BUILD_DIR)/$(APP_NAME)

clean:
	rm -rf $(BUILD_DIR) $(APP_NAME).zip

zip: clean deps
	zip -r $(APP_NAME).zip . -x ".git/*" -x "bin/*"
	@echo "Repository zipped as $(APP_NAME).zip"
