package slackbot

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

// ParseRepositoryArgument parses and validates a repository path relative to GOPATH/src.
func ParseRepositoryArgument(raw string) (string, error) {
	repository := strings.Trim(strings.TrimSpace(raw), "`<>")
	repository = strings.ReplaceAll(repository, `\`, "/")
	cleaned := path.Clean(repository)
	if repository == "" || cleaned == "." || cleaned == ".." ||
		strings.HasPrefix(cleaned, "../") || path.IsAbs(cleaned) ||
		strings.Contains(cleaned, "$") || strings.ContainsRune(cleaned, '\x00') ||
		(len(cleaned) >= 2 && cleaned[1] == ':') {
		return "", fmt.Errorf("invalid repository path %q", raw)
	}
	return cleaned, nil
}

// ResolveRepositoryPath confines a repository argument to root, including
// after resolving symlinks.
func ResolveRepositoryPath(root, raw string) (string, error) {
	repository, err := ParseRepositoryArgument(raw)
	if err != nil {
		return "", err
	}
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		return "", fmt.Errorf("resolve repository root: %w", err)
	}
	resolvedRoot, err = filepath.Abs(resolvedRoot)
	if err != nil {
		return "", fmt.Errorf("resolve repository root: %w", err)
	}
	resolvedRepository, err := filepath.EvalSymlinks(filepath.Join(resolvedRoot, filepath.FromSlash(repository)))
	if err != nil {
		return "", fmt.Errorf("resolve repository path: %w", err)
	}
	resolvedRepository, err = filepath.Abs(resolvedRepository)
	if err != nil {
		return "", fmt.Errorf("resolve repository path: %w", err)
	}
	relative, err := filepath.Rel(resolvedRoot, resolvedRepository)
	if err != nil || relative == ".." || strings.HasPrefix(relative, ".."+string(filepath.Separator)) {
		return "", fmt.Errorf("repository path %q escapes GOPATH/src", raw)
	}
	return resolvedRepository, nil
}
