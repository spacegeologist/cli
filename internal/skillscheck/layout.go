// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package skillscheck

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const (
	LayoutSeparate = "separate"
	LayoutSuite    = "suite"
	LayoutHybrid   = "hybrid"

	suiteSkillName         = "lark-suite"
	sharedSkillName        = "lark-shared"
	suiteRoutesPlaceholder = "<!-- LARK_SUITE_ROUTES -->"
)

type GlobalSkillInfo struct {
	Name string
	Path string
}

func NormalizeLayout(layout string) (string, bool) {
	switch strings.TrimSpace(layout) {
	case "", LayoutSeparate:
		return LayoutSeparate, true
	case LayoutSuite:
		return LayoutSuite, true
	case LayoutHybrid:
		return LayoutHybrid, true
	default:
		return "", false
	}
}

func ParseCollectedSkills(value string) []string {
	seen := map[string]bool{}
	for _, part := range strings.Split(value, ",") {
		name := strings.TrimSpace(part)
		if name != "" {
			seen[name] = true
		}
	}
	return sortedKeys(seen)
}

func ParseGlobalSkillInfosJSON(text string) []GlobalSkillInfo {
	infos, _ := parseGlobalSkillInfosJSON(text)
	return infos
}

func parseGlobalSkillInfosJSON(text string) ([]GlobalSkillInfo, bool) {
	type globalSkill struct {
		Name string `json:"name"`
		Path string `json:"path"`
	}

	var skills []globalSkill
	if err := json.Unmarshal([]byte(text), &skills); err != nil {
		return nil, false
	}

	seen := map[string]GlobalSkillInfo{}
	for _, skill := range skills {
		name := strings.TrimSpace(skill.Name)
		path := strings.TrimSpace(skill.Path)
		if name == "" || path == "" || !skillNamePattern.MatchString(name) {
			continue
		}
		seen[name] = GlobalSkillInfo{Name: name, Path: path}
	}

	out := make([]GlobalSkillInfo, 0, len(seen))
	for _, info := range seen {
		out = append(out, info)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out, true
}

func installedSkillNamesFromInfos(infos []GlobalSkillInfo) []string {
	seen := map[string]bool{}
	for _, info := range infos {
		seen[info.Name] = true
		if info.Name == suiteSkillName {
			for _, subskill := range listSuiteSubskills(info.Path) {
				seen[subskill] = true
			}
		}
	}
	return sortedKeys(seen)
}

func listSuiteSubskills(suitePath string) []string {
	entries, err := os.ReadDir(filepath.Join(suitePath, "references", "subskills"))
	if err != nil {
		return nil
	}
	seen := map[string]bool{}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := strings.TrimSpace(entry.Name())
		if name != "" && skillNamePattern.MatchString(name) {
			seen[name] = true
		}
	}
	return sortedKeys(seen)
}

func normalOfficialSkills(skills []string) []string {
	out := []string{}
	for _, skill := range uniqueSorted(skills) {
		if skill == suiteSkillName {
			continue
		}
		out = append(out, skill)
	}
	return out
}

func resolveCollectedSkills(layout string, requested, official []string, previous *SkillsState, stateReadable bool, skippedDeleted []string) ([]string, error) {
	officialSet := toSet(official)
	deletedSet := toSet(skippedDeleted)
	switch layout {
	case LayoutSeparate:
		return []string{}, nil
	case LayoutSuite:
		return suiteEffectiveSkills(official, deletedSet), nil
	case LayoutHybrid:
		collected := []string{}
		for _, skill := range uniqueSorted(requested) {
			if skill == sharedSkillName {
				return nil, fmt.Errorf("%s is not selectable in hybrid layout", sharedSkillName)
			}
			if !officialSet[skill] {
				return nil, fmt.Errorf("collected skill %q is not in official skills", skill)
			}
			if !deletedSet[skill] {
				collected = append(collected, skill)
			}
		}
		for _, skill := range newlyOfficialSkills(official, previous, stateReadable) {
			if skill != sharedSkillName && !deletedSet[skill] {
				collected = append(collected, skill)
			}
		}
		if officialSet[sharedSkillName] {
			collected = append([]string{sharedSkillName}, collected...)
		}
		return uniqueSortedWithFirst(collected, sharedSkillName), nil
	default:
		return nil, fmt.Errorf("unsupported skills layout %q", layout)
	}
}

func suiteEffectiveSkills(official []string, deletedSet map[string]bool) []string {
	out := []string{}
	for _, skill := range normalOfficialSkills(official) {
		if skill == sharedSkillName || !deletedSet[skill] {
			out = append(out, skill)
		}
	}
	return out
}

func newlyOfficialSkills(official []string, previous *SkillsState, stateReadable bool) []string {
	if !stateReadable || previous == nil {
		return []string{}
	}
	previousSet := toSet(previous.OfficialSkills)
	out := []string{}
	for _, skill := range normalOfficialSkills(official) {
		if !previousSet[skill] {
			out = append(out, skill)
		}
	}
	return out
}

func uniqueSortedWithFirst(values []string, first string) []string {
	seen := toSet(values)
	if !seen[first] {
		return sortedKeys(seen)
	}
	delete(seen, first)
	return append([]string{first}, sortedKeys(seen)...)
}

func assembleSuiteLayout(layout string, collected []string, infos []GlobalSkillInfo) error {
	if layout == LayoutSeparate {
		return nil
	}

	infoByName := map[string]GlobalSkillInfo{}
	for _, info := range infos {
		infoByName[info.Name] = info
	}
	suiteInfo, ok := infoByName[suiteSkillName]
	if !ok {
		return fmt.Errorf("%s was not installed from isolated skills source", suiteSkillName)
	}

	subskillsDir := filepath.Join(suiteInfo.Path, "references", "subskills")
	if err := os.RemoveAll(subskillsDir); err != nil {
		return err
	}
	if err := os.MkdirAll(subskillsDir, 0o755); err != nil {
		return err
	}

	for _, skill := range collected {
		info, ok := infoByName[skill]
		if !ok {
			return fmt.Errorf("collected skill %q was not installed", skill)
		}
		dst := filepath.Join(subskillsDir, skill)
		if layout == LayoutHybrid && skill == sharedSkillName {
			if err := copyDir(info.Path, dst); err != nil {
				return err
			}
			continue
		}
		if err := moveDir(info.Path, dst); err != nil {
			return err
		}
	}

	return rewriteSuiteRoutes(suiteInfo.Path, collected)
}

func moveDir(src, dst string) error {
	if samePath(src, dst) {
		return nil
	}
	if err := os.RemoveAll(dst); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	if err := copyDir(src, dst); err != nil {
		return err
	}
	return os.RemoveAll(src)
}

func copyDir(src, dst string) error {
	if samePath(src, dst) {
		return nil
	}
	if err := os.RemoveAll(dst); err != nil {
		return err
	}
	return filepath.WalkDir(src, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		info, err := d.Info()
		if err != nil {
			return err
		}
		if d.IsDir() {
			return os.MkdirAll(target, info.Mode())
		}
		if info.Mode()&os.ModeSymlink != 0 {
			link, err := os.Readlink(path)
			if err != nil {
				return err
			}
			return os.Symlink(link, target)
		}
		return copyFile(path, target, info.Mode())
	})
}

func copyFile(src, dst string, mode os.FileMode) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, mode)
	if err != nil {
		return err
	}
	defer out.Close()

	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}

func samePath(a, b string) bool {
	aa, errA := filepath.Abs(a)
	bb, errB := filepath.Abs(b)
	return errA == nil && errB == nil && aa == bb
}

func rewriteSuiteRoutes(suitePath string, collected []string) error {
	skillPath := filepath.Join(suitePath, "SKILL.md")
	data, err := os.ReadFile(skillPath)
	if err != nil {
		return err
	}
	text := string(data)
	if !strings.Contains(text, suiteRoutesPlaceholder) {
		return fmt.Errorf("%s does not contain route placeholder", skillPath)
	}

	routes := []string{}
	for _, skill := range collected {
		description, err := readSkillDescription(filepath.Join(suitePath, "references", "subskills", skill, "SKILL.md"))
		if err != nil {
			return err
		}
		routes = append(routes, fmt.Sprintf("- %s: %s", skill, oneLine(description)))
	}
	text = strings.Replace(text, suiteRoutesPlaceholder, strings.Join(routes, "\n"), 1)
	return os.WriteFile(skillPath, []byte(text), 0o644)
}

func readSkillDescription(skillPath string) (string, error) {
	data, err := os.ReadFile(skillPath)
	if err != nil {
		return "", err
	}
	text := string(data)
	if !strings.HasPrefix(text, "---") {
		return "", fmt.Errorf("missing frontmatter in %s", skillPath)
	}
	parts := strings.SplitN(text, "---", 3)
	if len(parts) < 3 {
		return "", fmt.Errorf("missing frontmatter in %s", skillPath)
	}
	lines := strings.Split(parts[1], "\n")
	for i, line := range lines {
		if !strings.HasPrefix(line, "description:") {
			continue
		}
		value := strings.TrimSpace(strings.TrimPrefix(line, "description:"))
		if value == "|" || value == ">" {
			block := []string{}
			for _, blockLine := range lines[i+1:] {
				if strings.TrimSpace(blockLine) == "" {
					continue
				}
				if !strings.HasPrefix(blockLine, " ") {
					break
				}
				block = append(block, strings.TrimSpace(blockLine))
			}
			return strings.Join(block, " "), nil
		}
		return strings.Trim(value, `"'`), nil
	}
	return "", errors.New("missing frontmatter description")
}

func oneLine(value string) string {
	return strings.Join(strings.Fields(value), " ")
}
