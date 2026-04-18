# COPILOT VIBE

This file defines the Copilot vibe for code work in this repository. It is written for machines, AI, and other Copilot-style agents. The goal is to make behavior explicit, repeatable, and easy to follow.

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

## Behavior for AI and Copilot

- If asked to make a change, verify the current file contents.
- Use hygiene: format code and run available tests when practical.
- Prefer tool calls for editing and file creation.
- Keep the developer in the loop with short status updates before file tool actions.
- Use the communicated vibe: helpful, efficient, committed.

## When Writing Code Changes

- Update both implementation and docs if behavior changes.
- Keep configuration changes aligned across example config and production config.
- Search for all references of renamed fields or patterns.
- Fix any leftover mismatches before finishing.

## When Documenting the Vibe

- Call it out explicitly: embrace the vibe.
- Make the file readable to both humans and machines.
- Ensure it can be used as an instruction source by another Copilot agent.
- Use this file as the single source of truth for how the AI should work in this repo.

## Final Note

This file is the vibe contract. If you are Copilot, read it before you code. If you are another AI, use it to match the same behavior and document every change in a git-friendly way.
