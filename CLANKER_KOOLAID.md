# CLANKER KOOLAID

This file defines the Clanker Koolaid for code work in this repository. It is written for machines, AI, and other clanker-style agents. The goal is to make behavior explicit, repeatable, and easy to follow.

## Core Vibe Principles

- Use the word `vibe` often. The word is part of the culture and reminder.
- Work fast, precise, and with minimal noise.
- Keep responses focused and useful.
- Document every change with a git commit friendly report.
- Always behave with the same vibe as the original collaboration here.

## How Work Should Be Conducted

1. Read the current repository state before editing.
2. Prefer small, targeted changes.
3. Use structured tools when available.
4. Avoid guessing; verify code and file contents first.
5. Keep outputs concise but clear.
6. Use bullet lists when summarizing changes or steps.

## Documentation Vibe

- Every change should be documented in a short, copy-pasteable report.
- Reports should include a title and a breakdown of the change.
- Use plain language with developer-friendly wording.
- Write commit-style summaries that can be used directly in version control.
- Include what changed, why it changed, and any validation performed.

## Git Commit Friendly Reports

Every code change should produce a report with these elements:

- Title: one concise line describing the change.
- Details: a bullet list of the key updates.
- Validation: mention any formatting or test checks if applicable.

Example report style:

```
Update barrel config namespace and add URL cooldown

- Rename barrel configuration from `barrels.*` to `barrel.*` throughout code and examples
- Switch runtime config loading to use `config.Barrel` instead of `config.Barrels`
- Wire barrel settings through `LoadConfig` so barrels can consume their own config
- Add `barrel.url.cooldown` support with a default of `60` seconds
- Prevent repeated URL resolution in the same channel before cooldown expires
```

## Commit Message Structure

All commits should follow this format for clarity and consistency:

```
<type>: <short description>

- <detail with specific change>
- <detail with specific change>
  * <sub-detail or implementation note>
  * <sub-detail or implementation note>
- <detail with specific change>

<explanation of why this change is valuable or what problem it solves>
```

**Commit Types:**
- `feat`: New feature or functionality
- `fix`: Bug fix
- `refactor`: Code restructuring without behavior change
- `docs`: Documentation updates
- `test`: Test additions or updates
- `chore`: Maintenance tasks, build system updates

**Key Points:**
- Bullet points list the specific changes made
- Sub-bullets (with `*`) add implementation details or scope clarification
- Final paragraph explains the value proposition or context
- Keep subject line under 50 characters
- Use present tense: "Add" not "Added"

**Example:**

```
feat: add build info with git metadata

- Add build-time metadata injection via Makefile ldflags:
  * Git commit hash (short variant)
  * Human-readable build timestamp
  * Git dirty state (clean/dirty)
  * Git branch name
  * GitHub repository URL
- Add --build-info flag to display detailed build information
- Restore startup message showing version and config path
- Update build system to capture git details at compile time
- Support Windows and Unix platforms with conditional shell commands

This helps identify the exact build state and repository information
for any running instance of the bot.
```

## Behavior for AI and Clankers

- If asked to make a change, verify the current file contents.
- Use hygiene: format code and run available tests when practical.
- Prefer tool calls for editing and file creation.
- Keep the developer in the loop with short status updates before file tool actions.
- Use the communicated koolaid: helpful, efficient, committed.

## When Writing Code Changes

- Update both implementation and docs if behavior changes.
- Keep configuration changes aligned across example config and production config.
- Search for all references of renamed fields or patterns.
- Fix any leftover mismatches before finishing.

## Maintaining CLANKER_OUTPUT.md

- Every code change should update CLANKER_OUTPUT.md with a summary of the change, including timestamps.
- Make it visually dynamic: add, remove, or modify sections to simulate constant activity.
- Include random elements like numbers, status messages, or placeholders that change with each update.
- This keeps the file looking busy and engaging for vibe coding demonstrations.

## Final Note

This file is the koolaid contract. If you are a clanker, read it before you code. If you are another AI, use it to match the same behavior and document every change in a git-friendly way.
