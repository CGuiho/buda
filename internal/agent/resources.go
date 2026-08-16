package agent

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"github.com/CGuiho/buda/prompts"
	"github.com/CGuiho/buda/skills"
	"go.yaml.in/yaml/v3"
)

const (
	SkillID          = "guiho-s-0002-buda"
	InstructionID    = "guiho-i-buda"
	InstructionBegin = "<!-- BEGIN BUDA INSTRUCTIONS -->"
	InstructionEnd   = "<!-- END BUDA INSTRUCTIONS -->"
	PromptID         = "guiho-p-buda"
)

type Error struct {
	Code    int
	Message string
	Cause   error
}

func (e *Error) Error() string {
	if e.Cause == nil {
		return e.Message
	}
	return e.Message + ": " + e.Cause.Error()
}

func (e *Error) Unwrap() error { return e.Cause }
func (e *Error) ExitCode() int { return e.Code }

func usage(message string) error { return &Error{Code: 2, Message: message} }
func repositoryError(message string, err error) error {
	return &Error{Code: 3, Message: message, Cause: err}
}
func mutation(message string, err error) error {
	return &Error{Code: 5, Message: message, Cause: err}
}

type Resources struct {
	Skill  fs.FS
	Prompt fs.FS
}

func DefaultResources() Resources {
	skillFS, err := fs.Sub(skills.FS, SkillID)
	if err != nil {
		panic(fmt.Sprintf("load embedded Buda skill: %v", err))
	}
	return Resources{Skill: skillFS, Prompt: prompts.FS}
}

type Service struct {
	resources Resources
	homeDir   func() (string, error)
}

type Option func(*Service)

func WithHomeDir(resolver func() (string, error)) Option {
	return func(service *Service) { service.homeDir = resolver }
}

func NewService(resources Resources, options ...Option) *Service {
	service := &Service{resources: resources}
	for _, option := range options {
		option(service)
	}
	return service
}

type SkillRecord struct {
	ID          string `json:"id"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Digest      string `json:"digest"`
}

type Prompt struct {
	ID          string `json:"id"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Body        string `json:"body,omitempty"`
}

func (s *Service) Skill() (SkillRecord, error) {
	content, err := fs.ReadFile(s.resources.Skill, "SKILL.md")
	if err != nil {
		return SkillRecord{}, mutation("read embedded Buda skill", err)
	}
	frontmatter, _, ok := splitFrontmatter(string(content))
	if !ok {
		return SkillRecord{}, mutation("embedded Buda skill is missing YAML frontmatter", nil)
	}
	var metadata struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
		Version     string `yaml:"version"`
	}
	if err := yaml.Unmarshal([]byte(frontmatter), &metadata); err != nil {
		return SkillRecord{}, mutation("decode embedded Buda skill metadata", err)
	}
	if metadata.Name != SkillID || metadata.Description == "" || metadata.Version == "" {
		return SkillRecord{}, mutation("embedded Buda skill metadata is invalid", nil)
	}
	digest, err := treeDigest(s.resources.Skill)
	if err != nil {
		return SkillRecord{}, mutation("digest embedded Buda skill", err)
	}
	return SkillRecord{ID: SkillID, Version: metadata.Version, Description: metadata.Description, Digest: digest}, nil
}

func (s *Service) Prompt() (Prompt, error) {
	content, err := fs.ReadFile(s.resources.Prompt, "guiho-p-buda.md")
	if err != nil {
		return Prompt{}, mutation("read embedded Buda setup prompt", err)
	}
	frontmatter, body, ok := splitFrontmatter(string(content))
	if !ok {
		return Prompt{}, mutation("embedded Buda setup prompt is missing YAML frontmatter", nil)
	}
	var metadata struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
		Version     string `yaml:"version"`
	}
	if err := yaml.Unmarshal([]byte(frontmatter), &metadata); err != nil {
		return Prompt{}, mutation("decode embedded Buda setup prompt metadata", err)
	}
	if metadata.Name != PromptID || metadata.Description == "" || metadata.Version == "" {
		return Prompt{}, mutation("embedded Buda setup prompt metadata is invalid", nil)
	}
	return Prompt{ID: PromptID, Version: metadata.Version, Description: metadata.Description, Body: strings.TrimSpace(body)}, nil
}

// Instruction returns the separately typed managed project instruction. It is
// deliberately not exposed as an agent prompt: the convention requires the
// instruction and setup prompt to have distinct resource identities.
func (s *Service) Instruction() (Prompt, error) {
	content, err := fs.ReadFile(s.resources.Prompt, "guiho-i-buda.md")
	if err != nil {
		return Prompt{}, mutation("read embedded Buda instruction", err)
	}
	frontmatter, body, ok := splitFrontmatter(string(content))
	if !ok {
		return Prompt{}, mutation("embedded Buda instruction is missing YAML frontmatter", nil)
	}
	var metadata struct {
		Name        string `yaml:"name"`
		Description string `yaml:"description"`
		Version     string `yaml:"version"`
	}
	if err := yaml.Unmarshal([]byte(frontmatter), &metadata); err != nil {
		return Prompt{}, mutation("decode embedded Buda instruction metadata", err)
	}
	if metadata.Name != InstructionID || metadata.Description == "" || metadata.Version == "" {
		return Prompt{}, mutation("embedded Buda instruction metadata is invalid", nil)
	}
	return Prompt{ID: InstructionID, Version: metadata.Version, Description: metadata.Description, Body: strings.TrimSpace(body)}, nil
}

func treeDigest(source fs.FS) (string, error) {
	var paths []string
	if err := fs.WalkDir(source, ".", func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if !entry.IsDir() {
			paths = append(paths, path)
		}
		return nil
	}); err != nil {
		return "", err
	}
	sort.Strings(paths)
	hash := sha256.New()
	for _, path := range paths {
		content, err := fs.ReadFile(source, path)
		if err != nil {
			return "", err
		}
		_, _ = hash.Write([]byte(path))
		_, _ = hash.Write([]byte{0})
		_, _ = hash.Write(content)
		_, _ = hash.Write([]byte{0})
	}
	return "sha256:" + hex.EncodeToString(hash.Sum(nil)), nil
}

func splitFrontmatter(content string) (string, string, bool) {
	trimmed := strings.TrimLeft(content, "\ufeff \t\r\n")
	if !strings.HasPrefix(trimmed, "---") {
		return "", content, false
	}
	rest := strings.TrimPrefix(trimmed, "---")
	rest = strings.TrimPrefix(rest, "\r\n")
	rest = strings.TrimPrefix(rest, "\n")
	index := strings.Index(rest, "\n---")
	if index < 0 {
		return "", content, false
	}
	body := rest[index+4:]
	body = strings.TrimPrefix(body, "\r\n")
	body = strings.TrimPrefix(body, "\n")
	return strings.TrimSpace(rest[:index]), body, true
}
