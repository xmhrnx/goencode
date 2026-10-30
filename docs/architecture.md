# Architecture

GoEncode has three layers. Each layer depends only on the ones below it.

```
main.go              wires the window, the store and the root view
ui/                  everything drawn on screen
  theme/             palettes (dark, light), spacing, radii, text styles
  icons/             stroke icons and the logo
  components/        one reusable widget per file (Select, Slider, FileCard, …)
  views/             one screen part per file, built from components
internal/
  store/             application state and actions (the only thing views change)
  settings/          the recipe: presets, validation, ffmpeg argument building
  catalog/           what GoEncode knows about containers, codecs and encoders
  ffmpeg/            finding, probing and running ffmpeg
  jobs/              a file in the batch and its status
  platform/          preferences, app folders, opening files
```

## Data flow

1. **`ffmpeg`** finds the programs (`Locate`) and reads their capabilities (`Detect`).
   `TestHardware` test-encodes a tiny clip with each GPU encoder, so only working ones are
   offered.
2. **`catalog`** describes containers (which codecs each one accepts) and encoder *profiles*.
   A profile lists the encoder's rate controls (CRF, CQ, bitrate, …) and its options, with help
   text written for that encoder. Encoders without a curated profile get a generic one, built
   from ffmpeg's own option list.
3. **`settings`** holds the recipe applied to the whole batch.
   - `Normalize` fixes a recipe so it stays valid for its container and this ffmpeg.
   - `MakePlan` decides, for each stream of a file, whether to copy, encode, convert, burn in or
     drop it.
   - `Args` turns a file, the recipe and an optional time window (used by previews) into ffmpeg
     arguments.

   All of this is plain Go with no UI, and it is unit-tested in `args_test.go`. The tests in
   `integration_test.go` run real encodes.
4. **`store`** owns the state: the files, the recipe, the queue and the preview.
   - Work that blocks (probing, encoding, extracting frames) runs on goroutines, which send
     their results back with `Host.Post`. As a result, state is only changed on the UI goroutine.
   - The preferences, including the last recipe, are saved to the app's data folder.
5. **`ui/views`** read the store while each frame is built, and change it only by calling store
   actions.

## UI conventions

- **One component per file.** Each file in `ui/components` and `ui/views` exports one component.
  Helpers that belong to a view stay unexported in that view's file. Small helpers shared by
  several settings views live in `views/choices.go`.
- **giode rebuilds the view every frame.**
  - Plain constructors (`SectionCard`, `Pill`) are cheap.
  - Widgets that hold state (selects, sliders, switches, accordions) are looked up with
    `giode.Keyed(scope, key, …)` and need a stable key.
  - If a key includes the encoder (`"v-quality/"+encoder`), switching encoders creates a fresh
    widget, with the right range and choices.
- **Hover help.** `components.InfoLabel` is a label with an ⓘ popover. `OptionControl` uses it
  for every encoder option, so a new option gets its explanation from the `Help` field of its
  `catalog.Option`.
- **Colors.** All colors come from theme tokens (`"theme.primary"`) or from values derived in
  `ui/theme/palette.go`, so dark and light mode follow automatically.
- **Wide and phone layouts.** `views.App` switches layout at 720 dp. Wide windows show the files
  and the settings side by side. Phones switch between them with a segmented control. Option
  rows stack their label above the control when the space is narrower than 360 dp.

## Adding an encoder profile

Register it in the matching `internal/catalog/profiles_*.go` file. Give it a label, a summary,
its rate controls (with their flag, range, default and what "better" means), and its options.
Mark options `Simple` to show them before the "More" section. List any flags the profile sets
itself in `Managed`, so they don't appear twice under "All ffmpeg options".

## Adding a container

Add an entry to `internal/catalog/containers.go`. It needs the muxer, the extension, the codecs
it accepts for each kind (patterns such as `pcm_*` work), the preferred codecs to fall back to,
and any extra output arguments it needs.
