BINARY_LOGIN=jinwoo-login
BINARY_CHANNEL=jinwoo-channel

build: build-login build-channel

build-login:
	go build -o bin/$(BINARY_LOGIN) ./cmd/login

build-channel:
	go build -o bin/$(BINARY_CHANNEL) ./cmd/channel

run: build
	@./bin/$(BINARY_LOGIN) & echo $$! > .login.pid; \
	./bin/$(BINARY_CHANNEL) & echo $$! > .channel.pid; \
	trap 'kill $$(cat .login.pid) $$(cat .channel.pid) 2>/dev/null; rm -f .login.pid .channel.pid' EXIT INT TERM; \
	wait

run-login: build-login
	./bin/$(BINARY_LOGIN)

run-channel: build-channel
	./bin/$(BINARY_CHANNEL)

stop:
	@pkill -f jinwoo- 2>/dev/null || true
	@rm -f .login.pid .channel.pid
	@echo "Servers stopped"

clean: stop
	rm -rf bin