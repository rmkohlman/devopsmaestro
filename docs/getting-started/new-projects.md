# Starting New Projects

Create a new project from scratch with DevOpsMaestro.

---

## Overview

When starting a new project with dvm:

1. **Create your project directory** - Standard project setup
2. **Initialize your code** - git, package manager, etc.
3. **Add to dvm** - Track with DevOpsMaestro
4. **Build & attach** - Start coding in a container

---

## Step-by-Step Guide

### 1. Initialize dvm (one-time)

If you haven't already:

```bash
dvm init
```

### 2. Create Your Project Directory

```bash
mkdir ~/Developer/my-new-project
cd ~/Developer/my-new-project
```

### 3. Initialize Your Code

Set up your project as you normally would:

=== "Go"

    ```bash
    git init
    go mod init github.com/myuser/my-new-project
    touch main.go
    ```

=== "Python"

    ```bash
    git init
    python -m venv venv
    touch requirements.txt
    touch main.py
    ```

=== "Node.js"

    ```bash
    git init
    npm init -y
    touch index.js
    ```

=== "Rust"

    ```bash
    cargo init
    # or: cargo new my-new-project
    ```

### 4. Add to dvm

```bash
dvm create project my-new-project --from-cwd
```

Add a description to help remember what it's for:

```bash
dvm create project my-new-project --from-cwd --description "New REST API service"
```

### 5. Set Active Context

```bash
dvm use project my-new-project
```

### 6. Create a Workspace

```bash
dvm create workspace dev
dvm use workspace dev
```

### 7. Build the Container

```bash
dvm build
```

dvm auto-detects your language and sets up:

- Language runtime and tools
- LSP servers for Neovim
- Common development utilities

### 8. Start Coding

```bash
dvm attach
```

You're now in your containerized environment!

---

## Complete Examples

### New Go Project

```bash
# Create and setup project
mkdir ~/Developer/go-api
cd ~/Developer/go-api
git init
go mod init github.com/myuser/go-api

# Create initial file
cat > main.go << 'EOF'
package main

import "fmt"

func main() {
    fmt.Println("Hello, DevOpsMaestro!")
}
EOF

# Add to dvm
dvm create proj go-api --from-cwd
dvm use proj go-api
dvm create ws dev
dvm use ws dev

# Build and attach
dvm build
dvm attach

# Inside container:
# go run main.go
# nvim main.go  (with gopls LSP!)
```

### New Python Project

```bash
# Create and setup project
mkdir ~/Developer/python-app
cd ~/Developer/python-app
git init

# Create initial files
cat > requirements.txt << 'EOF'
fastapi>=0.100.0
uvicorn>=0.23.0
EOF

cat > main.py << 'EOF'
from fastapi import FastAPI

app = FastAPI()

@app.get("/")
def read_root():
    return {"Hello": "DevOpsMaestro"}
EOF

# Add to dvm
dvm create proj python-app --from-cwd
dvm use proj python-app
dvm create ws dev
dvm use ws dev

# Build and attach
dvm build
dvm attach

# Inside container:
# pip install -r requirements.txt
# uvicorn main:app --reload
```

### New Node.js Project

```bash
# Create and setup project
mkdir ~/Developer/node-app
cd ~/Developer/node-app
git init
npm init -y

# Create initial file
cat > index.js << 'EOF'
const express = require('express');
const app = express();
const port = 3000;

app.get('/', (req, res) => {
  res.json({ message: 'Hello, DevOpsMaestro!' });
});

app.listen(port, () => {
  console.log(`Server running at http://localhost:${port}`);
});
EOF

npm install express

# Add to dvm
dvm create proj node-app --from-cwd
dvm use proj node-app
dvm create ws dev
dvm use ws dev

# Build and attach
dvm build
dvm attach

# Inside container:
# node index.js
# nvim index.js  (with TypeScript LSP!)
```

---

## Project Templates (Coming Soon)

!!! note "Future Feature"
    
    Project templates will allow you to scaffold new projects with pre-configured setups:
    
    ```bash
    # Future syntax
    dvm create project my-api --template go-api
    dvm create project my-app --template python-fastapi
    ```

---

## Tips for New Projects

### Use Descriptive Names

```bash
# Good
dvm create proj user-auth-service --from-cwd
dvm create proj data-pipeline --from-cwd

# Less helpful
dvm create proj app1 --from-cwd
```

### Add Descriptions

```bash
dvm create proj my-api --from-cwd --description "REST API for user management"
```

### Create Purpose-Specific Workspaces

```bash
# Main development
dvm create ws dev

# For testing
dvm create ws test --description "Running test suites"

# For debugging
dvm create ws debug --description "Debugging with extra tools"
```

---

## Verify Your Setup

```bash
# Check everything is configured
dvm get ctx
dvm status

# View project details
dvm get project my-new-project -o yaml
```

---

## Next Steps

- [Building & Attaching](../dvm/build-attach.md) - Learn about container lifecycle
- [Configuring Workspaces](../configuration/yaml-schema.md) - Customize your environment
- [nvp Plugins](../nvp/plugins.md) - Set up Neovim plugins
