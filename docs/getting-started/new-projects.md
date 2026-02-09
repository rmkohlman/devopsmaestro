# Starting New Apps

Create a new app from scratch with DevOpsMaestro.

---

## Overview

When starting a new app with dvm:

1. **Create your app directory** - Standard app setup
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

### 2. Create Your App Directory

```bash
mkdir ~/Developer/my-new-app
cd ~/Developer/my-new-app
```

### 3. Initialize Your Code

Set up your app as you normally would:

=== "Go"

    ```bash
    git init
    go mod init github.com/myuser/my-new-app
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
dvm create app my-new-app --from-cwd
```

Add a description to help remember what it's for:

```bash
dvm create app my-new-app --from-cwd --description "New REST API service"
```

### 5. Set Active Context

```bash
dvm use app my-new-app
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

### New Go App

```bash
# Create and setup app
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
dvm create app go-api --from-cwd
dvm use app go-api
dvm create ws dev
dvm use ws dev

# Build and attach
dvm build
dvm attach

# Inside container:
# go run main.go
# nvim main.go  (with gopls LSP!)
```

### New Python App

```bash
# Create and setup app
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
dvm create app python-app --from-cwd
dvm use app python-app
dvm create ws dev
dvm use ws dev

# Build and attach
dvm build
dvm attach

# Inside container:
# pip install -r requirements.txt
# uvicorn main:app --reload
```

### New Node.js App

```bash
# Create and setup app
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
dvm create app node-app --from-cwd
dvm use app node-app
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

## App Templates (Coming Soon)

!!! note "Future Feature"
    
    App templates will allow you to scaffold new apps with pre-configured setups:
    
    ```bash
    # Future syntax
    dvm create app my-api --template go-api
    dvm create app my-app --template python-fastapi
    ```

---

## Tips for New Apps

### Use Descriptive Names

```bash
# Good
dvm create app user-auth-service --from-cwd
dvm create app data-pipeline --from-cwd

# Less helpful
dvm create app app1 --from-cwd
```

### Add Descriptions

```bash
dvm create app my-api --from-cwd --description "REST API for user management"
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

# View app details
dvm get app my-new-app -o yaml
```

---

## Next Steps

- [Building & Attaching](../dvm/build-attach.md) - Learn about container lifecycle
- [Configuring Workspaces](../configuration/yaml-schema.md) - Customize your environment
- [nvp Plugins](../nvp/plugins.md) - Set up Neovim plugins
