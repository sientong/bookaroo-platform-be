.PHONY: test test-unit test-integration

# Run all tests
test: test-unit test-integration

# Run unit tests
test-unit:
	go test -v ./tests/handlers/...

# Run integration tests
test-integration:
	go test -v ./tests/integration/...

# Create test database
setup-test-db:
	createdb bookaroo_test

# Drop test database
drop-test-db:
	dropdb bookaroo_test

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./tests/...
	go tool cover -html=coverage.out
