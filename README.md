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

- A tag is stacked when it has two or more attributes and either
  - it has at least `-attrs` attributes (default 3; 0 turns this off), or
  - its opening tag, placed at the indentation of its line, would end beyond `-width`
    columns (default 100; a tab counts `-tab` columns, default 4). Siblings earlier on the
    same line don't count.
- The children of a stacked tag go on their own lines (`>{ x }</a>` becomes `>`, `{ x }`,
  `</a>`), except in `pre` and `textarea`.
- Tags that are already stacked are left alone. `<script>` and `<style>` are not touched.
- With `-sort` (off by default) the attributes of every tag are put in this order: `id`,
  `class`, `name`/`type`/`slot`, everything else, `x-*`/`@*`/`:*` (Alpine), `hx-*` (htmx),
  `data-*`/`aria-*`. Spreads, conditional attributes and attributes with a computed name stay
  where they are and nothing moves across them. Sorting changes the order of the attributes
  in the HTML (and so in the generated Go code); nothing else does.
- Attribute values, `class` strings and everything else are not touched.
- A file with nothing to change is returned byte for byte.
- Imports and `<script>`/`<style>` formatting are `templ fmt`'s job: run it afterwards.

## Safe to run on a project

Apart from `-sort`, templwrap changes layout only: the Go code `templ generate` produces
stays the same (the positions in its error messages aside). The tests check this for tags
with inline text, `pre`, `textarea`, spreads, conditional attributes, components with
children and control flow. To check a project yourself, run
`templ generate -stdout -f file.templ` before and after and diff the two.

## Use

As a tool dependency of a project (pins the version with the project's templ):

```bash
go get -tool github.com/rnkch/templwrap@latest
go tool templwrap < views/page.templ | go tool templ fmt -stdin-filepath views/page.templ
```

Or install it once:

```bash
go install github.com/rnkch/templwrap@latest
```

Neovim with [conform.nvim](https://github.com/stevearc/conform.nvim):

```lua
formatters_by_ft = { templ = { "templwrap", "templ" } },
formatters = {
  templwrap = { command = "templwrap", args = {}, stdin = true },
},
```

## Which files

Only `.templ` files: templwrap reads them with templ's own parser. Other Go template syntaxes
(`html/template` and the like) are not touched.

## Compatibility

templ's parser package has no stability promise (it is v0.3.x). templwrap needs templ
v0.3.1070 or newer (`Element.OpenTagRange`) and is tested against that version and the
latest one on every push and weekly. Built as a `go tool`, it is compiled against the
templ version of your project.

## Licence

MIT
