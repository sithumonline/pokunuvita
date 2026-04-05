package main

import (
	"archive/tar"
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"time"

	"github.com/docker/docker/api/types/build"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/sst/opencode-sdk-go"
	"github.com/sst/opencode-sdk-go/option"
)

const (
	imageTag            = "opencode-app:latest"
	containerName       = "opencode-instance"
	relativeMountPath   = "./opencode_data"
	containerMountPath  = "/root/.local/share/opencode"
	repositoryMountPath = "/app/repository"

	gh_username   = "sithumonline"
	gh_repository = "movie-box"

	hostIP        = "0.0.0.0"
	hostPort      = "3000"
	containerPort = "3000"

	dockerfilePath = "Dockerfile"
	sessionTitle   = "opencode development session"
	promptText     = `
You are working inside a Docker container on a Go repository. Complete the following upgrade workflow end-to-end:

## 1. Fix git push authentication
- Before doing anything else, reconfigure the git remote to use the GH_TOKEN env var:
  REPO_URL=$(git remote get-url origin)
  OWNER_REPO=$(echo $REPO_URL | sed 's/.*github.com[:/]\(.*\)\.git/\1/')
  git remote set-url origin https://${GH_TOKEN}@github.com/${OWNER_REPO}.git
- Verify the remote is set correctly: git remote get-url origin

## 2. Create a branch
- Create and switch to a new branch named: 'chore/dependency-upgrades'

## 3. Upgrade the Go version (if possible)
- Check the current Go version in 'go.mod'
- Find the latest stable Go version available in this environment ('go version' or check https://go.dev/dl/)
- If a newer stable version is available, update the 'go' directive in 'go.mod' accordingly
- Run 'go mod tidy' after any version change

## 4. Upgrade all dependencies
- Run 'go get -u ./...' to upgrade all direct and indirect dependencies to their latest minor/patch versions
- Run 'go mod tidy' to clean up 'go.mod' and 'go.sum'

## 5. Verify everything is okay
- Run 'go build ./...' — must pass with no errors
- Run 'go vet ./...' — must pass with no warnings
- Run 'go test ./...' — all tests must pass
- If any step fails, diagnose and fix the issue before proceeding. Do NOT continue to commit if checks fail.

## 6. Commit and push
- Stage all changes: 'git add go.mod go.sum'
- Write a clear, conventional commit message summarising what was upgraded (e.g. 'chore: upgrade go version and dependencies')
- Push the branch: 'git push -u origin chore/dependency-upgrades'

## 7. Create a pull request via the GitHub REST API
- Detect the default base branch: 'git remote show origin | grep 'HEAD branch' | awk '{print $NF}''
- Build the PR body by summarising the Go version change (if any) and key dependency bumps from 'git diff origin/<base>...HEAD -- go.mod'
- Create the PR using curl and the GH_TOKEN env var:

curl -s -X POST \
  -H "Authorization: Bearer ${GH_TOKEN}" \
  -H "Accept: application/vnd.github+json" \
  -H "X-GitHub-Api-Version: 2022-11-28" \
  https://api.github.com/repos/${OWNER_REPO}/pulls \
  -d '{
    "title": "chore: upgrade Go version and dependencies",
    "head": "chore/dependency-upgrades",
    "base": "<default-branch>",
    "body": "<generated-body>"
  }'

- Parse the response and report the PR URL ('html_url' field) on success
- If the API returns an error (non-201 status), print the full response and stop

Report back with the outcome of each step.
	`

	stopContainerOnExit = true
)

func imageExists(ctx context.Context, cli *client.Client, imageName string) (bool, error) {
	images, err := cli.ImageList(ctx, image.ListOptions{
		Filters: filters.NewArgs(filters.Arg("reference", imageName)),
	})
	if err != nil {
		return false, fmt.Errorf("failed to get image list: %w", err)
	}
	return len(images) > 0, nil
}

func buildImageFromDockerfileOnly(
	ctx context.Context,
	logger *slog.Logger,
	cli *client.Client,
	dockerfilePath string,
	imageTag string,
) error {
	pr, pw := io.Pipe()

	go func() {
		tw := tar.NewWriter(pw)
		defer tw.Close()
		defer pw.Close()

		content, err := os.ReadFile(dockerfilePath)
		if err != nil {
			_ = pw.CloseWithError(fmt.Errorf("read dockerfile %q: %w", dockerfilePath, err))
			return
		}

		hdr := &tar.Header{
			Name: "Dockerfile",
			Mode: 0600,
			Size: int64(len(content)),
		}

		if err := tw.WriteHeader(hdr); err != nil {
			_ = pw.CloseWithError(fmt.Errorf("write dockerfile header: %w", err))
			return
		}

		if _, err := tw.Write(content); err != nil {
			_ = pw.CloseWithError(fmt.Errorf("write dockerfile content: %w", err))
			return
		}
	}()

	res, err := cli.ImageBuild(ctx, pr, build.ImageBuildOptions{
		Tags:        []string{imageTag},
		Dockerfile:  "Dockerfile",
		Remove:      true,
		ForceRemove: true,
	})
	if err != nil {
		return fmt.Errorf("failed to build the image: %w", err)
	}
	defer res.Body.Close()

	scanner := bufio.NewScanner(res.Body)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 1024*1024)

	for scanner.Scan() {
		logger.Info("docker build", "output", scanner.Text())
	}

	return scanner.Err()
}

func ensureContainerRunning(
	ctx context.Context,
	logger *slog.Logger,
	cli *client.Client,
	name, imageTag, hostIP, hostPort, containerPort, mountSource, dataMountTarget, gh_uname, gh_repo string,
) (string, bool, error) {
	containers, err := cli.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filters.NewArgs(filters.Arg("name", name)),
	})
	if err != nil {
		return "", false, fmt.Errorf("failed to get containers list: %w", err)
	}

	dataDir := fmt.Sprintf("%s/opencode_data", mountSource)
	if err := os.MkdirAll(dataDir, 0755); err != nil {
		return "", false, fmt.Errorf("failed to create: %s: %w", dataDir, err)
	} else {
		s, _ := os.Stat(dataDir)
		logger.DebugContext(ctx, "this dir already exit", "dir", dataDir, "stat", s)
	}

	for _, ctr := range containers {
		for _, n := range ctr.Names {
			if n != fmt.Sprintf("/%s", containerName) {
				continue
			}

			if ctr.State == "running" {
				logger.Info("container already running", "container_id", ctr.ID, "name", name)
				return ctr.ID, false, nil
			}

			logger.Info("starting existing container", "container_id", ctr.ID, "state", ctr.State)
			if err := cli.ContainerStart(ctx, ctr.ID, container.StartOptions{}); err != nil {
				return "", false, err
			}
			return ctr.ID, true, nil
		}
	}

	portKey, err := nat.NewPort("tcp", containerPort)
	if err != nil {
		return "", false, fmt.Errorf("failed to bind the port to continer: %w", err)
	}

	envList := []string{fmt.Sprintf("GH_USERNAME=%s", gh_uname), fmt.Sprintf("GH_REPOSITORY=%s", gh_repo)}
	if env, ok := os.LookupEnv("OPENAI_API_KEY"); ok {
		envList = append(envList, fmt.Sprintf("OPENAI_API_KEY=%s", env))
		logger.DebugContext(ctx, "OPENAI_API_KEY found")
	} else {
		logger.DebugContext(ctx, "OPENAI_API_KEY not found")
	}
	if env, ok := os.LookupEnv("GH_TOKEN"); ok {
		envList = append(envList, fmt.Sprintf("GH_TOKEN=%s", env))
		logger.DebugContext(ctx, "GH_TOKEN found")
	} else {
		logger.DebugContext(ctx, "GH_TOKEN not found")
	}

	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image:        imageTag,
		ExposedPorts: nat.PortSet{portKey: struct{}{}},
		Env:          envList,
	}, &container.HostConfig{
		PortBindings: nat.PortMap{portKey: []nat.PortBinding{
			{
				HostIP:   hostIP,
				HostPort: hostPort,
			},
		}},
		Mounts: []mount.Mount{
			{
				Type:     mount.TypeBind,  // Use TypeVolume for named volumes
				Source:   dataDir,         //"/path/on/host", // Absolute path on host
				Target:   dataMountTarget, //"/path/in/container",
				ReadOnly: false,           // Optional: set to true for read-only
			},
		},
	}, nil, nil, name)
	if err != nil {
		return "", false, fmt.Errorf("failed to create the container named '%s': %w", name, err)
	}

	if err := cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", false, fmt.Errorf("failed to start the container name '%s': %w", name, err)
	}

	logger.Info("created and started new container", "container_id", resp.ID, "name", name)
	return resp.ID, true, nil
}

func ensureSession(
	ctx context.Context,
	logger *slog.Logger,
	client *opencode.Client,
	title string,
) (*opencode.Session, error) {
	sessions, err := client.Session.List(ctx, opencode.SessionListParams{})
	if err != nil {
		return nil, fmt.Errorf("failed to get the sessions list: %w", err)
	}
	if sessions == nil {
		return nil, errors.New("session list returned nil")
	}

	if len(*sessions) == 0 {
		logger.Info("creating new session", "title", title)
		return client.Session.New(ctx, opencode.SessionNewParams{
			Title: opencode.F(title),
		})
	}

	logger.Info("reusing existing session", "session_id", (*sessions)[0].ID, "count", len(*sessions))
	return &(*sessions)[0], nil
}

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug,
	}))

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	dockerClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		logger.Error("failed to create docker client", "err", err)
		os.Exit(1)
	}
	defer dockerClient.Close()

	logger.Info("starting",
		"image", imageTag,
		"container", containerName,
		"host_port", hostPort,
		"container_port", containerPort,
	)

	imageExists, err := imageExists(ctx, dockerClient, imageTag)
	if err != nil {
		logger.Error("application failed in check image", "err", err)
		os.Exit(1)
	}

	if !imageExists {
		logger.Info("image not found, building", "image", imageTag)

		buildCtx, cancel := context.WithTimeout(ctx, 15*time.Minute)
		defer cancel()

		if err := buildImageFromDockerfileOnly(buildCtx, logger, dockerClient, dockerfilePath, imageTag); err != nil {
			logger.Error("application failed in build image", "err", err)
			os.Exit(1)
		}
	}

	// absBasePath, err := filepath.Abs(relativeMountPath)
	// if err != nil {
	// 	logger.Error("application failed to get absolute host path", "err", err)
	// 	os.Exit(1)
	// }

	absBasePath := "/Users/sithumsandeepa/pokunuvita"
	absBasePath = fmt.Sprintf("%s/%s/%s", absBasePath, gh_username, gh_repository)

	containerID, _, err := ensureContainerRunning(
		ctx,
		logger,
		dockerClient,
		containerName,
		imageTag,
		hostIP,
		hostPort,
		containerPort,
		absBasePath, //absMountPath,
		containerMountPath,
		gh_username,
		gh_repository,
	)
	if err != nil {
		logger.Error("application failed in ensure container", "err", err)
		os.Exit(1)
	}

	// if startedNow {
	// 	logger.Info("waiting for container startup")
	// 	time.Sleep(5 * time.Second) // do we really need this?
	// }

	opencodeOptions := []option.RequestOption{
		option.WithBaseURL(fmt.Sprintf("http://localhost:%s", hostPort)),
	}

	opencodeClient := opencode.NewClient(opencodeOptions...)

	var opencodeSession *opencode.Session
	containerCheckingTime := time.Now()
	for {
		_opencodeSession, err := ensureSession(ctx, logger, opencodeClient, sessionTitle)
		if err != nil {
			logger.Error("application failed in ensure session", "err", err)
			// os.Exit(1)
			time.Sleep(time.Second)
		} else {
			opencodeSession = _opencodeSession
			logger.Info("application session is running")
			break
		}
	}

	logger.Info("using session", "session_id", opencodeSession.ID, "directory", opencodeSession.Directory, "startupTime", time.Since(containerCheckingTime))

	promptResp, err := opencodeClient.Session.Prompt(context.TODO(), opencodeSession.ID, opencode.SessionPromptParams{
		Parts: opencode.F([]opencode.SessionPromptParamsPartUnion{
			opencode.TextPartInputParam{
				Type: opencode.F(opencode.TextPartInputTypeText),
				Text: opencode.F(promptText),
			},
		}),
	})

	if err != nil {
		logger.Error("application failed in send prompt", "err", err)
		os.Exit(1)
	}

	for _, part := range promptResp.Parts {
		if textPart, ok := part.AsUnion().(opencode.TextPart); ok {
			logger.Info("response", "text", textPart.Text)
		}
	}

	if stopContainerOnExit {
		logger.Info("stopping container", "container_id", containerID)
		if err := dockerClient.ContainerStop(ctx, containerID, container.StopOptions{}); err != nil {
			logger.Error("application failed in stop container", "err", err)
			os.Exit(1)
		}
	}

	logger.Info("completed successfully")
}
