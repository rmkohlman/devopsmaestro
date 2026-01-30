#!/bin/bash
#
# create-test-plugin.sh - Creates test plugin YAML for Part 3.2
#

cat > /tmp/test-plugin.yaml << 'EOF'
apiVersion: nvp.io/v1
kind: NvimPlugin
metadata:
  name: my-plugin
  description: Test plugin for verification
  category: testing
  tags: ["test", "manual"]
spec:
  repo: test/my-plugin
  branch: main
  event: VeryLazy
  config: |
    require("my-plugin").setup({
      enabled = true,
    })
EOF

echo "Created /tmp/test-plugin.yaml"
echo ""
echo "Now run:"
echo "  ./nvp apply -f /tmp/test-plugin.yaml"
