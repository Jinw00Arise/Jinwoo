BINARY_LOGIN=jinwoo-login
BINARY_CHANNEL=jinwoo-channel
CHANNEL_COUNT ?= 1
CHANNEL_BASE_PORT ?= 8585

build: build-login build-channel

build-login:
	go build -o bin/$(BINARY_LOGIN) ./cmd/login

build-channel:
	go build -o bin/$(BINARY_CHANNEL) ./cmd/channel

run: build
	@./bin/$(BINARY_LOGIN) & echo $$! > .login.pid; \
	for i in $$(seq 0 $$(($(CHANNEL_COUNT)-1))); do \
		port=$$(($(CHANNEL_BASE_PORT) + i)); \
		./bin/$(BINARY_CHANNEL) -channel=$$i -port=$$port & echo $$! > .channel$$i.pid; \
	done; \
	trap 'kill $$(cat .login.pid .channel*.pid) 2>/dev/null; rm -f .login.pid .channel*.pid' EXIT INT TERM; \
	wait

run-login: build-login
	./bin/$(BINARY_LOGIN)

run-channel: build-channel
	@for i in $$(seq 0 $$(($(CHANNEL_COUNT)-1))); do \
		port=$$(($(CHANNEL_BASE_PORT) + i)); \
		./bin/$(BINARY_CHANNEL) -channel=$$i -port=$$port & echo $$! > .channel$$i.pid; \
	done; \
	trap 'kill $$(cat .channel*.pid) 2>/dev/null; rm -f .channel*.pid' EXIT INT TERM; \
	wait

stop:
	@pkill -f jinwoo- 2>/dev/null || true
	@rm -f .login.pid .channel*.pid
	@echo "Servers stopped"

clean: stop
	rm -rf bin