#!/usr/bin/env node
// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

const fs = require("fs");
const path = require("path");

const repoRoot = path.resolve(__dirname, "..");
const skillsDir = path.join(repoRoot, "skills");
const templatePath = path.join(repoRoot, "isolated-skills", "lark-suite", "SKILL.md");
const outPath = path.join(repoRoot, "lark-suite.generated-from-descriptions.SKILL.md");
const statsPath = path.join(repoRoot, "lark-suite.generated-from-descriptions.stats.json");

function read(file) {
  return fs.readFileSync(file, "utf8");
}

function parseFrontmatterDescription(skillPath) {
  const text = read(skillPath);
  const match = text.match(/^---\r?\n([\s\S]*?)\r?\n---/);
  if (!match) {
    throw new Error(`missing frontmatter: ${skillPath}`);
  }

  const lines = match[1].split(/\r?\n/);
  for (let i = 0; i < lines.length; i++) {
    const line = lines[i];
    const desc = line.match(/^description:\s*(.*)$/);
    if (!desc) continue;
    const value = desc[1].trim();
    if (value === ">" || value === "|") {
      return parseIndentedBlock(lines, i + 1, value);
    }
    return unquoteYamlScalar(value);
  }
  throw new Error(`missing frontmatter description: ${skillPath}`);
}

function parseIndentedBlock(lines, start, style) {
  const block = [];
  for (let i = start; i < lines.length; i++) {
    const line = lines[i];
    if (!line.trim()) {
      block.push("");
      continue;
    }
    if (!line.startsWith(" ")) break;
    block.push(line.replace(/^  ?/, ""));
  }
  if (style === "|") {
    return block.join("\n").trim();
  }
  return block.map((line) => line.trim()).filter(Boolean).join(" ");
}

function unquoteYamlScalar(value) {
  if (
    (value.startsWith('"') && value.endsWith('"')) ||
    (value.startsWith("'") && value.endsWith("'"))
  ) {
    return value.slice(1, -1);
  }
  return value;
}

function approxTokens(text) {
  // Crude but stable enough for comparing generated variants.
  return Math.ceil(Array.from(text).length / 1.7);
}

function collectDescriptions() {
  const descriptions = new Map();
  for (const entry of fs.readdirSync(skillsDir, { withFileTypes: true })) {
    if (!entry.isDirectory()) continue;
    const name = entry.name;
    if (!name.startsWith("lark-") || name === "lark-suite") continue;
    const skillPath = path.join(skillsDir, name, "SKILL.md");
    if (!fs.existsSync(skillPath)) continue;
    descriptions.set(name, parseFrontmatterDescription(skillPath));
  }
  return descriptions;
}

function generate(template, descriptions) {
  const used = Array.from(descriptions.keys()).sort();
  const routes = used.map((skillName) => {
    const description = descriptions.get(skillName);
    return `- ${skillName}: ${description}`;
  });
  if (!template.includes("<!-- LARK_SUITE_ROUTES -->")) {
    throw new Error("missing route placeholder in lark-suite template");
  }

  return {
    text: template.replace("<!-- LARK_SUITE_ROUTES -->", routes.join("\n")),
    used,
  };
}

function main() {
  const template = read(templatePath);
  const descriptions = collectDescriptions();
  const { text, used } = generate(template, descriptions);
  fs.writeFileSync(outPath, text.endsWith("\n") ? text : `${text}\n`);

  const stats = {
    generated_file: outPath,
    template_file: templatePath,
    rule: "Fill isolated-skills/lark-suite/SKILL.md route placeholder from installed skill descriptions.",
    skill_count: used.length,
    generated_bytes: Buffer.byteLength(text),
    generated_chars: Array.from(text).length,
    generated_approx_tokens: approxTokens(text),
    template_bytes: Buffer.byteLength(template),
    template_chars: Array.from(template).length,
    template_approx_tokens: approxTokens(template),
    route_entries: used,
    entry_token_stats: used
      .map((skill) => {
        const description = descriptions.get(skill);
        return [skill, approxTokens(description), Array.from(description).length];
      })
      .sort((a, b) => b[1] - a[1] || a[0].localeCompare(b[0])),
  };
  fs.writeFileSync(statsPath, `${JSON.stringify(stats, null, 2)}\n`);
  console.log(`Generated ${path.relative(repoRoot, outPath)} from ${path.relative(repoRoot, templatePath)}`);
  console.log(`Updated ${path.relative(repoRoot, statsPath)}`);
}

main();
