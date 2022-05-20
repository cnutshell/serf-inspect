.PHONY: member
member:
	go build -o member cmd/main.go

.PHONY: clean
clean:
	rm -f member

