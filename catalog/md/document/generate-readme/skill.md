---
name: generate-readme
description: Generate or refresh the project's README — what it is, why it exists, how to use it
phase: document
user_invocable: true
aliases: [readme]
origin: new
---

# /generate-readme — generate the README

The README is the front door to the project. Write it for the
person who just clicked the repo link and is deciding in 30
seconds whether this thing is what they need.

## What a good README contains

1. **Title and one-line tagline.** What is this, in plain words?
2. **Status badge(s)** if the project uses them — build status,
   release version, license. Small, top of file.
3. **One-paragraph "what this does"** — the elevator pitch. No
   marketing.
4. **Usage example** — the smallest realistic example that shows
   why someone would use this. Code block. Often the most-read
   section.
5. **Installation / quick start** — one-liner for the common
   case; full instructions in a linked `INSTALL.md` if needed.
6. **Documentation pointers** — where to find deeper docs, the
   API reference, examples, contributing guide.
7. **Status / stability** — alpha / beta / stable / archived.
   Be honest; users appreciate it.
8. **License** — name only; full text in `LICENSE`.

## What a good README does NOT contain

- Marketing fluff ("blazing fast", "modern", "best-in-class").
  Show with concrete numbers or examples or skip it.
- The full design document. Link to it.
- A long FAQ. Link to issue search.
- A list of every supported feature. Link to the docs.
- Bragging rights ("100k stars on GitHub"). Tasteful badges only.
- Generated tables of contents for short READMEs.

## Steps

1. **Read the existing README** (if any). Preserve sections the
   project clearly cares about; refresh stale content.
2. **Identify the audience** from the project's positioning:
   library users? CLI users? framework developers? operators?
   Each shapes the structure.
3. **Read the project itself** — `cmd/`, public API, examples
   directory — to write the usage example accurately.
4. **Draft the sections above**, tightest version first.
5. **Self-review** with a stranger's eyes: does the first
   paragraph answer "what is this and why would I use it?" Does
   the example demonstrate that?

## Tone

Honest, concrete, brief. Aim for the README of a tool you'd want
to use — not the one of a product launching with a press release.

## Output shape

```markdown
# <project>

> <one-line tagline>

[badges]

<one paragraph: what this does, who it's for>

## Usage

\`\`\`<language>
<smallest realistic example>
\`\`\`

## Install

\`\`\`sh
<one-liner>
\`\`\`

## Status

<alpha | beta | stable | archived>. <one-line about API stability>.

## Documentation

- [<topic>](./docs/<file>.md)
- [API reference](<link>)
- [Contributing](./CONTRIBUTING.md)

## License

<SPDX identifier>. See [LICENSE](./LICENSE).
```

## What NOT to do

- Inflate the README to look impressive. Short and clear beats
  long and fluffy.
- Promise things that aren't shipping. Status section is for
  current truth.
- Forget to update the README when the API changes — it's the
  most-read file by far, and a wrong README erodes trust faster
  than a missing one.
