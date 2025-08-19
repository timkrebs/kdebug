#!/bin/bash

# Setup Git hooks for kdebug project
# This script installs pre-push hooks to run validation automatically

set -e

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}🔧 Setting up Git hooks for kdebug...${NC}\n"

# Check if we're in a git repository
if [ ! -d ".git" ]; then
    echo -e "${YELLOW}⚠️  Not in a Git repository. Please run from project root.${NC}"
    exit 1
fi

# Create hooks directory if it doesn't exist
mkdir -p .git/hooks

# Create pre-push hook
cat > .git/hooks/pre-push << 'EOF'
#!/bin/bash

# Pre-push hook for kdebug
# Runs validation checks before allowing push

echo "🔍 Running pre-push validation..."

# Check if pre-push script exists
if [ ! -f "scripts/pre-push.sh" ]; then
    echo "❌ Pre-push script not found at scripts/pre-push.sh"
    exit 1
fi

# Run the pre-push validation
if ./scripts/pre-push.sh; then
    echo "✅ Pre-push validation passed. Proceeding with push."
    exit 0
else
    echo "❌ Pre-push validation failed. Push blocked."
    echo "💡 To skip validation (not recommended): git push --no-verify"
    exit 1
fi
EOF

# Make the hook executable
chmod +x .git/hooks/pre-push

echo -e "${GREEN}✅ Pre-push hook installed successfully!${NC}"
echo -e "${GREEN}The hook will run automatically before each git push.${NC}\n"

echo -e "${BLUE}How it works:${NC}"
echo -e "  • Runs before every 'git push'"
echo -e "  • Validates code formatting, tests, linting, and security"
echo -e "  • Blocks push if any checks fail"
echo -e "  • Can be bypassed with 'git push --no-verify' (not recommended)\n"

echo -e "${BLUE}To test the hook:${NC}"
echo -e "  make pre-push      # Run validation manually"
echo -e "  git push origin    # Will trigger automatic validation\n"

echo -e "${BLUE}To disable the hook:${NC}"
echo -e "  rm .git/hooks/pre-push\n"
