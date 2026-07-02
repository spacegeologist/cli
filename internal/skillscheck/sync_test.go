// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package skillscheck

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/larksuite/cli/internal/selfupdate"
)

func TestParseSkillsListIgnoresUnsupportedFormat(t *testing.T) {
	input := `Installed skills:
- lark-calendar
- lark-mail
lark-im
custom-skill
lark-base@1.0.0
lark-cli-harness:dev@0.1.0
`
	got := ParseSkillsList(input)
	if len(got) != 0 {
		t.Fatalf("ParseSkillsList() = %#v, want empty result for unsupported format", got)
	}
}

func TestParseOfficialSkillsListAcceptsNonLarkOfficialNames(t *testing.T) {
	input := `Available Skills
│    lark-calendar
│    official-shared
│    bad/name
`
	got := ParseSkillsList(input)
	want := []string{"lark-calendar", "official-shared"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseSkillsList() (Available Skills) = %#v, want %#v", got, want)
	}
}

func TestParseGlobalSkillsList(t *testing.T) {
	input := `Global Skills

lark-approval ~/.agents/skills/lark-approval
  Agents: TRAE CN, TRAE, TRAE-SOLO, TRAE CLI, TRAE CLI (Coco) +3 more
lark-attendance ~/.agents/skills/lark-attendance
  Agents: TRAE CN, TRAE, TRAE-SOLO, TRAE CLI, TRAE CLI (Coco) +3 more
lark-base ~/.agents/skills/lark-base
  Agents: TRAE CN, TRAE, TRAE-SOLO, TRAE CLI, TRAE CLI (Coco) +3 more
lark-calendar ~/.agents/skills/lark-calendar
  Agents: TRAE CN, TRAE, TRAE-SOLO, TRAE CLI, TRAE CLI (Coco) +3 more
dogfood ~/.hermes/skills/dogfood
  Agents: Hermes Agent
yuanbao ~/.hermes/skills/yuanbao
  Agents: Hermes Agent
`
	got := ParseSkillsList(input)
	want := []string{"dogfood", "lark-approval", "lark-attendance", "lark-base", "lark-calendar", "yuanbao"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseSkillsList() (Global Skills) = %#v, want %#v", got, want)
	}
}

func TestParseGlobalSkillsListWithANSI(t *testing.T) {
	input := "\x1b[1mGlobal Skills\x1b[0m\n\n" +
		"\x1b[36mlark-calendar\x1b[0m \x1b[38;5;102m~/.agents/skills/lark-calendar\x1b[0m\n" +
		"  \x1b[38;5;102mAgents:\x1b[0m TRAE CN, TRAE +3 more\n" +
		"\x1b[36mdogfood\x1b[0m \x1b[38;5;102m~/.hermes/skills/dogfood\x1b[0m\n" +
		"  \x1b[38;5;102mAgents:\x1b[0m Hermes Agent\n" +
		"\nTip: Use the -y flag to run in non-interactive mode (for CI and AI agents).\n"
	got := ParseSkillsList(input)
	want := []string{"dogfood", "lark-calendar"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseSkillsList() (ANSI Global Skills) = %#v, want %#v", got, want)
	}
}

func TestParseGlobalSkillsListWithIndentedGroupedRows(t *testing.T) {
	input := `Global Skills

General
  lark-apps ~/.agents/skills/lark-apps
  lark-base ~/.agents/skills/lark-base
`
	got := ParseSkillsList(input)
	want := []string{"lark-apps", "lark-base"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseSkillsList() (indented Global Skills) = %#v, want %#v", got, want)
	}
}

func TestParseGlobalSkillsJSON(t *testing.T) {
	input := `[
  {"name":"lark-calendar","path":"/Users/example/.agents/skills/lark-calendar","scope":"global","agents":["Codex"]},
  {"name":"lark-mail@1.2.3","path":"/Users/example/.agents/skills/lark-mail","scope":"global","agents":["Codex"]},
  {"name":"lark-calendar","path":"/Users/example/.agents/skills/lark-calendar","scope":"global","agents":["Codex"]},
  {"name":"  lark-base  ","path":"/Users/example/.agents/skills/lark-base","scope":"global","agents":["Codex"]},
  {"name":""},
  {"name":"   "},
  {"name":"bad skill"}
]`
	got := ParseGlobalSkillsJSON(input)
	want := []string{"lark-base", "lark-calendar", "lark-mail@1.2.3"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseGlobalSkillsJSON() = %#v, want %#v", got, want)
	}
}

func TestParseGlobalSkillsJSONInvalidOrUnsupported(t *testing.T) {
	for _, input := range []string{
		`not json`,
		`{"name":"lark-calendar"}`,
		`[]`,
	} {
		if got := ParseGlobalSkillsJSON(input); len(got) != 0 {
			t.Fatalf("ParseGlobalSkillsJSON(%q) = %#v, want empty", input, got)
		}
	}
}

func TestParseOfficialSkillsIndexJSON(t *testing.T) {
	input := `{
  "skills": [
    {"name":"lark-calendar","description":"Calendar","files":["SKILL.md"]},
    {"name":"lark-mail","description":"Mail","files":["SKILL.md","references/lark-mail-search.md"]},
    {"name":"  lark-base  ","description":"Base","files":[]},
    {"name":"lark-calendar","description":"duplicate","files":["SKILL.md"]},
    {"name":"custom-skill","description":"not official","files":["SKILL.md"]},
    {"name":"bad skill","description":"invalid","files":["SKILL.md"]},
    {"name":"","description":"empty","files":["SKILL.md"]}
  ]
}`
	got, err := ParseOfficialSkillsIndexJSON(input)
	if err != nil {
		t.Fatalf("ParseOfficialSkillsIndexJSON() err = %v, want nil", err)
	}
	want := []string{"custom-skill", "lark-base", "lark-calendar", "lark-mail"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ParseOfficialSkillsIndexJSON() = %#v, want %#v", got, want)
	}
}

func TestParseOfficialSkillsIndexJSONInvalidOrUnsupported(t *testing.T) {
	for _, input := range []string{
		`not json`,
		`[{"name":"lark-calendar"}]`,
		`{"name":"lark-calendar"}`,
		`{"skills":[]}`,
		`{"skills":[{"name":"bad skill"}]}`,
	} {
		got, err := ParseOfficialSkillsIndexJSON(input)
		if err == nil && len(got) != 0 {
			t.Fatalf("ParseOfficialSkillsIndexJSON(%q) = %#v, want empty", input, got)
		}
	}
}

func TestPlanNormal_WithReadableStatePreservesDeletedAndAddsNew(t *testing.T) {
	previous := &SkillsState{OfficialSkills: []string{"lark-calendar", "lark-mail"}}
	got := PlanSync(SyncInput{
		Version:        "1.0.33",
		OfficialSkills: []string{"lark-calendar", "lark-mail", "lark-new"},
		LocalSkills:    []string{"lark-calendar", "lark-custom"},
		PreviousState:  previous,
		StateReadable:  true,
		Force:          false,
	})

	assertStrings(t, got.ToUpdate, []string{"lark-calendar", "lark-new"})
	assertStrings(t, got.Added, []string{"lark-new"})
	assertStrings(t, got.SkippedDeleted, []string{"lark-mail"})
}

func TestPlanNormal_MissingStateInstallsAllOfficial(t *testing.T) {
	got := PlanSync(SyncInput{
		Version:        "1.0.33",
		OfficialSkills: []string{"lark-calendar", "lark-mail", "lark-new"},
		LocalSkills:    []string{"lark-calendar"},
		StateReadable:  false,
		Force:          false,
	})

	assertStrings(t, got.ToUpdate, []string{"lark-calendar", "lark-mail", "lark-new"})
	assertStrings(t, got.Added, []string{"lark-calendar", "lark-mail", "lark-new"})
	assertStrings(t, got.SkippedDeleted, []string{})
}

func TestPlanForceRestoresAllOfficial(t *testing.T) {
	got := PlanSync(SyncInput{
		Version:        "1.0.33",
		OfficialSkills: []string{"lark-calendar", "lark-mail", "lark-new"},
		LocalSkills:    []string{"lark-calendar"},
		PreviousState:  &SkillsState{OfficialSkills: []string{"lark-calendar", "lark-mail"}},
		StateReadable:  true,
		Force:          true,
	})

	assertStrings(t, got.ToUpdate, []string{"lark-calendar", "lark-mail", "lark-new"})
	assertStrings(t, got.Added, []string{})
	assertStrings(t, got.SkippedDeleted, []string{})
}

func TestPlanSuiteInstallsAllNormalOfficialSkills(t *testing.T) {
	got := PlanSync(SyncInput{
		Version:        "1.0.33",
		Layout:         LayoutSuite,
		OfficialSkills: []string{"lark-calendar", "lark-suite", "lark-mail"},
		LocalSkills:    []string{"lark-suite"},
	})

	assertStrings(t, got.OfficialSkills, []string{"lark-calendar", "lark-mail"})
	assertStrings(t, got.ToUpdate, []string{"lark-calendar", "lark-mail"})
	assertStrings(t, got.SkippedDeleted, []string{})
}

type fakeSkillsRunner struct {
	officialIndexOut string
	officialOut      string
	globalJSONOut    string
	globalOut        string
	officialIndexErr error
	officialErr      error
	globalJSONErr    error
	globalErr        error
	installErr       error
	installAllErr    error
	installSuiteErr  error
	installed        [][]string
	installedAll     int
	installedSuite   int
	listedIndex      int
	listedOfficial   int
	listedGlobalJSON int
	listedGlobalText int
}

func officialSkillsOutput(names ...string) string {
	var b strings.Builder
	b.WriteString("Available Skills\n")
	for _, name := range names {
		b.WriteString("│    ")
		b.WriteString(name)
		b.WriteString("\n")
	}
	return b.String()
}

func officialSkillsIndexOutput(names ...string) string {
	var b strings.Builder
	b.WriteString(`{"skills":[`)
	for i, name := range names {
		if i > 0 {
			b.WriteString(",")
		}
		fmt.Fprintf(&b, `{"name":%q,"description":"test skill","files":["SKILL.md"]}`, name)
	}
	b.WriteString(`]}`)
	return b.String()
}

func globalSkillsOutput(names ...string) string {
	var b strings.Builder
	b.WriteString("Global Skills\n\n")
	for _, name := range names {
		b.WriteString(name)
		b.WriteString(" ~/.agents/skills/")
		b.WriteString(name)
		b.WriteString("\n  Agents: Claude Code\n")
	}
	return b.String()
}

func globalSkillsJSONOutput(names ...string) string {
	var b strings.Builder
	b.WriteString("[")
	for i, name := range names {
		if i > 0 {
			b.WriteString(",")
		}
		fmt.Fprintf(&b, `{"name":%q,"path":"/Users/example/.agents/skills/%s","scope":"global","agents":["Codex"]}`, name, name)
	}
	b.WriteString("]")
	return b.String()
}

func globalSkillsJSONFromDir(dir string, names ...string) string {
	var b strings.Builder
	b.WriteString("[")
	for i, name := range names {
		if i > 0 {
			b.WriteString(",")
		}
		fmt.Fprintf(&b, `{"name":%q,"path":%q,"scope":"global","agents":["Codex"]}`, name, filepath.Join(dir, name))
	}
	b.WriteString("]")
	return b.String()
}

func createTestSkill(t *testing.T, dir, name, description string) {
	t.Helper()
	skillDir := filepath.Join(dir, name)
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := fmt.Sprintf("---\nname: %s\ndescription: %s\n---\n\n# %s\n", name, description, name)
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func createTestSuiteSkill(t *testing.T, dir string) {
	t.Helper()
	skillDir := filepath.Join(dir, suiteSkillName)
	if err := os.MkdirAll(skillDir, 0o755); err != nil {
		t.Fatal(err)
	}
	content := "---\nname: lark-suite\ndescription: suite\n---\n\n## 能力路由\n\n" + suiteRoutesPlaceholder + "\n"
	if err := os.WriteFile(filepath.Join(skillDir, "SKILL.md"), []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func (f *fakeSkillsRunner) ListOfficialSkillsIndex() *selfupdate.NpmResult {
	f.listedIndex++
	r := &selfupdate.NpmResult{}
	r.Stdout.WriteString(f.officialIndexOut)
	r.Err = f.officialIndexErr
	return r
}

func (f *fakeSkillsRunner) ListOfficialSkills() *selfupdate.NpmResult {
	f.listedOfficial++
	r := &selfupdate.NpmResult{}
	r.Stdout.WriteString(f.officialOut)
	r.Err = f.officialErr
	return r
}

func (f *fakeSkillsRunner) ListGlobalSkillsJSON() *selfupdate.NpmResult {
	f.listedGlobalJSON++
	r := &selfupdate.NpmResult{}
	r.Stdout.WriteString(f.globalJSONOut)
	r.Err = f.globalJSONErr
	return r
}

func (f *fakeSkillsRunner) ListGlobalSkills() *selfupdate.NpmResult {
	f.listedGlobalText++
	r := &selfupdate.NpmResult{}
	r.Stdout.WriteString(f.globalOut)
	r.Err = f.globalErr
	return r
}

func (f *fakeSkillsRunner) InstallSkill(nameList []string) *selfupdate.NpmResult {
	f.installed = append(f.installed, nameList)
	r := &selfupdate.NpmResult{}
	r.Err = f.installErr
	return r
}

func (f *fakeSkillsRunner) InstallAllSkills() *selfupdate.NpmResult {
	f.installedAll++
	r := &selfupdate.NpmResult{}
	r.Err = f.installAllErr
	return r
}

func (f *fakeSkillsRunner) InstallSuiteSkill() *selfupdate.NpmResult {
	f.installedSuite++
	r := &selfupdate.NpmResult{}
	r.Err = f.installSuiteErr
	return r
}

func TestSyncSkills_WritesStateAndDoesNotWriteStamp(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", dir)
	if err := WriteState(SkillsState{
		Version:        "1.0.30",
		OfficialSkills: []string{"lark-calendar", "lark-mail"},
		UpdatedAt:      "2026-05-18T00:00:00Z",
	}); err != nil {
		t.Fatal(err)
	}

	runner := &fakeSkillsRunner{
		officialIndexOut: officialSkillsIndexOutput("lark-calendar", "lark-mail", "lark-new"),
		officialOut:      officialSkillsOutput("lark-calendar", "lark-mail", "lark-new"),
		globalJSONOut:    globalSkillsJSONOutput("lark-calendar", "lark-custom"),
		globalOut:        globalSkillsOutput("lark-mail"),
	}
	result := SyncSkills(SyncOptions{
		Version: "1.0.33",
		Runner:  runner,
		Now:     func() time.Time { return time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC) },
	})

	if result.Err != nil {
		t.Fatalf("SyncSkills() err = %v, want nil", result.Err)
	}
	assertStrings(t, runner.installed[0], []string{"lark-calendar", "lark-new"})
	if runner.listedGlobalJSON != 1 {
		t.Fatalf("listedGlobalJSON = %d, want 1", runner.listedGlobalJSON)
	}
	if runner.listedGlobalText != 0 {
		t.Fatalf("listedGlobalText = %d, want 0 when JSON list succeeds", runner.listedGlobalText)
	}

	state, readable, err := ReadState()
	if err != nil || !readable {
		t.Fatalf("ReadState() = (_, %v, %v), want readable", readable, err)
	}
	assertStrings(t, state.OfficialSkills, []string{"lark-calendar", "lark-mail", "lark-new"})
	assertStrings(t, state.UpdatedSkills, []string{"lark-calendar", "lark-new"})
	assertStrings(t, state.AddedOfficialSkills, []string{"lark-new"})
	assertStrings(t, state.SkippedDeletedSkills, []string{"lark-mail"})
	if state.Layout != LayoutSeparate {
		t.Fatalf("state.Layout = %q, want %q", state.Layout, LayoutSeparate)
	}
	if _, err := os.Stat(filepath.Join(dir, "skills.stamp")); !os.IsNotExist(err) {
		t.Fatalf("skills.stamp exists or stat failed with unexpected err: %v", err)
	}
}

func TestSyncSkills_SuiteAssemblesSubskillsAndRoutes(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", t.TempDir())
	createTestSkill(t, dir, "lark-calendar", "Calendar operations")
	createTestSkill(t, dir, "lark-shared", "Shared auth and troubleshooting")
	createTestSuiteSkill(t, dir)

	runner := &fakeSkillsRunner{
		officialIndexOut: officialSkillsIndexOutput("lark-calendar", "lark-shared"),
		globalJSONOut:    globalSkillsJSONFromDir(dir, "lark-calendar", "lark-shared", "lark-suite"),
	}
	result := SyncSkills(SyncOptions{
		Version: "1.0.33",
		Layout:  LayoutSuite,
		Runner:  runner,
		Now:     func() time.Time { return time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC) },
	})

	if result.Err != nil {
		t.Fatalf("SyncSkills() err = %v, want nil", result.Err)
	}
	if runner.installedSuite != 1 {
		t.Fatalf("installedSuite = %d, want 1", runner.installedSuite)
	}
	if _, err := os.Stat(filepath.Join(dir, "lark-calendar")); !os.IsNotExist(err) {
		t.Fatalf("top-level lark-calendar still exists or stat failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "lark-suite", "references", "subskills", "lark-calendar", "SKILL.md")); err != nil {
		t.Fatalf("nested lark-calendar missing: %v", err)
	}
	suite, err := os.ReadFile(filepath.Join(dir, "lark-suite", "SKILL.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(suite), "- lark-calendar: Calendar operations") {
		t.Fatalf("suite routes were not generated from descriptions:\n%s", string(suite))
	}
	state, readable, err := ReadState()
	if err != nil || !readable {
		t.Fatalf("ReadState() = (_, %v, %v), want readable", readable, err)
	}
	if state.Layout != LayoutSuite {
		t.Fatalf("state.Layout = %q, want %q", state.Layout, LayoutSuite)
	}
}

func TestSyncSkills_HybridCopiesSharedAndMovesCollected(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", t.TempDir())
	createTestSkill(t, dir, "lark-calendar", "Calendar operations")
	createTestSkill(t, dir, "lark-mail", "Mail operations")
	createTestSkill(t, dir, "lark-shared", "Shared auth and troubleshooting")
	createTestSuiteSkill(t, dir)

	runner := &fakeSkillsRunner{
		officialIndexOut: officialSkillsIndexOutput("lark-calendar", "lark-mail", "lark-shared"),
		globalJSONOut:    globalSkillsJSONFromDir(dir, "lark-calendar", "lark-mail", "lark-shared", "lark-suite"),
	}
	result := SyncSkills(SyncOptions{
		Version:         "1.0.33",
		Layout:          LayoutHybrid,
		CollectedSkills: []string{"lark-calendar"},
		Runner:          runner,
		Now:             time.Now,
	})

	if result.Err != nil {
		t.Fatalf("SyncSkills() err = %v, want nil", result.Err)
	}
	if _, err := os.Stat(filepath.Join(dir, "lark-calendar")); !os.IsNotExist(err) {
		t.Fatalf("top-level lark-calendar still exists or stat failed: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "lark-mail", "SKILL.md")); err != nil {
		t.Fatalf("top-level lark-mail should remain: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "lark-shared", "SKILL.md")); err != nil {
		t.Fatalf("top-level lark-shared should remain in hybrid: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "lark-suite", "references", "subskills", "lark-shared", "SKILL.md")); err != nil {
		t.Fatalf("nested lark-shared copy missing: %v", err)
	}
	state, readable, err := ReadState()
	if err != nil || !readable {
		t.Fatalf("ReadState() = (_, %v, %v), want readable", readable, err)
	}
	assertStrings(t, state.CollectedSkills, []string{"lark-calendar"})
}

func TestSyncSkills_HybridCollectsNewOfficialSkillsIntoSuite(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", t.TempDir())
	if err := WriteState(SkillsState{
		Version:         "1.0.32",
		Layout:          LayoutHybrid,
		OfficialSkills:  []string{"lark-calendar", "lark-shared"},
		CollectedSkills: []string{"lark-calendar"},
		UpdatedAt:       "2026-05-18T00:00:00Z",
	}); err != nil {
		t.Fatal(err)
	}
	createTestSkill(t, dir, "lark-calendar", "Calendar operations")
	createTestSkill(t, dir, "lark-new", "New operations")
	createTestSkill(t, dir, "lark-shared", "Shared auth and troubleshooting")
	createTestSuiteSkill(t, dir)

	runner := &fakeSkillsRunner{
		officialIndexOut: officialSkillsIndexOutput("lark-calendar", "lark-new", "lark-shared"),
		globalJSONOut:    globalSkillsJSONFromDir(dir, "lark-calendar", "lark-new", "lark-shared", "lark-suite"),
	}
	result := SyncSkills(SyncOptions{
		Version:         "1.0.33",
		Layout:          LayoutHybrid,
		CollectedSkills: []string{"lark-calendar"},
		Runner:          runner,
		Now:             time.Now,
	})

	if result.Err != nil {
		t.Fatalf("SyncSkills() err = %v, want nil", result.Err)
	}
	if _, err := os.Stat(filepath.Join(dir, "lark-suite", "references", "subskills", "lark-new", "SKILL.md")); err != nil {
		t.Fatalf("new official skill should be collected into suite: %v", err)
	}
	state, readable, err := ReadState()
	if err != nil || !readable {
		t.Fatalf("ReadState() = (_, %v, %v), want readable", readable, err)
	}
	assertStrings(t, state.CollectedSkills, []string{"lark-calendar", "lark-new"})
}

func TestSyncSkills_SuiteExcludesUserDeletedSubskillAndRebuildsRoutes(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", t.TempDir())
	if err := WriteState(SkillsState{
		Version:        "1.0.32",
		Layout:         LayoutSuite,
		OfficialSkills: []string{"lark-calendar", "lark-mail", "lark-shared"},
		UpdatedAt:      "2026-05-18T00:00:00Z",
	}); err != nil {
		t.Fatal(err)
	}
	createTestSkill(t, dir, "lark-calendar", "Calendar operations")
	createTestSkill(t, dir, "lark-new", "New operations")
	createTestSkill(t, dir, "lark-shared", "Shared auth and troubleshooting")
	createTestSuiteSkill(t, dir)

	runner := &fakeSkillsRunner{
		officialIndexOut: officialSkillsIndexOutput("lark-calendar", "lark-mail", "lark-new", "lark-shared"),
		globalJSONOut:    globalSkillsJSONFromDir(dir, "lark-calendar", "lark-new", "lark-shared", "lark-suite"),
	}
	result := SyncSkills(SyncOptions{
		Version: "1.0.33",
		Layout:  LayoutSuite,
		Runner:  runner,
		Now:     time.Now,
	})

	if result.Err != nil {
		t.Fatalf("SyncSkills() err = %v, want nil", result.Err)
	}
	if _, err := os.Stat(filepath.Join(dir, "lark-suite", "references", "subskills", "lark-mail")); !os.IsNotExist(err) {
		t.Fatalf("deleted subskill should not be regenerated, stat err: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, "lark-suite", "references", "subskills", "lark-new", "SKILL.md")); err != nil {
		t.Fatalf("new official skill should be regenerated into suite: %v", err)
	}
	state, readable, err := ReadState()
	if err != nil || !readable {
		t.Fatalf("ReadState() = (_, %v, %v), want readable", readable, err)
	}
	assertStrings(t, state.SkippedDeletedSkills, []string{"lark-mail"})
}

func TestSyncSkills_SuiteInstallFailureDoesNotFallbackToSeparate(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", t.TempDir())
	runner := &fakeSkillsRunner{
		officialIndexOut: officialSkillsIndexOutput("lark-calendar", "lark-shared"),
		globalJSONOut:    globalSkillsJSONFromDir(dir, "lark-calendar", "lark-shared"),
		installErr:       fmt.Errorf("ordinary install failed"),
	}

	result := SyncSkills(SyncOptions{
		Version: "1.0.33",
		Layout:  LayoutSuite,
		Runner:  runner,
		Now:     time.Now,
	})

	if result.Action != "failed" || result.Err == nil {
		t.Fatalf("SyncSkills() = %+v, want failed result", result)
	}
	if runner.installedAll != 0 {
		t.Fatalf("installedAll = %d, want 0; suite must not silently fallback to separate", runner.installedAll)
	}
	if state, readable, err := ReadState(); err != nil || readable || state != nil {
		t.Fatalf("ReadState() = (%+v, %v, %v), want no successful state", state, readable, err)
	}
}

func TestSyncSkills_SuiteSpecialSourceFailureDoesNotWriteState(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", t.TempDir())
	createTestSkill(t, dir, "lark-calendar", "Calendar operations")
	createTestSkill(t, dir, "lark-shared", "Shared auth and troubleshooting")
	runner := &fakeSkillsRunner{
		officialIndexOut: officialSkillsIndexOutput("lark-calendar", "lark-shared"),
		globalJSONOut:    globalSkillsJSONFromDir(dir, "lark-calendar", "lark-shared"),
		installSuiteErr:  fmt.Errorf("special source failed"),
	}

	result := SyncSkills(SyncOptions{
		Version: "1.0.33",
		Layout:  LayoutSuite,
		Runner:  runner,
		Now:     time.Now,
	})

	if result.Action != "failed" || result.Err == nil {
		t.Fatalf("SyncSkills() = %+v, want failed result", result)
	}
	if runner.installedAll != 0 {
		t.Fatalf("installedAll = %d, want 0; special source failure must not fallback to separate", runner.installedAll)
	}
	if state, readable, err := ReadState(); err != nil || readable || state != nil {
		t.Fatalf("ReadState() = (%+v, %v, %v), want no successful state", state, readable, err)
	}
}

func TestSyncSkills_OfficialIndexSuccessSkipsOfficialListCommand(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", dir)
	runner := &fakeSkillsRunner{
		officialIndexOut: officialSkillsIndexOutput("lark-calendar", "lark-mail", "lark-new"),
		officialOut:      officialSkillsOutput("lark-should-not-be-used"),
		globalJSONOut:    globalSkillsJSONOutput("lark-calendar"),
		globalOut:        globalSkillsOutput("lark-mail"),
	}

	result := SyncSkills(SyncOptions{Version: "1.0.33", Runner: runner, Now: time.Now})
	if result.Err != nil {
		t.Fatalf("SyncSkills() err = %v, want nil", result.Err)
	}
	assertStrings(t, result.Official, []string{"lark-calendar", "lark-mail", "lark-new"})
	assertStrings(t, runner.installed[0], []string{"lark-calendar", "lark-mail", "lark-new"})
	if runner.listedIndex != 1 {
		t.Fatalf("listedIndex = %d, want 1", runner.listedIndex)
	}
	if runner.listedOfficial != 0 {
		t.Fatalf("listedOfficial = %d, want 0 when index succeeds", runner.listedOfficial)
	}
}

func TestSyncSkills_OfficialIndexFailureFallsBackToOfficialList(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", dir)
	runner := &fakeSkillsRunner{
		officialIndexErr: fmt.Errorf("index unavailable"),
		officialOut:      officialSkillsOutput("lark-calendar", "lark-mail"),
		globalJSONOut:    globalSkillsJSONOutput("lark-calendar"),
	}

	result := SyncSkills(SyncOptions{Version: "1.0.33", Runner: runner, Now: time.Now})
	if result.Err != nil {
		t.Fatalf("SyncSkills() err = %v, want nil", result.Err)
	}
	assertStrings(t, result.Official, []string{"lark-calendar", "lark-mail"})
	if runner.listedIndex != 1 || runner.listedOfficial != 1 {
		t.Fatalf("listed index/official = %d/%d, want 1/1", runner.listedIndex, runner.listedOfficial)
	}
	if runner.installedAll != 0 {
		t.Fatalf("installedAll = %d, want 0", runner.installedAll)
	}
}

func TestSyncSkills_OfficialIndexEmptyFallsBackToOfficialList(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", dir)
	runner := &fakeSkillsRunner{
		officialIndexOut: `{"skills":[]}`,
		officialOut:      officialSkillsOutput("lark-calendar", "lark-mail"),
		globalJSONOut:    globalSkillsJSONOutput("lark-calendar"),
	}

	result := SyncSkills(SyncOptions{Version: "1.0.33", Runner: runner, Now: time.Now})
	if result.Err != nil {
		t.Fatalf("SyncSkills() err = %v, want nil", result.Err)
	}
	assertStrings(t, result.Official, []string{"lark-calendar", "lark-mail"})
	if runner.listedIndex != 1 || runner.listedOfficial != 1 {
		t.Fatalf("listed index/official = %d/%d, want 1/1", runner.listedIndex, runner.listedOfficial)
	}
}

func TestSyncSkills_OfficialDiscoveryFailuresFallBackToFullInstallWithReasons(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", dir)
	runner := &fakeSkillsRunner{
		officialIndexErr: fmt.Errorf("index unavailable"),
		officialErr:      fmt.Errorf("list failed"),
		installAllErr:    nil,
	}

	result := SyncSkills(SyncOptions{Version: "1.0.33", Runner: runner, Now: time.Now})
	if result.Action != "fallback_synced" {
		t.Fatalf("SyncSkills() action = %q, want fallback_synced", result.Action)
	}
	if runner.installedAll != 1 {
		t.Fatalf("installedAll = %d, want 1", runner.installedAll)
	}
	if !strings.Contains(result.Detail, "official skills index failed") || !strings.Contains(result.Detail, "official skills list failed") {
		t.Fatalf("SyncSkills() detail = %q, want both discovery failure reasons", result.Detail)
	}
}

func TestSyncSkills_OfficialDiscoveryEmptyFallsBackToFullInstallWithReasons(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", dir)
	runner := &fakeSkillsRunner{
		officialIndexOut: `{"skills":[]}`,
		installAllErr:    nil,
	}

	result := SyncSkills(SyncOptions{Version: "1.0.33", Runner: runner, Now: time.Now})
	if result.Action != "fallback_synced" {
		t.Fatalf("SyncSkills() action = %q, want fallback_synced", result.Action)
	}
	if runner.installedAll != 1 {
		t.Fatalf("installedAll = %d, want 1", runner.installedAll)
	}
	if !strings.Contains(result.Detail, "official skills index contains no skills") || !strings.Contains(result.Detail, "official skills list returned no skills") {
		t.Fatalf("SyncSkills() detail = %q, want both empty discovery reasons", result.Detail)
	}
}

func TestSyncSkills_ListOfficialFailureFallsBackToFullInstall(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", dir)
	runner := &fakeSkillsRunner{
		officialIndexErr: fmt.Errorf("index unavailable"),
		officialErr:      fmt.Errorf("list failed"),
		installAllErr:    nil,
	}

	result := SyncSkills(SyncOptions{Version: "1.0.33", Runner: runner, Now: time.Now})
	if result.Action != "fallback_synced" {
		t.Fatalf("SyncSkills() action = %q, want fallback_synced", result.Action)
	}
	if runner.installedAll != 1 {
		t.Fatalf("installedAll = %d, want 1", runner.installedAll)
	}
	if len(runner.installed) != 0 {
		t.Fatalf("installed = %#v, want no incremental installs", runner.installed)
	}

	state, readable, err := ReadState()
	if err != nil || !readable {
		t.Fatalf("ReadState() = (_, %v, %v), want readable", readable, err)
	}
	if state.Version != "1.0.33" {
		t.Fatalf("state.Version = %q, want %q", state.Version, "1.0.33")
	}
	assertStrings(t, state.OfficialSkills, []string{})
}

func TestSyncSkills_ListOfficialFailureAndFullInstallFails(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", dir)
	runner := &fakeSkillsRunner{
		officialIndexErr: fmt.Errorf("index unavailable"),
		officialErr:      fmt.Errorf("list failed"),
		installAllErr:    fmt.Errorf("full install failed"),
	}

	result := SyncSkills(SyncOptions{Version: "1.0.33", Runner: runner, Now: time.Now})
	if result.Action != "fallback_failed" {
		t.Fatalf("SyncSkills() action = %q, want fallback_failed", result.Action)
	}
	if result.Err == nil {
		t.Fatalf("SyncSkills() err = nil, want error")
	}
	if !strings.Contains(result.Err.Error(), "full skills install failed") {
		t.Fatalf("SyncSkills() err = %v, want full install failure", result.Err)
	}
}

func TestSyncSkills_GlobalJSONFailureFallsBackToTextList(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", dir)
	runner := &fakeSkillsRunner{
		officialIndexOut: officialSkillsIndexOutput("lark-calendar", "lark-mail"),
		officialOut:      officialSkillsOutput("lark-calendar", "lark-mail"),
		globalJSONErr:    fmt.Errorf("json list failed"),
		globalOut:        globalSkillsOutput("lark-calendar"),
	}

	result := SyncSkills(SyncOptions{Version: "1.0.33", Runner: runner, Now: time.Now})
	if result.Err != nil {
		t.Fatalf("SyncSkills() err = %v, want nil", result.Err)
	}
	if result.Action != "synced" {
		t.Fatalf("SyncSkills() action = %q, want synced", result.Action)
	}
	assertStrings(t, result.Updated, []string{"lark-calendar", "lark-mail"})
	if runner.listedGlobalJSON != 1 || runner.listedGlobalText != 1 {
		t.Fatalf("listed JSON/text = %d/%d, want 1/1", runner.listedGlobalJSON, runner.listedGlobalText)
	}
	if runner.installedAll != 0 {
		t.Fatalf("installedAll = %d, want 0", runner.installedAll)
	}
}

func TestSyncSkills_LocalListsFailureFallsBackToFullInstall(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", dir)
	runner := &fakeSkillsRunner{
		officialIndexOut: officialSkillsIndexOutput("lark-calendar", "lark-mail"),
		officialOut:      officialSkillsOutput("lark-calendar", "lark-mail"),
		globalJSONErr:    fmt.Errorf("json list failed with /Users/example/.agents/skills/lark-calendar agents Codex"),
		globalErr:        fmt.Errorf("text list failed with /Users/example/.agents/skills/lark-mail agents Codex"),
	}

	result := SyncSkills(SyncOptions{Version: "1.0.33", Runner: runner, Now: time.Now})
	if result.Action != "fallback_synced" {
		t.Fatalf("SyncSkills() action = %q, want fallback_synced", result.Action)
	}
	if len(runner.installed) != 0 {
		t.Fatalf("installed = %#v, want no incremental installs", runner.installed)
	}
	if runner.installedAll != 1 {
		t.Fatalf("installedAll = %d, want 1", runner.installedAll)
	}
	if strings.Contains(result.Detail, "/Users/example") || strings.Contains(result.Detail, "agents") {
		t.Fatalf("SyncSkills() detail leaks local command output: %q", result.Detail)
	}
}

func TestSyncSkills_EmptyLocalJSONInstallsAllOfficialIncrementally(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", dir)
	runner := &fakeSkillsRunner{
		officialIndexOut: officialSkillsIndexOutput("lark-calendar", "lark-mail"),
		officialOut:      officialSkillsOutput("lark-calendar", "lark-mail"),
		globalJSONOut:    `[]`,
		globalOut:        "Some unrecognized output format\n",
	}

	result := SyncSkills(SyncOptions{Version: "1.0.33", Runner: runner, Now: time.Now})
	if result.Action != "synced" {
		t.Fatalf("SyncSkills() action = %q, want synced", result.Action)
	}
	if len(runner.installed) != 1 {
		t.Fatalf("installed = %#v, want one incremental install", runner.installed)
	}
	assertStrings(t, runner.installed[0], []string{"lark-calendar", "lark-mail"})
	if runner.installedAll != 0 {
		t.Fatalf("installedAll = %d, want 0", runner.installedAll)
	}
}

func TestSyncSkills_EmptyToUpdateFallsBackToFullInstall(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", dir)
	if err := WriteState(SkillsState{
		Version:        "1.0.30",
		OfficialSkills: []string{"lark-calendar", "lark-mail"},
		UpdatedAt:      "2026-05-18T00:00:00Z",
	}); err != nil {
		t.Fatal(err)
	}

	runner := &fakeSkillsRunner{
		officialIndexOut: officialSkillsIndexOutput("lark-calendar", "lark-mail"),
		officialOut:      officialSkillsOutput("lark-calendar", "lark-mail"),
		globalOut:        globalSkillsOutput(),
		installAllErr:    nil,
	}

	result := SyncSkills(SyncOptions{Version: "1.0.33", Runner: runner, Now: time.Now})
	if result.Action != "fallback_synced" {
		t.Fatalf("SyncSkills() action = %q, want fallback_synced", result.Action)
	}
	if len(runner.installed) != 0 {
		t.Fatalf("installed = %#v, want no incremental installs", runner.installed)
	}
	if runner.installedAll != 1 {
		t.Fatalf("installedAll = %d, want 1 (fallback triggered)", runner.installedAll)
	}
	assertStrings(t, result.Official, []string{"lark-calendar", "lark-mail"})
	assertStrings(t, result.Updated, []string{"lark-calendar", "lark-mail"})
	assertStrings(t, result.Added, []string{"lark-calendar", "lark-mail"})
	assertStrings(t, result.SkippedDeleted, []string{})
}

func TestSyncSkills_InstallFailureFallsBackToFullInstall(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", dir)
	runner := &fakeSkillsRunner{
		officialIndexOut: officialSkillsIndexOutput("lark-calendar", "lark-mail"),
		officialOut:      officialSkillsOutput("lark-calendar", "lark-mail"),
		globalJSONOut:    globalSkillsJSONOutput("lark-calendar", "lark-mail"),
		globalOut:        globalSkillsOutput("lark-calendar", "lark-mail"),
		installErr:       fmt.Errorf("incremental boom"),
		installAllErr:    nil,
	}

	result := SyncSkills(SyncOptions{Version: "1.0.33", Runner: runner, Now: time.Now})
	if result.Action != "fallback_synced" {
		t.Fatalf("SyncSkills() action = %q, want fallback_synced", result.Action)
	}
	if len(runner.installed) != 1 {
		t.Fatalf("installed = %d calls, want 1", len(runner.installed))
	}
	if runner.installedAll != 1 {
		t.Fatalf("installedAll = %d, want 1 (fallback triggered)", runner.installedAll)
	}

	state, readable, err := ReadState()
	if err != nil || !readable {
		t.Fatalf("ReadState() = (_, %v, %v), want readable", readable, err)
	}
	if state.Version != "1.0.33" {
		t.Fatalf("state.Version = %q, want %q", state.Version, "1.0.33")
	}
	assertStrings(t, state.OfficialSkills, []string{"lark-calendar", "lark-mail"})
}

func TestSyncSkills_InstallFailureAndFullInstallFails(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", dir)
	runner := &fakeSkillsRunner{
		officialIndexOut: officialSkillsIndexOutput("lark-calendar", "lark-mail"),
		officialOut:      officialSkillsOutput("lark-calendar", "lark-mail"),
		globalJSONOut:    globalSkillsJSONOutput("lark-calendar", "lark-mail"),
		globalOut:        globalSkillsOutput("lark-calendar", "lark-mail"),
		installErr:       fmt.Errorf("incremental boom"),
		installAllErr:    fmt.Errorf("full install boom"),
	}

	result := SyncSkills(SyncOptions{Version: "1.0.33", Runner: runner, Now: time.Now})
	if result.Action != "fallback_failed" {
		t.Fatalf("SyncSkills() action = %q, want fallback_failed", result.Action)
	}
	if result.Err == nil {
		t.Fatalf("SyncSkills() err = nil, want error")
	}
	if !strings.Contains(result.Detail, "incremental boom") {
		t.Fatalf("SyncSkills() detail = %q, want incremental error text", result.Detail)
	}
	if !strings.Contains(result.Err.Error(), "full skills install failed") {
		t.Fatalf("SyncSkills() err = %v, want full install failure", result.Err)
	}
}

func TestSyncSkills_NilRunnerFails(t *testing.T) {
	result := SyncSkills(SyncOptions{Version: "1.0.33", Now: time.Now})
	if result.Err == nil || !strings.Contains(result.Err.Error(), "skills runner is nil") {
		t.Fatalf("SyncSkills() err = %v, want nil runner failure", result.Err)
	}
}

func TestSyncSkills_ParseEmptyWithNonEmptyStdoutFallsBack(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", dir)
	runner := &fakeSkillsRunner{
		officialIndexErr: fmt.Errorf("index unavailable"),
		officialOut:      "Some unrecognized output format\n",
		installAllErr:    nil,
	}

	result := SyncSkills(SyncOptions{Version: "1.0.33", Runner: runner, Now: time.Now})
	if result.Action != "fallback_synced" {
		t.Fatalf("SyncSkills() action = %q, want fallback_synced", result.Action)
	}
	if runner.installedAll != 1 {
		t.Fatalf("installedAll = %d, want 1", runner.installedAll)
	}
}

func TestSyncSkills_ParseEmptyWithNonEmptyStdoutAndFullInstallFails(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", dir)
	runner := &fakeSkillsRunner{
		officialIndexErr: fmt.Errorf("index unavailable"),
		officialOut:      "Some unrecognized output format\n",
		installAllErr:    fmt.Errorf("full install failed"),
	}

	result := SyncSkills(SyncOptions{Version: "1.0.33", Runner: runner, Now: time.Now})
	if result.Action != "fallback_failed" {
		t.Fatalf("SyncSkills() action = %q, want fallback_failed", result.Action)
	}
	if result.Err == nil {
		t.Fatalf("SyncSkills() err = nil, want error")
	}
}

func assertStrings(t *testing.T, got, want []string) {
	t.Helper()
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("got %#v, want %#v", got, want)
	}
}

func TestSyncSkills_FallbackWithUnknownOfficialWritesMinimalState(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", dir)
	runner := &fakeSkillsRunner{
		officialIndexErr: fmt.Errorf("index unavailable"),
		officialOut:      "Some unrecognized output format\n",
		installAllErr:    nil,
	}

	result := SyncSkills(SyncOptions{Version: "1.0.33", Runner: runner, Now: time.Now})
	if result.Action != "fallback_synced" {
		t.Fatalf("SyncSkills() action = %q, want fallback_synced", result.Action)
	}

	state, readable, err := ReadState()
	if err != nil || !readable {
		t.Fatalf("ReadState() = (_, %v, %v), want readable", readable, err)
	}
	if state.Version != "1.0.33" {
		t.Fatalf("state.Version = %q, want %q", state.Version, "1.0.33")
	}
	assertStrings(t, state.OfficialSkills, []string{})
	assertStrings(t, state.UpdatedSkills, []string{})
	assertStrings(t, state.AddedOfficialSkills, []string{})
}

func TestSyncSkills_FallbackWithKnownOfficialWritesFullState(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", dir)
	runner := &fakeSkillsRunner{
		officialIndexOut: officialSkillsIndexOutput("lark-calendar", "lark-mail"),
		officialOut:      officialSkillsOutput("lark-calendar", "lark-mail"),
		globalJSONOut:    globalSkillsJSONOutput("lark-calendar", "lark-mail"),
		globalOut:        globalSkillsOutput("lark-calendar", "lark-mail"),
		installErr:       fmt.Errorf("incremental boom"),
		installAllErr:    nil,
	}

	result := SyncSkills(SyncOptions{Version: "1.0.33", Runner: runner, Now: time.Now})
	if result.Action != "fallback_synced" {
		t.Fatalf("SyncSkills() action = %q, want fallback_synced", result.Action)
	}

	state, readable, err := ReadState()
	if err != nil || !readable {
		t.Fatalf("ReadState() = (_, %v, %v), want readable", readable, err)
	}
	assertStrings(t, state.OfficialSkills, []string{"lark-calendar", "lark-mail"})
	assertStrings(t, state.UpdatedSkills, []string{"lark-calendar", "lark-mail"})
	assertStrings(t, state.AddedOfficialSkills, []string{"lark-calendar", "lark-mail"})
}

func TestSyncSkills_FallbackResultContainsMetadata(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", dir)
	runner := &fakeSkillsRunner{
		officialIndexOut: officialSkillsIndexOutput("lark-calendar", "lark-mail"),
		officialOut:      officialSkillsOutput("lark-calendar", "lark-mail"),
		globalJSONOut:    globalSkillsJSONOutput("lark-calendar", "lark-mail"),
		globalOut:        globalSkillsOutput("lark-calendar", "lark-mail"),
		installErr:       fmt.Errorf("incremental boom"),
		installAllErr:    nil,
	}

	result := SyncSkills(SyncOptions{Version: "1.0.33", Runner: runner, Now: time.Now})
	if result.Action != "fallback_synced" {
		t.Fatalf("SyncSkills() action = %q, want fallback_synced", result.Action)
	}
	assertStrings(t, result.Official, []string{"lark-calendar", "lark-mail"})
	assertStrings(t, result.Updated, []string{"lark-calendar", "lark-mail"})
	assertStrings(t, result.Added, []string{"lark-calendar", "lark-mail"})
	assertStrings(t, result.SkippedDeleted, []string{})
	if !strings.Contains(result.Detail, "incremental boom") {
		t.Fatalf("SyncSkills() detail = %q, want incremental error text", result.Detail)
	}
}

func TestSyncSkills_FallbackBreaksDegradationLoop(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("LARKSUITE_CLI_CONFIG_DIR", dir)
	runner := &fakeSkillsRunner{
		officialIndexErr: fmt.Errorf("index unavailable"),
		officialErr:      fmt.Errorf("list failed"),
		installAllErr:    nil,
	}

	result1 := SyncSkills(SyncOptions{Version: "1.0.33", Runner: runner, Now: time.Now})
	if result1.Action != "fallback_synced" {
		t.Fatalf("first sync: action = %q, want fallback_synced", result1.Action)
	}

	state, readable, err := ReadState()
	if err != nil || !readable {
		t.Fatalf("ReadState() after first sync = (_, %v, %v), want readable", readable, err)
	}
	if state.Version != "1.0.33" {
		t.Fatalf("state.Version = %q, want %q", state.Version, "1.0.33")
	}

	runner2 := &fakeSkillsRunner{
		officialIndexOut: officialSkillsIndexOutput("lark-calendar", "lark-mail"),
		officialOut:      officialSkillsOutput("lark-calendar", "lark-mail"),
		globalJSONOut:    globalSkillsJSONOutput("lark-calendar", "lark-mail"),
		globalOut:        globalSkillsOutput("lark-calendar", "lark-mail"),
	}
	result2 := SyncSkills(SyncOptions{Version: "1.0.33", Runner: runner2, Now: time.Now})
	if result2.Action != "synced" {
		t.Fatalf("second sync: action = %q, want synced (no fallback loop)", result2.Action)
	}
	if runner2.installedAll != 0 {
		t.Fatalf("second sync: installedAll = %d, want 0 (incremental, not fallback)", runner2.installedAll)
	}
}
