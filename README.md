# DevOpsMaestro
## Description
 
### Overview

DevOpsMaestro is designed to facilitate seamless development inside containers by integrating Neovim directly within the containerized environment. This approach ensures that developers can work in isolated, reproducible environments without affecting their local machine. Each workspace created by DevOpsMaestro runs inside a Docker container, with Neovim pre-installed and configured according to the project's needs. This allows for a consistent development setup across different machines and team members.

### Neovim Configuration and Management

Neovim configurations within DevOpsMaestro are managed using decoupled YAML files, allowing for a custom setup for each workspace. These YAML files define the Neovim environment, plugins, key mappings, and other settings. When a workspace is created, DevOpsMaestro reads the corresponding YAML configuration file and applies it to the Neovim setup inside the container. This allows each workspace to have its unique development environment tailored to the specific needs of the project, without affecting other workspaces or the local machine.


DevOpsMaestro is a CLI tool designed to streamline the management of development environments, including scaffolding projects, managing dependencies, and facilitating GitOps practices. It allows developers to quickly spin up isolated workspaces, manage their dependencies, and create consistent development environments using a GitOps approach. The tool integrates with Docker, Kubernetes, and leverages technologies like Neovim, Viper, and Cobra to offer a powerful yet flexible development workflow.

DevOpsMaestro is a comprehensive CLI tool for managing development environments, testing, and deployments. It is free for personal and individual use under the GNU General Public License v3.0 (GPL-3.0). For corporate or business use, a commercial license is required.

## Table of Contents

- [Project Title](#devopsmaestro)
- [Description](#description)
- [Table of Contents](#table-of-contents)
- [Features](#features)
- [Installation](#installation)
- [Usage](#usage)
  - [Commands](#commands)
  - [Examples](#examples)
- [Configuration](#configuration)
  - [Environment Variables](#environment-variables)
  - [Configuration Files](#configuration-files)
- [Database Setup](#database-setup)
- [Backup and Restore](#backup-and-restore)
- [License](#license)
- [Contributing](#contributing)
- [Authors](#authors)
- [Acknowledgements](#acknowledgements)

## Features

- Scaffold new projects with pre-configured environments.
- Manage Docker and Kubernetes environments seamlessly.
- Integration with Neovim for coding inside containers.
- GitOps-based configuration management.
- Backup and restore project states using YAML files.
- Supports various databases (SQLite, PostgreSQL).
- Dynamic and isolated workspaces
- Automated task execution with reusable templates
- Centralized data management system
- Integration with Docker and Kubernetes
- Structured approach to handling projects, workspaces, dependencies, and advanced features like workflows and pipelines

## Getting Started

## Installation

### Prerequisites

- [Go 1.16+](https://golang.org/dl/)
- [Docker](https://www.docker.com/get-started)
- [Kubernetes](https://kubernetes.io/docs/setup/)
- [Neovim](https://neovim.io/)

### Steps

1. Clone the repository:
   ```bash
   git clone https://github.com/rmkohlman/devopsmaestro.git
   cd devopsmaestro



	2.	Install dependencies:
    go mod download
 
 3.	Build the CLI tool:
  go build -o dvm

4.	(Optional) Install the tool globally:
  sudo mv dvm /usr/local/bin/

	5.	Verify installation:
    dvm --help


Instructions for setting up the project, including any prerequisites, and steps for installation.

### 6. Usage

```markdown
## Usage

### Commands

- `dvm create project <name>`: Create a new project.
- `dvm create workspace <name>`: Create a new workspace within a project.
- `dvm list projects`: List all projects.
- `dvm admin migrate`: Apply database migrations.
- `dvm admin backup`: Create a backup of the current database state.
- `dvm apply -f <file>`: Restore a database state from a backup file.

### Examples

- **Create a new project**:
  ```bash
  dvm create project myproject

•	Backup the current state:
  dvm admin backup

•	Restore from a backup:
  dvm apply -f backups/backup_20230729_123456.yaml


Instructions on how to use the tool, including a list of commands and examples of how to execute them.

### 7. Configuration

```markdown
## Configuration

### Environment Variables

- `DVM_DATABASE`: The type of database to use (e.g., `SQLITE`, `POSTGRES`).
- `DVM_DATABASE_PATH`: The path to the SQLite database file (for SQLite).
- `DVM_DATABASE_HOST`: Database host (for PostgreSQL).
- `DVM_DATABASE_PORT`: Database port (for PostgreSQL).
- `DVM_DATABASE_USERNAME`: Database username (for PostgreSQL).
- `DVM_DATABASE_PASSWORD`: Database password (for PostgreSQL).

### Configuration Files

You can optionally configure DevOpsMaestro using a configuration file.

Example `config.yaml`:

```yaml
database:
  type: postgres
  host: localhost
  port: 5432
  username: myuser
  password: mypassword
  name: devopsmaestro



Details on configuring the project, including environment variables and configuration files.

### 8. Database Setup

```markdown
## Database Setup

Before using the tool, you need to set up the database.

1. Initialize the database:
   ```bash
   dvm admin init

2.	Apply database migrations:
  dvm admin migrate


Instructions for setting up and initializing the database.

### 9. Backup and Restore

```markdown
## Backup and Restore

### Backup

To create a backup of the current database state:

```bash
dvm admin backup


To restore from a backup:
  dvm apply -f backups/backup_20230729_123456.yaml

  This will restore the database to the state captured in the specified backup file.

Guidance on how to use backup and restore features.

### 10. License

```markdown
## License

This project is licensed under the GNU General Public License v3.0. See the [LICENSE](LICENSE) file for details.


### DevOpsMaestro Objects Overview

DevOpsMaestro consists of several key objects that manage different aspects of development environments, dependencies, and workflows. Each object plays a specific role in ensuring a seamless and organized development process.

#### 1. Project
- **Description**: A project is the top-level entity in DevOpsMaestro. It contains one or more workspaces and dependencies, representing a complete development environment. Projects help organize different development efforts and manage resources effectively.

#### 2. Workspace
- **Description**: A workspace is an isolated development environment within a project. It runs inside a Docker container and is configured with a specific programming language and development tools, including Neovim. Each workspace can have its own set of dependencies and a custom Neovim configuration managed through YAML files.

#### 3. Dependency
- **Description**: Dependencies are external services or libraries that a workspace needs to function. These could include databases, message queues, or any other external service required by the workspace. Dependencies are linked to both workspaces and projects, allowing for shared usage across multiple workspaces.

#### 4. Task
- **Description**: A task is a single, discrete action that can be performed within a workspace or project. Tasks might include running tests, building code, or deploying a service. Tasks can be combined into workflows for more complex operations.

#### 5. Workflow
- **Description**: A workflow is a series of tasks that are executed in a specific order to achieve a particular goal, such as deploying an application or running a suite of tests. Workflows help automate repetitive sequences of actions within a development or deployment pipeline.

#### 6. Pipeline
- **Description**: A pipeline is a higher-level construct that organizes workflows into a cohesive process, typically representing the full cycle from code development to deployment. Pipelines are essential for managing continuous integration and continuous deployment (CI/CD) processes.

#### 7. Orchestration
- **Description**: Orchestration refers to the management and coordination of multiple pipelines, ensuring that they run in the correct order and handle dependencies between them. Orchestration is crucial for complex applications with multiple interdependent components.

#### 8. Prototype
- **Description**: A prototype is a reusable template for context objects, tasks, workflows, or any other part of the system. Prototypes allow for quick creation and customization of environments, ensuring consistency across different projects and workspaces.

#### 9. Data Lake
- **Description**: The Data Lake is a global storage system that holds large datasets used across multiple projects or workspaces. It is typically used for shared resources that need to be accessed by various parts of the system.

#### 10. Data Store
- **Description**: A Data Store is a project-specific storage area for managing data related to a particular project. It is used to keep track of project-related information, such as configuration settings, logs, or temporary data.

#### 11. Data Record
- **Description**: A Data Record is the smallest unit of data storage within a workspace. It is typically associated with a specific task or dependency and is used to store configuration details, state information, or small datasets.

#### 12. Context
- **Description**: A context is a YAML-based configuration file that defines settings for projects, workspaces, dependencies, and other objects. Contexts allow for flexible and decoupled configuration management, enabling custom setups for different environments.

#### 13. Storage
- **Description**: Storage refers to the volumes attached to projects, workspaces, and dependencies. Storage volumes hold persistent data that needs to be preserved across container restarts or shared between different components within a project.


# DevOpsMaestro CLI Commands

### dvm create project `<project-name>`
- **Description**: Creates a new project in DevOpsMaestro.

### dvm create workspace `<workspace-name>` --project `<project-name>` --language `<language>`
- **Description**: Creates a new workspace within a specified project, allowing you to define the programming language used in the workspace.

### dvm create dependency `<dependency-name>` --project `<project-name>` --workspace `<workspace-name>`
- **Description**: Adds a new dependency to a specified workspace within a project.

### dvm list projects
- **Description**: Lists all projects managed by DevOpsMaestro.

### dvm list workspaces --project `<project-name>`
- **Description**: Lists all workspaces under a specified project.

### dvm list dependencies --project `<project-name>` --workspace `<workspace-name>`
- **Description**: Lists all dependencies for a specified workspace within a project.

### dvm use project `<project-name>`
- **Description**: Switches the context to the specified project, making it the active project for subsequent commands.

### dvm use workspace `<workspace-name>` --project `<project-name>`
- **Description**: Switches the context to the specified workspace within a project, making it the active workspace for subsequent commands.

### dvm release project
- **Description**: Releases the current project context, returning to the global context.

### dvm release workspace
- **Description**: Releases the current workspace context, returning to the project context.

### dvm admin init
- **Description**: Initializes the database and prepares the environment for first-time use.

### dvm admin migrate
- **Description**: Applies database migrations to ensure the schema is up-to-date.

### dvm admin migrate --backup
- **Description**: Applies database migrations with an automatic backup before applying changes.

### dvm admin snapshot
- **Description**: Creates a snapshot of the current database state and stores it as a YAML file.

### dvm admin backup
- **Description**: Creates a backup of the current database state and stores it as a YAML file.

### dvm apply -f `<file>`
- **Description**: Restores the database state from a specified YAML file.

### dvm get project `<project-name>` --output [yaml|json]
- **Description**: Retrieves detailed information about a specified project in YAML or JSON format.

### dvm get workspace `<workspace-name>` --project `<project-name>` --output [yaml|json]
- **Description**: Retrieves detailed information about a specified workspace in a project in YAML or JSON format.

### dvm get dependency `<dependency-name>` --workspace `<workspace-name>` --project `<project-name>` --output [yaml|json]
- **Description**: Retrieves detailed information about a specified dependency in a workspace in YAML or JSON format.

### dvm get context `<context-name>` --workspace `<workspace-name>` --project `<project-name>` --output [yaml|json]
- **Description**: Retrieves detailed information about a specific context within a workspace or project in YAML or JSON format.

### dvm delete project `<project-name>`
- **Description**: Deletes a specified project and all its associated workspaces and dependencies.

### dvm delete workspace `<workspace-name>` --project `<project-name>`
- **Description**: Deletes a specified workspace within a project, including all its dependencies.

### dvm delete dependency `<dependency-name>` --workspace `<workspace-name>` --project `<project-name>`
- **Description**: Deletes a specified dependency within a workspace in a project.

### dvm attach workspace `<workspace-name>` --project `<project-name>`
- **Description**: Attaches to a specified workspace, allowing you to start working within that environment.

### dvm detach workspace `<workspace-name>` --project `<project-name>`
- **Description**: Detaches from the specified workspace, stopping the associated environment.

### dvm list storage --project `<project-name>` --workspace `<workspace-name>`
- **Description**: Lists all storage volumes associated with the specified project and workspace.

### dvm reset storage --project `<project-name>` --workspace `<workspace-name>`
- **Description**: Resets the storage volumes for a specified project and workspace.

### dvm list context --project `<project-name>` --workspace `<workspace-name>`
- **Description**: Lists all context objects associated with the specified project and workspace.

### Use Commands

### dvm use project `<project-name>`
- **Description**: Switches the context to the specified project, making it the active project for subsequent commands.

### dvm use project none
- **Description**: Clears the current project context, returning to the global context where no specific project is active.

### dvm use workspace `<workspace-name>` --project `<project-name>`
- **Description**: Switches the context to the specified workspace within a project, making it the active workspace for subsequent commands.

### dvm use workspace none
- **Description**: Clears the current workspace context, returning to the project context without any specific workspace being active.

### Release Commands

### dvm release project
- **Description**: Releases the current project context, returning to the global context.

### dvm release workspace
- **Description**: Releases the current workspace context, returning to the project context.

### dvm release --all
- **Description**: Releases both the current project and workspace contexts, returning to the global context with no active project or workspace.



## Configuration Management

DevOpsMaestro uses GitOps principles for configuration management. You can export and import configurations via YAML files for easy environment setup and sharing.

## Neovim Integration

DevOpsMaestro allows you to configure Neovim setups using `init.lua` files managed as part of the workspace context. This provides a consistent and portable development environment.

## Contributing

Contributions are welcome! Please read the [CONTRIBUTING.md](CONTRIBUTING.md) file for guidelines on how to contribute to this project.

## Authors

- **Robert Kohlman** - *Initial work* - [rmkohlman](https://github.com/rmkohlman)

## Acknowledgements

- Special thanks to [OpenAI](https://www.openai.com/) for the guidance.
- Inspiration from [kubectl](https://kubernetes.io/docs/reference/kubectl/overview/).

## License

DevOpsMaestro is licensed under the GNU General Public License v3.0. See the [LICENSE](LICENSE) file for more details.

For commercial use, please refer to the [LICENSE-COMMERCIAL.txt](LICENSE-COMMERCIAL.txt) file.

---

This `README.md` file provides a comprehensive overview of the DevOpsMaestro project, including installation instructions, usage examples, and details on each command. Feel free to update and customize it further as needed.