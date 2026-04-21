# Pokunuvita

**Pokunuvita** is an automated code-generation service designed to streamline the PR workflow. By providing a repository and a set of instructions, the system orchestrates an isolated environment to execute logic, commit changes, and submit pull requests.

## How it Works
1. **Targeting:** Specify the repository and the required modifications.
2. **Orchestration:** The service spins up a dedicated Docker container.
3. **Execution:** It runs `opencode` within the container to perform the requested tasks.
4. **Integration:** Changes are automatically pushed, and a Pull Request is generated for review.

---

## LinkedIn Live Stream Strategy

### Optimized Captions & Hooks

**Caption:** [Live Build] Can we automate a full PR workflow in Go? Let’s find out.

**Post:** I’m building an automated orchestration service that takes a simple prompt and turns it into a Pull Request. It spins up an isolated Docker environment, runs `opencode` to generate the logic, and pushes the changes directly to your repo. No manual setup, just clean code delivered via PR.

**The Goal for this stream:**
* One...

#Golang #BuildInPublic #SoftwareEngineering #KualaLumpur #KLTech

---

## Development Log

### Mar 22, 2026
* Installed `opencode` on macOS environment.
* Successfully initialized the `opencode` server.
* Established a connection via the **Go SDK**.
* Verified end-to-end communication by passing prompts and logging full responses.

### Mar 23, 2026
* Implemented session mapping for `opencode` to manage multiple streams.
* Optimized response handling to stream text parts (improving perceived latency).
* Initiated Docker image construction for `opencode` (including repository cloning logic).
* Began testing container orchestration using **Docker Compose**.

### Mar 24, 2026
* Finalized the Docker image and verified stable runtime.
* Successfully connected to the `opencode` instance running within the Docker container for Go-based tasks.
* Confirmed file-system integrity by verifying directory and file structures via the prompt interface.

### Mar 25, 2026
* Implemented full container lifecycle management (Create, Build, and Destroy) using the **Docker Go SDK**.
* Established connectivity to `opencode` instances running within containers provisioned via the Docker client library.

## Mar 26, 2026
* Refactored logging and error handling to improve system observability.
* Enhanced modularity by abstracting core logic into new functions and updating existing ones for better separation of concerns.

## Mar 27, 2026
* Refined the core logic and architecture, drawing inspiration from AI-assisted refactoring patterns.
* Successfully implemented Docker Volume Mounts using the Go client library to manage persistent data across container lifecycles.

## Mar 28, 2026
* Set up the official Git repository for the project.
* Installing and configuring the GitHub CLI (gh) inside the Docker container.
* Initiated development of the AI-driven Pull Request orchestration workflow.

## Mar 29, 2026
* Perfecting the logic that allows the AI to open a Pull Request once the code is ready.

## Mar 30, 2026
* Finally managed to push the code to GitHub and open a PR.

## Mar 31, 2026
* Began implementing dynamic Docker builds.

## Apr 01, 2026
* Still working on dynamic Docker builds.

## Apr 02, 2026
* Clone the repository to mount inside of the docker container instead of clone that on during build time.

## Apr 03, 2026
* Fix bugs of the repo cloning process.
* Did a bit of refactoring and removed commented codes.

## Apr 04, 2026
* Fix "Error response from daemon: crun: cannot stat `/Users/sithumsandeepa/pokunuvita/sithumonline/movie-box/opencode_data`: No such file or directory: OCI runtime attempted to invoke a command that was not found".
* Start working dynamic Docker build with the help of `entrypoint.sh` and it worked with the docker compose.

## Apr 05, 2026
* Get the Kimi 2 API key and try k2.5
* Start to build the API server for handle multiple repose at once.
* Start to design database for API server.

## Apr 09, 2026
* Tried to make Kimi 2 API works but couldn't make it.

## Apr 10, 2026
* Make it worked with manually setting credentials.

## Apr 11, 2026
* Managed to set auth credentials via API (client.Execute)
