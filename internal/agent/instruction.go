package agent

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/CGuiho/buda/internal/config"
)

type InstructionContext struct {
	Wiki   string `json:"wiki"`
	WikiID string `json:"wiki_id"`
	Bundle string `json:"bundle"`
}

type InstructionTarget struct {
	Path      string `json:"path"`
	Exists    bool   `json:"exists"`
	Managed   bool   `json:"managed"`
	Current   bool   `json:"current"`
	Changed   bool   `json:"changed,omitempty"`
	Malformed bool   `json:"malformed,omitempty"`
}

func (s *Service) InstructionTemplate(context InstructionContext) (string, error) {
	prompt, err := s.Prompt()
	if err != nil {
		return "", err
	}
	body := strings.ReplaceAll(prompt.Body, "{{WIKI_ID}}", context.WikiID)
	body = strings.ReplaceAll(body, "{{BUNDLE}}", filepath.ToSlash(context.Bundle))
	return strings.TrimRight(body, "\r\n"), nil
}

func (s *Service) ReadInstructionContext(wiki string) (InstructionContext, error) {
	if strings.TrimSpace(wiki) == "" {
		return InstructionContext{}, repositoryError("--wiki is required; Buda never selects a wiki implicitly", nil)
	}
	absolute, err := filepath.Abs(wiki)
	if err != nil {
		return InstructionContext{}, repositoryError("resolve --wiki", err)
	}
	absolute = filepath.Clean(absolute)
	value, err := config.Load(filepath.Join(absolute, config.FileName))
	if err != nil {
		return InstructionContext{}, repositoryError("load strict buda.yaml", err)
	}
	return InstructionContext{Wiki: absolute, WikiID: value.WikiID, Bundle: filepath.Clean(value.Bundle)}, nil
}

func (s *Service) ListInstructions(wiki string) ([]InstructionTarget, error) {
	context, err := s.ReadInstructionContext(wiki)
	if err != nil {
		return nil, err
	}
	template, err := s.InstructionTemplate(context)
	if err != nil {
		return nil, err
	}
	targets, err := instructionTargets(context.Wiki, true)
	if err != nil {
		return nil, err
	}
	results := make([]InstructionTarget, 0, len(targets))
	for _, path := range targets {
		content, readErr := os.ReadFile(path)
		exists := readErr == nil
		if readErr != nil && !os.IsNotExist(readErr) {
			return nil, mutation("read instruction target", readErr)
		}
		blocks, parseErr := parseInstructionBlocks(string(content))
		result := InstructionTarget{Path: path, Exists: exists, Malformed: parseErr != nil}
		if parseErr == nil && len(blocks) > 0 {
			result.Managed = true
			result.Current = len(blocks) == 1 && normalizeNewlines(blocks[0].body) == normalizeNewlines(template)
		}
		results = append(results, result)
	}
	return results, nil
}

func (s *Service) ApplyInstructions(wiki string) ([]InstructionTarget, error) {
	return s.mutateInstructions(wiki, false)
}

func (s *Service) RemoveInstructions(wiki string) ([]InstructionTarget, error) {
	return s.mutateInstructions(wiki, true)
}

func (s *Service) mutateInstructions(wiki string, remove bool) ([]InstructionTarget, error) {
	context, err := s.ReadInstructionContext(wiki)
	if err != nil {
		return nil, err
	}
	template, err := s.InstructionTemplate(context)
	if err != nil {
		return nil, err
	}
	targets, err := instructionTargets(context.Wiki, !remove)
	if err != nil {
		return nil, err
	}
	stages := make([]fileStage, 0, len(targets))
	results := make([]InstructionTarget, 0, len(targets))
	for _, path := range targets {
		content, readErr := os.ReadFile(path)
		exists := readErr == nil
		if readErr != nil && !os.IsNotExist(readErr) {
			cleanupFileStages(stages)
			return nil, mutation("read instruction target", readErr)
		}
		next, reconcileErr := reconcileInstructions(string(content), template, remove)
		if reconcileErr != nil {
			cleanupFileStages(stages)
			return nil, mutation("refuse malformed Buda instruction markers in "+path, reconcileErr)
		}
		changed := string(content) != next
		result := InstructionTarget{Path: path, Exists: exists, Managed: !remove, Current: !remove, Changed: changed}
		results = append(results, result)
		if !changed {
			continue
		}
		mode := fs.FileMode(0o644)
		if info, statErr := os.Stat(path); statErr == nil {
			mode = info.Mode().Perm()
		}
		stage, stageErr := prepareFileStage(path, []byte(next), mode)
		if stageErr != nil {
			cleanupFileStages(stages)
			return nil, stageErr
		}
		stages = append(stages, stage)
	}
	if err := commitFileStages(stages); err != nil {
		return nil, err
	}
	return results, nil
}

func (s *Service) InstalledInstruction(wiki, name string) (string, error) {
	context, err := s.ReadInstructionContext(wiki)
	if err != nil {
		return "", err
	}
	switch strings.ToUpper(strings.TrimSpace(name)) {
	case "AGENTS.MD":
		name = "AGENTS.md"
	case "CLAUDE.MD":
		name = "CLAUDE.md"
	default:
		return "", usage("instruction target must be AGENTS.md or CLAUDE.md")
	}
	path := filepath.Join(context.Wiki, name)
	content, err := os.ReadFile(path)
	if err != nil {
		return "", mutation("read installed instruction target", err)
	}
	blocks, err := parseInstructionBlocks(string(content))
	if err != nil {
		return "", mutation("parse installed Buda instruction block", err)
	}
	if len(blocks) != 1 {
		return "", mutation("installed instruction target does not contain exactly one Buda block", nil)
	}
	return InstructionBegin + newlineStyle(string(content)) + blocks[0].body + newlineStyle(string(content)) + InstructionEnd, nil
}

func instructionTargets(wiki string, createDefault bool) ([]string, error) {
	candidates := []string{filepath.Join(wiki, "AGENTS.md"), filepath.Join(wiki, "CLAUDE.md")}
	var targets []string
	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			targets = append(targets, path)
		} else if !os.IsNotExist(err) {
			return nil, mutation("inspect instruction target", err)
		}
	}
	if len(targets) == 0 && createDefault {
		targets = append(targets, candidates[0])
	}
	return targets, nil
}

type instructionBlock struct {
	start int
	end   int
	body  string
}

func parseInstructionBlocks(content string) ([]instructionBlock, error) {
	upper := strings.ToUpper(content)
	if strings.Count(upper, "<!-- BEGIN BUDA") != strings.Count(content, InstructionBegin) ||
		strings.Count(upper, "<!-- END BUDA") != strings.Count(content, InstructionEnd) {
		return nil, fmt.Errorf("unrecognized Buda marker text")
	}
	var blocks []instructionBlock
	cursor := 0
	for {
		beginRelative := strings.Index(content[cursor:], InstructionBegin)
		endRelative := strings.Index(content[cursor:], InstructionEnd)
		if beginRelative < 0 && endRelative < 0 {
			break
		}
		if beginRelative < 0 || (endRelative >= 0 && endRelative < beginRelative) {
			return nil, fmt.Errorf("end marker appears without a preceding begin marker")
		}
		begin := cursor + beginRelative
		bodyStart := begin + len(InstructionBegin)
		endInBody := strings.Index(content[bodyStart:], InstructionEnd)
		if endInBody < 0 {
			return nil, fmt.Errorf("begin marker is not closed")
		}
		endStart := bodyStart + endInBody
		if strings.Contains(content[bodyStart:endStart], InstructionBegin) {
			return nil, fmt.Errorf("nested begin marker")
		}
		body := strings.Trim(content[bodyStart:endStart], "\r\n")
		blocks = append(blocks, instructionBlock{start: begin, end: endStart + len(InstructionEnd), body: body})
		cursor = endStart + len(InstructionEnd)
	}
	return blocks, nil
}

func reconcileInstructions(content, template string, remove bool) (string, error) {
	blocks, err := parseInstructionBlocks(content)
	if err != nil {
		return "", err
	}
	newline := newlineStyle(content)
	canonicalBody := strings.ReplaceAll(normalizeNewlines(strings.Trim(template, "\r\n")), "\n", newline)
	canonical := InstructionBegin + newline + canonicalBody + newline + InstructionEnd
	if len(blocks) == 0 {
		if remove {
			return content, nil
		}
		result := content
		if result != "" && !strings.HasSuffix(result, "\n") {
			result += newline
		}
		if result != "" && !strings.HasSuffix(result, newline+newline) {
			result += newline
		}
		return result + canonical + newline, nil
	}
	var output strings.Builder
	cursor := 0
	inserted := false
	for _, block := range blocks {
		output.WriteString(content[cursor:block.start])
		if !remove && !inserted {
			output.WriteString(canonical)
			inserted = true
		}
		cursor = block.end
	}
	output.WriteString(content[cursor:])
	return output.String(), nil
}

func newlineStyle(content string) string {
	if strings.Contains(content, "\r\n") {
		return "\r\n"
	}
	return "\n"
}

func normalizeNewlines(value string) string {
	return strings.ReplaceAll(value, "\r\n", "\n")
}

type fileStage struct {
	destination string
	temporary   string
	backup      string
	existed     bool
	committed   bool
}

func prepareFileStage(destination string, content []byte, mode fs.FileMode) (fileStage, error) {
	stage := fileStage{destination: destination}
	parent := filepath.Dir(destination)
	if err := os.MkdirAll(parent, 0o755); err != nil {
		return stage, mutation("create instruction parent", err)
	}
	temporary, err := os.CreateTemp(parent, ".buda-instruction-*")
	if err != nil {
		return stage, mutation("create staged instruction file", err)
	}
	stage.temporary = temporary.Name()
	defer temporary.Close()
	if _, err := io.Copy(temporary, bytes.NewReader(content)); err != nil {
		_ = os.Remove(stage.temporary)
		return stage, mutation("write staged instruction file", err)
	}
	if err := temporary.Sync(); err != nil {
		_ = os.Remove(stage.temporary)
		return stage, mutation("sync staged instruction file", err)
	}
	if err := temporary.Chmod(mode); err != nil {
		_ = os.Remove(stage.temporary)
		return stage, mutation("set staged instruction mode", err)
	}
	if _, err := os.Stat(destination); err == nil {
		stage.existed = true
	} else if !os.IsNotExist(err) {
		_ = os.Remove(stage.temporary)
		return stage, mutation("inspect instruction target", err)
	}
	stage.backup, err = reserveSibling(parent, ".buda-instruction-backup-*")
	return stage, err
}

func commitFileStages(stages []fileStage) error {
	for index := range stages {
		stage := &stages[index]
		if stage.existed {
			if err := os.Rename(stage.destination, stage.backup); err != nil {
				_ = rollbackFileStages(stages)
				return mutation("stage existing instruction file", err)
			}
		}
		if err := os.Rename(stage.temporary, stage.destination); err != nil {
			_ = rollbackFileStages(stages)
			return mutation("atomically replace instruction file", err)
		}
		stage.temporary = ""
		stage.committed = true
	}
	cleanupFileStages(stages)
	return nil
}

func rollbackFileStages(stages []fileStage) error {
	var failures []string
	for index := len(stages) - 1; index >= 0; index-- {
		stage := &stages[index]
		if stage.committed {
			if err := os.Remove(stage.destination); err != nil && !os.IsNotExist(err) {
				failures = append(failures, err.Error())
			}
		}
		if stage.existed {
			if _, err := os.Stat(stage.backup); err == nil {
				if err := os.Rename(stage.backup, stage.destination); err != nil {
					failures = append(failures, err.Error())
				}
			}
		}
	}
	cleanupFileStages(stages)
	if len(failures) > 0 {
		return fmt.Errorf("%s", strings.Join(failures, "; "))
	}
	return nil
}

func cleanupFileStages(stages []fileStage) {
	for _, stage := range stages {
		if stage.temporary != "" {
			_ = os.Remove(stage.temporary)
		}
		if stage.backup != "" {
			_ = os.Remove(stage.backup)
		}
	}
}
