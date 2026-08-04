package constants

import "strings"

const (
	ApplicationRuntimeNodeJS = "NodeJS"
	ApplicationRuntimeGo     = "Go"
	ApplicationRuntimeJava   = "Java"
	ApplicationRuntimePython = "Python"
	ApplicationRuntimeDotNet = "DotNet"
	ApplicationRuntimeStatic = "Static"
	ApplicationRuntimeDocker = "Docker"

	ApplicationStatusDraft    = "Draft"
	ApplicationStatusReady    = "Ready"
	ApplicationStatusArchived = "Archived"

	ApplicationDefaultBranch = "main"
)

var ValidApplicationRuntimes = []string{
	ApplicationRuntimeNodeJS,
	ApplicationRuntimeGo,
	ApplicationRuntimeJava,
	ApplicationRuntimePython,
	ApplicationRuntimeDotNet,
	ApplicationRuntimeStatic,
	ApplicationRuntimeDocker,
}

var ValidApplicationStatuses = []string{
	ApplicationStatusDraft,
	ApplicationStatusReady,
	ApplicationStatusArchived,
}

func NormalizeApplicationRuntime(value string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "nodejs":
		return ApplicationRuntimeNodeJS, true
	case "go":
		return ApplicationRuntimeGo, true
	case "java":
		return ApplicationRuntimeJava, true
	case "python":
		return ApplicationRuntimePython, true
	case "dotnet", "dot-net", ".net":
		return ApplicationRuntimeDotNet, true
	case "static":
		return ApplicationRuntimeStatic, true
	case "docker":
		return ApplicationRuntimeDocker, true
	default:
		return "", false
	}
}

func NormalizeApplicationStatus(value string) (string, bool) {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "draft":
		return ApplicationStatusDraft, true
	case "ready":
		return ApplicationStatusReady, true
	case "archived":
		return ApplicationStatusArchived, true
	default:
		return "", false
	}
}

func IsValidApplicationRuntime(value string) bool {
	_, ok := NormalizeApplicationRuntime(value)
	return ok
}

func IsValidApplicationStatus(value string) bool {
	_, ok := NormalizeApplicationStatus(value)
	return ok
}
