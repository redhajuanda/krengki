---
name: /png-split-transparent
id: png-split-transparent
category: Media
description: Split one transparent PNG into multiple standalone PNG files by extracting disconnected visible regions
---

Split a PNG with a transparent background into multiple cropped PNG files. This command runs the `png-split-transparent` skill.

**Skill**: Use `.agents/skills/png-split-transparent/SKILL.md` and its `scripts/` CLI.

**Input**: The user should provide:
- Source PNG path
- Optional `padding`
- Optional `min-pixels`
- Optional `alpha-threshold`
- Optional output directory

If the source path is missing, ask for it before running anything.

## Steps

1. Read the `png-split-transparent` skill.
2. Run the CLI from the skill directory:

```bash
python3 -m scripts /absolute/path/to/input.png
```

3. If the user needs tuning, rerun with flags such as:

```bash
python3 -m scripts /absolute/path/to/input.png --padding 24 --min-pixels 400 --alpha-threshold 8
```

4. Report:
- output directory
- number of extracted PNG files
- whether any tiny fragments were filtered out by thresholds
