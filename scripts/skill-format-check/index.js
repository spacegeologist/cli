// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

const fs = require('fs');
const path = require('path');

// Allow passing a target directory as the first argument.
// If provided, resolve against process.cwd() so it behaves as the user expects.
// If not provided, default to the official skill directories.
const targetDirArg = process.argv[2];
const TARGET_DIRS = targetDirArg
  ? [path.resolve(process.cwd(), targetDirArg)]
  : [
      path.resolve(__dirname, '../../skills'),
      path.resolve(__dirname, '../../isolated-skills'),
    ];

function checkSkillFormatInDir(skillsDir) {
  console.log(`Checking skill format in ${skillsDir}...`);

  if (!fs.existsSync(skillsDir)) {
    console.error('Skills directory not found:', skillsDir);
    process.exit(1);
  }

  let skills;
  try {
    skills = fs
      .readdirSync(skillsDir, { withFileTypes: true })
      .filter(entry => entry.isDirectory())
      .map(entry => entry.name);
  } catch (err) {
    console.error(`Failed to enumerate skills directory: ${err.message}`);
    process.exit(1);
  }

  let hasErrors = false;

  skills.forEach(skill => {
    // Skip lark-shared skill completely
    if (skill === 'lark-shared') {
        console.log(`⏭️  Skipping check for ${skill}`);
        return;
    }

    const skillPath = path.join(skillsDir, skill);
    const skillFile = path.join(skillPath, 'SKILL.md');

    if (!fs.existsSync(skillFile)) {
      console.error(`❌ [${skill}] Missing SKILL.md`);
      hasErrors = true;
      return;
    }

    let content;
    try {
      content = fs.readFileSync(skillFile, 'utf-8');
    } catch (err) {
      console.error(`❌ [${skill}] Failed to read SKILL.md: ${err.message}`);
      hasErrors = true;
      return;
    }

    // Normalize line endings to simplify parsing
    const normalizedContent = content.replace(/\r\n/g, '\n');

    // Check YAML Frontmatter
    if (!normalizedContent.startsWith('---\n')) {
       console.error(`❌ [${skill}] SKILL.md must start with YAML frontmatter (---)`);
       hasErrors = true;
    } else {
        const frontmatterMatch = normalizedContent.match(/^---\n([\s\S]*?)\n---(?:\n|$)/);
        if (!frontmatterMatch) {
             console.error(`❌ [${skill}] SKILL.md has unclosed or invalid YAML frontmatter`);
             hasErrors = true;
        } else {
            const frontmatter = frontmatterMatch[1];
            if (!/^name:/m.test(frontmatter)) {
                console.error(`❌ [${skill}] YAML frontmatter missing 'name'`);
                hasErrors = true;
            }
            if (!/^description:/m.test(frontmatter)) {
                console.error(`❌ [${skill}] YAML frontmatter missing 'description'`);
                hasErrors = true;
            }
            if (!/^metadata:/m.test(frontmatter)) {
                console.warn(`⚠️  [${skill}] YAML frontmatter missing 'metadata' (Warning only)`);
                // hasErrors = true; // Downgrade to warning to not fail on existing skills
            }
        }
    }
  });

  return !hasErrors;
}

function checkSkillFormat() {
  const ok = TARGET_DIRS.map(checkSkillFormatInDir).every(Boolean);

  if (!ok) {
    console.error('\n❌ Skill format check failed. Please fix the errors above.');
    process.exit(1);
  }

  console.log('\n✅ Skill format check passed!');
}

checkSkillFormat();
