.PHONY: lint
lint:
	golangci-lint run

.PHONY: watch
watch:
	watchexec --stop-timeout=0s --debounce=1s --wrap-process=session --restart -- "go run ."
