test:
	@go test -coverprofile cover.out

coverage:
	@go test -coverprofile cover.out && go tool cover -html=cover.out

.PHONY: coverage test
