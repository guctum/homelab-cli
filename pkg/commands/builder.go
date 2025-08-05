package commands

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var buildCmd = &cobra.Command{
	Use:   "builder",
	Short: "Build tag and ship images",
	Long:  `Build and tag multiarch images with Homelab naming conventions`,
	RunE: func(cmd *cobra.Command, args []string) error {
		registry, _ := cmd.Flags().GetString("registry")
		project, _ := cmd.Flags().GetString("project")
		image, _ := cmd.Flags().GetString("image")
		context, _ := cmd.Flags().GetString("context")

		return BuildContainerImage(registry, project, image, context)
	},
}

func init() {
	buildCmd.Flags().String("registry", "", "Registry URL (required)")
	buildCmd.Flags().String("project", "", "Project name (required)")
	buildCmd.Flags().String("image", "", "Image name (required)")
	buildCmd.Flags().String("context", ".", "Build context (defaults to current directory)")

	buildCmd.MarkFlagRequired("registry")
	buildCmd.MarkFlagRequired("project")
	buildCmd.MarkFlagRequired("image")
}

func ExecuteCommand(command string, args ...string) (string, error) {
	cmd := exec.Command(command, args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return "", fmt.Errorf("%w: %s", err, stderr.String())
	}

	return stdout.String(), nil
}

func getCommitID() (string, error) {
	output, err := ExecuteCommand("git", "log", "-n", "1", "--format=%h")
	if err != nil {
		return "", fmt.Errorf("failed to get git commit: %w", err)
	}
	return strings.TrimSpace(output), nil
}

func getBranch() (string, error) {
	output, err := ExecuteCommand("git", "rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", fmt.Errorf("failed to get git branch: %w", err)
	}
	return strings.TrimSpace(output), nil
}

// Function that executes dockerx build
func BuildContainerImage(registryURL, project, imageName string, context string) error {
	nyc, err := time.LoadLocation("America/New_York")
	if err != nil {
		return fmt.Errorf("failed to load timezone: %w", err)
	}
	date := time.Now().In(nyc).Format("200601021504")

	commit, err := getCommitID()
	if err != nil {
		return err
	}

	branch, err := getBranch()
	if err != nil {
		return err
	}

	stage := "development"
	if branch == "main" {
		stage = "release"
	}

	fmt.Printf("Building image from branch: %s\n", branch)
	fmt.Printf("Building image with commit: %s\n", commit)
	fmt.Printf("Building with stage: %s\n", stage)
	fmt.Printf("Image tag generated: %s-%s-%s\n", date, commit, stage)

	tag := fmt.Sprintf("%s/%s/%s:%s-%s-%s", registryURL, project, imageName, date, commit, stage)

	_, err = ExecuteCommand("docker", "buildx", "build",
		"--platform", "linux/amd64,linux/arm64",
		"-t", tag,
		"--push",
		context)

	if err != nil {
		return fmt.Errorf("docker buildx build failed: %w", err)
	}

	return nil
}
