# Creating New Projects

Start a new project from scratch with DevOpsMaestro and get a fully-configured development environment instantly.

---

## Overview

When starting a new project with DevOpsMaestro:

1. **Create your project directory** - Standard app setup
2. **Initialize your codebase** - git, language setup, dependencies
3. **Add to DevOpsMaestro** - Track with the hierarchy
4. **Build & develop** - Start coding in a containerized environment

The key advantage: **DevOpsMaestro gives you a production-ready development environment from day one**, with LSP servers, debuggers, linters, and Neovim configured automatically.

---

## Language-Specific Quickstarts

### Go API Project

Perfect for REST APIs, CLI tools, or microservices:

```bash
# 1. Create project directory
mkdir ~/Developer/go-user-api
cd ~/Developer/go-user-api

# 2. Initialize Go module
go mod init github.com/yourorg/go-user-api

# 3. Create initial API code
cat > main.go << 'EOF'
package main

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"
)

type User struct {
    ID   int    `json:"id"`
    Name string `json:"name"`
}

func main() {
    http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
        users := []User{
            {ID: 1, Name: "Alice"},
            {ID: 2, Name: "Bob"},
        }
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(users)
    })

    fmt.Println("Server starting on :8080")
    log.Fatal(http.ListenAndServe(":8080", nil))
}
EOF

# 4. Add to DevOpsMaestro
dvm init  # Only needed once per machine
dvm create ecosystem mycompany
dvm create domain backend-services  
dvm create app go-user-api --from-cwd --description "User management API"
dvm create workspace dev

# 5. Build and develop
dvm build
dvm attach

# Inside container:
# go run main.go &
# curl http://localhost:8080/users
# nvim main.go  # Full Go LSP with gopls!
```

**What you get:**
- Go 1.25+ with all standard tools
- `gopls` LSP server for autocompletion, error detection  
- `dlv` debugger pre-configured
- Neovim with Go syntax highlighting and plugins
- Air for hot reloading (if you add it to go.mod)

### Python FastAPI Project  

Perfect for web APIs and data science services:

```bash
# 1. Create project directory
mkdir ~/Developer/python-fastapi-app
cd ~/Developer/python-fastapi-app

# 2. Initialize Python project
cat > requirements.txt << 'EOF'
fastapi>=0.100.0
uvicorn[standard]>=0.23.0
pydantic>=2.0.0
EOF

cat > main.py << 'EOF'
from fastapi import FastAPI
from pydantic import BaseModel

app = FastAPI(title="User API", version="1.0.0")

class User(BaseModel):
    id: int
    name: str
    email: str

# In-memory storage for demo
users = [
    User(id=1, name="Alice", email="alice@example.com"),
    User(id=2, name="Bob", email="bob@example.com"),
]

@app.get("/")
def read_root():
    return {"message": "Welcome to User API"}

@app.get("/users", response_model=list[User])
def get_users():
    return users

@app.post("/users", response_model=User)
def create_user(user: User):
    users.append(user)
    return user

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
EOF

# 3. Add to DevOpsMaestro
dvm init  # Only needed once
dvm create ecosystem personal
dvm create domain webapps
dvm create app python-fastapi-app --from-cwd --description "FastAPI web service"
dvm create workspace dev

# 4. Build and develop
dvm build
dvm attach

# Inside container:
# pip install -r requirements.txt
# python main.py &
# curl http://localhost:8000/users
# nvim main.py  # Full Python LSP with pylsp!
```

**What you get:**
- Python 3.12+ with pip and venv
- `pylsp` (Python LSP Server) for autocompletion
- `black` formatter and `flake8` linter
- Neovim with Python syntax and autocompletion
- FastAPI development server with auto-reload

### Node.js/TypeScript Project

Perfect for web applications and APIs:

```bash
# 1. Create project directory
mkdir ~/Developer/node-typescript-api
cd ~/Developer/node-typescript-api

# 2. Initialize Node.js project
npm init -y
npm install express @types/express typescript ts-node nodemon
npm install --save-dev @types/node

# 3. Setup TypeScript
cat > tsconfig.json << 'EOF'
{
  "compilerOptions": {
    "target": "ES2022",
    "module": "commonjs",
    "outDir": "./dist",
    "rootDir": "./src",
    "strict": true,
    "esModuleInterop": true,
    "skipLibCheck": true,
    "forceConsistentCasingInFileNames": true
  },
  "include": ["src/**/*"],
  "exclude": ["node_modules", "dist"]
}
EOF

# 4. Create source structure
mkdir src
cat > src/app.ts << 'EOF'
import express from 'express';

interface User {
  id: number;
  name: string;
  email: string;
}

const app = express();
const port = 3000;

app.use(express.json());

// In-memory storage for demo
const users: User[] = [
  { id: 1, name: 'Alice', email: 'alice@example.com' },
  { id: 2, name: 'Bob', email: 'bob@example.com' },
];

app.get('/', (req, res) => {
  res.json({ message: 'Welcome to TypeScript API' });
});

app.get('/users', (req, res) => {
  res.json(users);
});

app.post('/users', (req, res) => {
  const user: User = req.body;
  users.push(user);
  res.status(201).json(user);
});

app.listen(port, () => {
  console.log(`Server running at http://localhost:${port}`);
});
EOF

# 5. Update package.json scripts
npm pkg set scripts.start="node dist/app.js"
npm pkg set scripts.dev="nodemon --exec ts-node src/app.ts"
npm pkg set scripts.build="tsc"

# 6. Add to DevOpsMaestro
dvm init
dvm create ecosystem personal  
dvm create domain webapps
dvm create app node-typescript-api --from-cwd --description "TypeScript Express API"
dvm create workspace dev

# 7. Build and develop
dvm build
dvm attach

# Inside container:
# npm run dev &
# curl http://localhost:3000/users  
# nvim src/app.ts  # Full TypeScript LSP!
```

**What you get:**
- Node.js 20+ with npm
- TypeScript compiler and `typescript-language-server`
- ESLint and Prettier pre-configured
- Neovim with TypeScript autocompletion and error detection
- Hot reloading with nodemon

### Rust CLI Project

Perfect for system tools and high-performance applications:

```bash
# 1. Create project with Cargo
cargo new --bin rust-cli-tool
cd rust-cli-tool

# 2. Add dependencies
cat >> Cargo.toml << 'EOF'

[dependencies]
clap = { version = "4.4", features = ["derive"] }
serde = { version = "1.0", features = ["derive"] }
serde_json = "1.0"
tokio = { version = "1.0", features = ["full"] }
EOF

# 3. Create CLI application
cat > src/main.rs << 'EOF'
use clap::{Arg, Command};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;

#[derive(Serialize, Deserialize, Debug)]
struct Config {
    name: String,
    version: String,
    features: Vec<String>,
}

#[tokio::main]
async fn main() {
    let matches = Command::new("rust-cli-tool")
        .version("1.0.0")
        .author("Your Name")
        .about("A sample Rust CLI tool")
        .subcommand(
            Command::new("info")
                .about("Show application info")
        )
        .subcommand(
            Command::new("config")
                .about("Show configuration")
                .arg(Arg::new("format")
                    .short('f')
                    .long("format")
                    .value_name("FORMAT")
                    .help("Output format (json|yaml)")
                )
        )
        .get_matches();

    match matches.subcommand() {
        Some(("info", _)) => {
            println!("Rust CLI Tool v1.0.0");
            println!("Built with DevOpsMaestro!");
        }
        Some(("config", sub_matches)) => {
            let config = Config {
                name: "rust-cli-tool".to_string(),
                version: "1.0.0".to_string(),
                features: vec!["async".to_string(), "json".to_string()],
            };

            let format = sub_matches.get_one::<String>("format").unwrap_or(&"json".to_string());
            
            match format.as_str() {
                "json" => println!("{}", serde_json::to_string_pretty(&config).unwrap()),
                _ => println!("{:#?}", config),
            }
        }
        _ => {
            println!("Use --help for usage information");
        }
    }
}
EOF

# 4. Add to DevOpsMaestro
dvm init
dvm create ecosystem personal
dvm create domain systems
dvm create app rust-cli-tool --from-cwd --description "High-performance CLI tool"
dvm create workspace dev

# 5. Build and develop
dvm build  
dvm attach

# Inside container:
# cargo run -- info
# cargo run -- config --format json
# nvim src/main.rs  # Full Rust LSP with rust-analyzer!
```

**What you get:**
- Rust stable with Cargo
- `rust-analyzer` LSP server for autocompletion and error detection
- `rustfmt` formatter and Clippy linter
- Neovim with Rust syntax highlighting and intelligent code completion
- Debug symbols enabled for `gdb`/`lldb` debugging

---

## Project Templates (Future Feature)

!!! note "Coming in v0.13.0"
    
    Project templates will allow you to scaffold new projects with pre-configured setups:
    
    ```bash
    # Future syntax
    dvm create project my-api --template go-rest-api
    dvm create project my-app --template python-fastapi  
    dvm create project my-tool --template rust-cli
    dvm create project my-web --template nextjs-app
    ```
    
    Templates will include:
    - Language-specific starter code
    - Best practice project structure  
    - CI/CD pipeline configurations
    - Testing frameworks pre-configured
    - Documentation templates

---

## Advanced Project Setups

### Microservices Project (Multiple Apps)

Set up multiple related services:

```bash
# 1. Create ecosystem and domain
dvm create ecosystem ecommerce
dvm create domain backend-services

# 2. Create multiple apps
mkdir ~/Developer/ecommerce && cd ~/Developer/ecommerce

# User service
mkdir user-service && cd user-service
go mod init github.com/ecommerce/user-service
echo 'package main\n\nfunc main() { println("User Service") }' > main.go
dvm create app user-service --from-cwd --description "User management service"
dvm create workspace dev

# Product service  
cd .. && mkdir product-service && cd product-service
go mod init github.com/ecommerce/product-service
echo 'package main\n\nfunc main() { println("Product Service") }' > main.go
dvm create app product-service --from-cwd --description "Product catalog service"
dvm create workspace dev

# Order service
cd .. && mkdir order-service && cd order-service  
go mod init github.com/ecommerce/order-service
echo 'package main\n\nfunc main() { println("Order Service") }' > main.go
dvm create app order-service --from-cwd --description "Order processing service"
dvm create workspace dev

# 3. View your microservices
dvm get apps
# NAME            DOMAIN            CREATED
# user-service    backend-services  2024-02-19 10:00
# product-service backend-services  2024-02-19 10:01  
# order-service   backend-services  2024-02-19 10:02

# 4. Work on any service
dvm use app user-service
dvm build && dvm attach
```

### Full-Stack Project (Multiple Domains)

Set up frontend and backend together:

```bash
# 1. Create ecosystem
dvm create ecosystem myapp

# 2. Create backend domain and app
dvm create domain backend
mkdir ~/Developer/myapp-backend && cd ~/Developer/myapp-backend
go mod init github.com/myorg/myapp-backend  
echo 'package main\n\nfunc main() { println("Backend API") }' > main.go
dvm create app myapp-backend --from-cwd --description "REST API backend"
dvm create workspace dev

# 3. Create frontend domain and app  
dvm create domain frontend
mkdir ~/Developer/myapp-frontend && cd ~/Developer/myapp-frontend
npm init -y && npm install react typescript @types/react
dvm create app myapp-frontend --from-cwd --description "React frontend"
dvm create workspace dev

# 4. Switch between frontend and backend
dvm use domain backend && dvm use app myapp-backend
dvm build && dvm attach  # Go development environment

# In another terminal:
dvm use domain frontend && dvm use app myapp-frontend  
dvm build && dvm attach  # Node.js/React development environment
```

---

## Project Organization Best Practices

### Naming Conventions

=== "Professional/Enterprise"

    ```bash
    # Use company/team structure
    dvm create ecosystem acme-corp
    dvm create domain customer-platform
    dvm create app user-auth-service
    dvm create workspace dev

    dvm create domain data-platform  
    dvm create app analytics-pipeline
    dvm create workspace prod-debug
    ```

=== "Personal Projects"

    ```bash
    # Use simpler personal structure
    dvm create ecosystem personal
    dvm create domain webapps
    dvm create app my-blog
    dvm create workspace dev

    dvm create domain tools
    dvm create app backup-script  
    dvm create workspace testing
    ```

=== "Open Source"

    ```bash
    # Mirror GitHub organization
    dvm create ecosystem kubernetes  # GitHub org
    dvm create domain core-components
    dvm create app kubectl
    dvm create workspace feature-new-command
    ```

### Workspace Strategies

Create workspaces that match your development workflow:

```bash
# By purpose
dvm create workspace dev --description "Main development"
dvm create workspace test --description "Running test suites"  
dvm create workspace debug --description "Debugging issues"
dvm create workspace demo --description "Demo preparation"

# By feature branch
dvm create workspace feature-auth --description "Authentication feature"
dvm create workspace feature-payments --description "Payment integration"
dvm create workspace hotfix-security --description "Security patch"

# By environment
dvm create workspace local-dev --description "Local development"
dvm create workspace integration --description "Integration testing"
dvm create workspace perf-testing --description "Performance testing"
```

### Adding Descriptions

Always add descriptions to make your project discoverable:

```bash
# Good descriptions
dvm create app user-management-api --from-cwd \
  --description "REST API for user authentication and profile management"

dvm create app data-processor --from-cwd \
  --description "ETL pipeline for processing customer analytics data"

# View descriptions
dvm get apps -o yaml  # Shows descriptions in YAML output
```

---

## Development Workflow Examples

### Typical Day with New Project

```bash
# Morning: Start new feature
dvm use app my-project
dvm create workspace feature-notifications --description "Push notification system"
dvm use workspace feature-notifications
dvm build
dvm attach

# Inside container: Work on code
git checkout -b feature/push-notifications
nvim src/notifications.py  # Full LSP support
python -m pytest tests/    # Run tests

# Exit container when done
exit

# Evening: Switch back to main development
dvm use workspace dev
dvm attach
```

### Code Review Workflow

```bash
# Reviewer: Pull down PR branch in separate workspace
cd ~/Developer/team-project
git fetch origin pull/123/head:pr-123
git checkout pr-123

dvm create workspace review-pr-123 --description "Code review for PR #123"
dvm use workspace review-pr-123
dvm build && dvm attach

# Review code in isolation with fresh environment
nvim src/new-feature.py
python -m pytest tests/test_new_feature.py

# Clean up when done
exit
dvm delete workspace review-pr-123
```

---

## Troubleshooting New Projects

### Language Not Detected

```bash
# Check what dvm detected
dvm build --verbose

# Force language detection by creating proper files
# For Python: requirements.txt or pyproject.toml
# For Node.js: package.json  
# For Go: go.mod
# For Rust: Cargo.toml
```

### Missing Dependencies

```bash
# Add dependencies before building
# Go example:
go mod tidy

# Python example:  
echo "fastapi>=0.100.0" >> requirements.txt

# Node.js example:
npm install express

# Then rebuild
dvm rebuild
```

### Port Conflicts

```bash
# Check what's running
dvm get workspaces  # Shows container status

# Kill conflicting containers
dvm detach          # Exit current
docker ps           # Check Docker containers
docker stop <id>    # Stop conflicting container
```

---

## Next Steps

After creating your project:

1. **[Build & Attach Guide](../dvm/build-attach.md)** - Learn container lifecycle management
2. **[Workspace Configuration](../configuration/yaml-schema.md)** - Customize your development environment  
3. **[Theme System](../nvp/themes.md)** - Set up visual themes that match your preferences
4. **[Plugin Management](../nvp/plugins.md)** - Add Neovim plugins for your language
5. **[CI/CD Integration](../advanced/ci-cd.md)** - Use DevOpsMaestro in your deployment pipeline

---

## Cheat Sheet for New Projects

```bash
# Quick Go API
mkdir go-api && cd go-api && go mod init app
dvm create eco personal && dvm create dom apis && dvm create a go-api --from-cwd && dvm create ws dev
dvm build && dvm attach

# Quick Python Web App  
mkdir python-web && cd python-web && echo "fastapi>=0.100.0" > requirements.txt
dvm create eco personal && dvm create dom web && dvm create a python-web --from-cwd && dvm create ws dev
dvm build && dvm attach

# Quick Node.js App
mkdir node-app && cd node-app && npm init -y && npm install express
dvm create eco personal && dvm create dom web && dvm create a node-app --from-cwd && dvm create ws dev  
dvm build && dvm attach

# Quick Rust Tool
cargo new rust-tool && cd rust-tool
dvm create eco personal && dvm create dom tools && dvm create a rust-tool --from-cwd && dvm create ws dev
dvm build && dvm attach
```
