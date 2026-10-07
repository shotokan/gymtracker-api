# ==============================================================================
# GymTracker API — Makefile
# ==============================================================================

MODULE  := github.com/shotokan/gymtracker-api
REGION  := us-east-1
STACK   := GymTrackerStack
BIN_DIR := bin
CMDS    := auth authorizer user plan session

GO_BUILD_FLAGS := -tags lambda.norpc -ldflags="-s -w"
GO_ENV         := GOOS=linux GOARCH=arm64 CGO_ENABLED=0

.PHONY: help build test cover cover-html lint clean \
        infra-install synth diff \
        deploy deploy-hotswap deploy-local destroy \
        secret assume whoami all

# ── Help ──────────────────────────────────────────────────────────────────────

help:
	@echo ""
	@echo "GymTracker API"
	@echo "=============="
	@echo "  make build            Compile Lambda binaries for Linux ARM64"
	@echo "  make test             Run tests with race detector"
	@echo "  make cover            Run tests and print coverage percentages"
	@echo "  make cover-html       Generate HTML coverage report"
	@echo "  make lint             Run golangci-lint"
	@echo "  make clean            Remove build artifacts"
	@echo "  make deploy           Full deploy via CDK (requires Docker)"
	@echo "  make deploy-hotswap   Update Lambda code only (~15s)"
	@echo "  make deploy-local     Build locally then deploy (no Docker)"
	@echo "  make diff             Show infra changes before applying"
	@echo "  make destroy          Tear down the stack"
	@echo "  make secret           Create JWT secret in SSM"
	@echo ""

# ── Go: Build ─────────────────────────────────────────────────────────────────

# Compile all Lambda binaries for Linux ARM64 (Graviton2).
build:
	@echo "→ Compiling Lambda binaries for Linux ARM64..."
	@mkdir -p $(foreach cmd,$(CMDS),$(BIN_DIR)/$(cmd))
	@$(foreach cmd,$(CMDS), \
		echo "  building cmd/$(cmd)..." && \
		$(GO_ENV) go build $(GO_BUILD_FLAGS) \
			-o $(BIN_DIR)/$(cmd)/bootstrap ./cmd/$(cmd);)
	@echo "✓ Build complete"

# ── Go: Test ──────────────────────────────────────────────────────────────────

# Run all unit tests with race detector.
test:
	@echo "→ Running tests..."
	go test ./internal/... ./pkg/... -v -race -count=1
	@echo "✓ Tests passed"

# ── Go: Cover ─────────────────────────────────────────────────────────────────

# Run tests and print coverage percentage per package.
# Uses -cover (not -coverprofile) to avoid covdata issues on Go 1.20+.
cover:
	@echo "→ Running coverage report..."
	go test ./internal/... ./pkg/... -cover -race -count=1 || true
	@echo "✓ Coverage complete"

# Generate a full HTML coverage report.
# Only runs on packages that have test files.
cover-html:
	@echo "→ Generating HTML coverage report..."
	go test \
		./internal/authorizer/application/... \
		./internal/login/application/... \
		./pkg/... \
		-coverprofile=coverage.out -race -count=1
	go tool cover -func=coverage.out | tail -1
	go tool cover -html=coverage.out -o coverage.html
	@echo "✓ Open coverage.html in your browser"

# ── Go: Lint ──────────────────────────────────────────────────────────────────

# Run golangci-lint. Install: brew install golangci-lint
lint:
	@echo "→ Running linter..."
	golangci-lint run ./internal/... ./pkg/... ./cmd/...
	@echo "✓ Lint passed"

# ── Go: Clean ─────────────────────────────────────────────────────────────────

# Remove all generated files.
clean:
	@echo "→ Cleaning build artifacts..."
	rm -rf $(BIN_DIR)/ coverage.out coverage.html
	rm -rf infra/dist/ infra/cdk.out/
	@echo "✓ Clean complete"

# ── CDK ───────────────────────────────────────────────────────────────────────

infra-install:
	@echo "→ Installing CDK dependencies..."
	cd infra && npm install --silent
	@echo "✓ CDK dependencies installed"

synth: infra-install
	@echo "→ Synthesizing CloudFormation template..."
	cd infra && npx cdk synth

diff: build infra-install
	@echo "→ Comparing local changes with deployed stack..."
	cd infra && npx cdk diff $(STACK)

# Full deploy: CDK compiles Go via Docker + updates all AWS resources.
deploy: build infra-install
	@echo "→ Deploying stack to AWS..."
	cd infra && npx cdk deploy $(STACK) \
		--require-approval never \
		--region $(REGION)
	@echo "✓ Deploy complete"

# Fast deploy: Lambda code only, no CloudFormation, ~15 seconds.
deploy-hotswap: infra-install
	@echo "→ Hotswap deploy (Lambda code only)..."
	cd infra && npx cdk deploy $(STACK) \
		--hotswap \
		--require-approval never \
		--region $(REGION)
	@echo "✓ Hotswap complete"

# Deploy without Docker: build locally first, then upload via CDK.
deploy-local: build infra-install
	@echo "→ Deploying with locally-compiled binaries..."
	cd infra && npx cdk deploy $(STACK) \
		--require-approval never \
		--region $(REGION)
	@echo "✓ Deploy complete"

destroy:
	@echo "→ Destroying stack $(STACK)..."
	@echo "⚠  DynamoDB table will be RETAINED."
	cd infra && npx cdk destroy $(STACK) --force
	@echo "✓ Stack destroyed"

# ── Secrets ───────────────────────────────────────────────────────────────────

# Create the JWT signing secret in SSM. Run once before the first deploy.
secret:
	@echo "→ Creating JWT secret in SSM Parameter Store..."
	aws ssm put-parameter \
		--name  "/gym-tracker/jwt-secret" \
		--value "$$(openssl rand -base64 48)" \
		--type  "String" \
		--overwrite \
		--region $(REGION)
	@echo "✓ Secret stored at /gym-tracker/jwt-secret"

# ── Pipeline ──────────────────────────────────────────────────────────────────

all: lint test cover deploy
