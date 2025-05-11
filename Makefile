# Makefile for alethic-ism-core-go SDK
.PHONY: test test-coverage lint fmt vet clean version help

# Module name
MODULE := github.com/quantumwake/alethic-ism-core-go

# Run tests
test:
	go test ./... -v

# Run tests with coverage
test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

# Run linter
lint:
	@if command -v golangci-lint > /dev/null; then \
		echo "Running golangci-lint..."; \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not found, installing..."; \
		go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest; \
		golangci-lint run ./...; \
	fi

# Format code
fmt:
	go fmt ./...

# Run go vet
vet:
	go vet ./...

# Version bump (defaults to patch version)
# Usage:
#   make version           (patch version bump)
#   make version TYPE=patch  (patch version bump)
#   make version TYPE=minor  (minor version bump)
#   make version TYPE=major  (major version bump)
version:
	@TYPE=$${TYPE:-patch}; \
	echo "Bumping $$TYPE version..."; \
	git fetch --tags; \
	LATEST_TAG=$$(git describe --tags --abbrev=0 2>/dev/null || echo ""); \
	if [[ -z "$$LATEST_TAG" ]]; then \
		if [[ "$$TYPE" == "major" ]]; then \
			MAJOR=1; MINOR=0; PATCH=0; \
		else \
			MAJOR=0; MINOR=1; PATCH=0; \
		fi; \
		OLD_TAG="<none>"; \
	else \
		OLD_TAG="$$LATEST_TAG"; \
		VERSION="$${LATEST_TAG#v}"; \
		IFS='.' read -r MAJOR MINOR PATCH <<< "$$VERSION"; \
		if [[ "$$TYPE" == "major" ]]; then \
			MAJOR=$$((MAJOR + 1)); \
			MINOR=0; \
			PATCH=0; \
		elif [[ "$$TYPE" == "minor" ]]; then \
			MINOR=$$((MINOR + 1)); \
			PATCH=0; \
		else \
			PATCH=$$((PATCH + 1)); \
		fi; \
	fi; \
	NEW_TAG="v$${MAJOR}.$${MINOR}.$${PATCH}"; \
	git tag -a "$$NEW_TAG" -m "Release $$NEW_TAG"; \
	git push origin "$$NEW_TAG"; \
	echo "➜ bumped $${OLD_TAG} → $${NEW_TAG}"; \
	if command -v gh > /dev/null; then \
		echo "Creating GitHub release for $$NEW_TAG..."; \
		gh release create "$$NEW_TAG" --title "Release $$NEW_TAG" --notes "Release $$NEW_TAG"; \
		echo "✓ GitHub release created"; \
	else \
		echo "⚠️  GitHub CLI (gh) not found. GitHub release not created."; \
		echo "   Install GitHub CLI to enable automatic releases: https://cli.github.com/"; \
	fi

# Aliases for backward compatibility
version-minor:
	@$(MAKE) version TYPE=minor

version-major:
	@$(MAKE) version TYPE=major

# Clean up 
clean:
	rm -rf coverage.out coverage.html

# Show help
help:
	@echo "Available targets:"
	@echo "  test           - Run all tests"
	@echo "  test-coverage  - Run tests with coverage report"
	@echo "  lint           - Run golangci-lint"
	@echo "  fmt            - Format Go code"
	@echo "  vet            - Run go vet"
	@echo "  version        - Create a version, git tag and GitHub release"
	@echo "                   Usage:"
	@echo "                     make version           (patch version bump, default)"
	@echo "                     make version TYPE=patch (patch version bump)"
	@echo "                     make version TYPE=minor (minor version bump)"
	@echo "                     make version TYPE=major (major version bump)"
	@echo "  version-minor  - Alias for 'make version TYPE=minor'"
	@echo "  version-major  - Alias for 'make version TYPE=major'"
	@echo "  clean          - Clean up generated files"
	@echo "  help           - Show this help message"
	@echo ""
	@echo "Note: GitHub releases require the GitHub CLI (gh) to be installed."
	@echo "      Install from: https://cli.github.com/"