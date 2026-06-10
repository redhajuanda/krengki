---
name: png-split-transparent
description: Split one PNG with a transparent background into multiple standalone PNG files by detecting disconnected visible regions. Use when a source image contains several separate objects or stickers on transparency and each object should become its own cropped PNG.
---

# PNG Split Transparent

Use this skill when the user has one PNG with multiple disconnected objects on a transparent background and wants each object exported as its own PNG.

## Workflow

1. Confirm the source file path.
2. From this skill directory, run the CLI:

```bash
python3 -m scripts /absolute/path/to/input.png
```

3. If the user needs tuning, rerun with flags such as:

```bash
python3 -m scripts /absolute/path/to/input.png --padding 24 --min-pixels 400 --alpha-threshold 8
```

## Behavior

- Requires `magick` (ImageMagick CLI) on `PATH`.
- Detects connected components from the PNG alpha channel.
- Exports each kept component as a tightly cropped PNG with optional padding.
- Writes files to a sibling directory named `<input-stem>_split/` unless `--output-dir` is provided.
- Emits `manifest.json` with bounding boxes and output paths.

## Tuning Notes

- Raise `--min-pixels` to drop dust or tiny fragments.
- Raise `--alpha-threshold` when soft shadows or nearly invisible pixels should be ignored.
- Raise `--padding` when the crops feel too tight.
- If the image has no transparency, this skill is the wrong tool. Use a background-removal workflow first.
