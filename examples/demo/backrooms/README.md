# The Backrooms

A first-person, infinite Backrooms exploration demo for the Zod terminal.
The world is a 2.5D raycaster rendered at 80x30 cells: half-block truecolor
by default, with a textured ASCII mode on `M`.

The renderer uses 96 colors in six 16-shade families (ceiling, carpet,
two wall sides, trim, and fluorescent light), a 16-step curved fog ramp,
and procedural 64x64 material atlases. Carpets have grain, stains, and
seams; ceilings have acoustic tiles with baked light panels and glow;
walls have wallpaper stripes, chair rails, baseboards, concrete pillars,
stained variants, and dark door frames. Floor and ceiling texels are
sampled with true perspective, and wall texels use the hit fraction and
wall height rather than a flat distance color.

## Run

```sh
go run . examples/demo/backrooms/backrooms.zd
```

The terminal must be at least 80x30. Truecolor is the default; add `--256`
for the xterm-256 color cube, `--ascii` to start in the Unicode ASCII
texture pack, or `--plain-ascii` for the strict 7-bit fallback. The
plain fallback is ASCII-only; `M` does not switch it back to half-blocks.

## Controls

| Input | Action |
| --- | --- |
| `W` `A` `S` `D` | Move and strafe |
| Arrow keys | Move, or turn with left/right |
| Left-button drag | Look |
| `M` | Toggle half-block / Unicode ASCII |
| `Q` or `Esc` | Quit |

Terminals do not report key-up events, so movement continues briefly after a
key press (`KEY_TIMEOUT` in `config.zd`). Mouse look only responds while the
left button is held.

## Headless

```sh
go run . examples/demo/backrooms/backrooms.zd --auto --frames=300
```

Options: `--frames=N`, `--seed=N`, `--ascii`, `--plain-ascii`, `--256`,
and `--dump=/path/to/frame.txt`. Headless mode prints the total time and
the SHA-256 of the last frame.

## Tests

The game's Zod suite runs from Go:

```sh
go test . -run TestBackroomsZodSuite -v
```

or directly:

```sh
go run . examples/demo/backrooms/tests.zd
```

The suite covers chunk determinism and connectivity, door-frame materials,
DDA geometry and stored ray state, collision, mouse parsing, drag
coalescing, 96-color palette ramps and fog tables, both glyph packs,
half-block/ASCII presenter output, color variety, and a golden frame.

## Terminal support

- SGR mouse reporting (`1000`/`1002`/`1006`) is required for drag-look.
  Keyboard controls work without it.
- The default ASCII pack uses Unicode box-drawing and block glyphs. Use
  `--plain-ascii` if the terminal font cannot render them.
- Windows Terminal is required; legacy conhost does not support SGR mouse.
- In tmux, `set -g mouse on` is needed for drag reports. Hold `Shift` while
  dragging to select text instead of looking.
- The game restores raw mode, mouse tracking, the alternate screen, autowrap,
  synchronized output, and the cursor on normal exit, `Q`/`Esc`, runtime
  errors, and `SIGINT`/`SIGTERM`/`SIGHUP`.
