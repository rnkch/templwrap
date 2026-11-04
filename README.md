# templwrap

Stacks the attributes of long [templ](https://templ.guide) elements, one attribute per line
with `>` on its own line:

```templ
<wa-button href="/add" { pageLink("/add")... } size="s" variant="brand">
```

becomes

```templ
<wa-button
	href="/add"
	{ pageLink("/add")... }
	size="s"
	variant="brand"
>
```

`templ fmt` keeps a tag stacked if you broke it up by hand but never breaks one itself
(see [a-h/templ#563](https://github.com/a-h/templ/discussions/563)). templwrap does the
breaking, using templ's own parser (`parser/v2`), and leaves the rest to `templ fmt`.

## Rules

- An element is stacked when it has two or more attributes and its opening tag, placed at
  the indentation of its line, would end beyond `-width` columns (default 100, a tab counts
  `-tab` columns, default 4). Siblings earlier on the same line don't count.
- Tags that are already stacked, and tags that already span lines, are left alone.
- Attribute values, `class` strings, children and everything else are not touched.
- A file with nothing to stack is returned byte for byte.
- Imports and `<script>`/`<style>` formatting are `templ fmt`'s job: run it afterwards.

## Use

As a tool dependency of a project (pins the version with the project's templ):

```bash
go get -tool github.com/rnkch/templwrap@latest
go tool templwrap -width 100 < views/page.templ | go tool templ fmt -stdin-filepath views/page.templ
```

Or install it once:

```bash
go install github.com/rnkch/templwrap@latest
```

Neovim with [conform.nvim](https://github.com/stevearc/conform.nvim):

```lua
formatters_by_ft = { templ = { "templwrap", "templ" } },
formatters = {
  templwrap = { command = "templwrap", args = { "-width", "100" }, stdin = true },
},
```

## Compatibility

templ's parser package has no stability promise (it is v0.3.x). templwrap needs templ
v0.3.1070 or newer (`Element.OpenTagRange`) and is tested against that version and the
latest one on every push and weekly. Built as a `go tool`, it is compiled against the
templ version of your project.

## Licence

MIT
