#!/bin/bash
#
# create-invalid-yaml-tests.sh - Creates invalid YAML files for Part 6.2
#

echo "Creating invalid YAML test files..."

# Wrong apiVersion
cat > /tmp/invalid-apiversion.yaml << 'EOF'
apiVersion: wrong/v1
kind: NvimPlugin
metadata:
  name: test
spec:
  repo: test/test
EOF
echo "  Created /tmp/invalid-apiversion.yaml"

# Wrong kind
cat > /tmp/invalid-kind.yaml << 'EOF'
apiVersion: nvp.io/v1
kind: WrongKind
metadata:
  name: test
spec:
  repo: test/test
EOF
echo "  Created /tmp/invalid-kind.yaml"

# Missing required fields
cat > /tmp/invalid-missing-repo.yaml << 'EOF'
apiVersion: nvp.io/v1
kind: NvimPlugin
metadata:
  name: test
spec: {}
EOF
echo "  Created /tmp/invalid-missing-repo.yaml"

echo ""
echo "Now test each one (all should fail):"
echo "  ./nvp apply -f /tmp/invalid-apiversion.yaml"
echo "  ./nvp apply -f /tmp/invalid-kind.yaml"
echo "  ./nvp apply -f /tmp/invalid-missing-repo.yaml"
