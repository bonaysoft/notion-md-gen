# notion-md-gen

`notion-md-gen` allows you to use Notion as a CMS for pages built with any static site generators. You can use it as a cli or even automate your blog repo to update itself with the Github Action.


## Requisites

- Notion database for your articles.
- Notion API secret token.
- A blog by any static site generators.

## Setup
```bash
brew install notion-md-gen
```

## Usage

### CLI

```bash
cd your-blog-dir
notion-md-gen
```

The binary looks for a config file called `notion-md-gen.json` or `notion-md-gen.yaml` in the directory where it is executed. You can see the example config in [notion-md-gen.json](notion-md-gen.json).

### Github Action

To use it as a Github Action, you can follow the example of the repository in [.github/worflows/notion.yml](.github/workflows/notion.yml).

## Compilation

This is only required if you are not going to use the repo as a Github Action. The compilation is simple as Golang installs everything for you.

```bash
go build -o ./bin/main cmd/main/main.go
```

You can compile any form of the app (cli or binary) by compiling the main file in any of the packages in `cmd/`.

