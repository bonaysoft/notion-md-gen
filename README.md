# notion-md-gen

[![](https://github.com/bonaysoft/notion-md-gen/workflows/build/badge.svg)](https://github.com/bonaysoft/notion-md-gen/actions?query=workflow%3Abuild)
[![codecov](https://codecov.io/gh/bonaysoft/notion-md-gen/branch/master/graph/badge.svg?token=XHG00YHOJF)](https://codecov.io/gh/bonaysoft/notion-md-gen)
[![](https://img.shields.io/github/v/release/bonaysoft/notion-md-gen.svg)](https://github.com/bonaysoft/notion-md-gen/releases)
[![](https://img.shields.io/github/license/bonaysoft/notion-md-gen.svg)](https://github.com/bonaysoft/notion-md-gen/blob/master/LICENSE)

`notion-md-gen` allows you to use Notion as a CMS for pages built with any static site generators. You can use it as a
cli or even automate your blog repo to update itself with the Github Action.

## Requisites

- Notion database for your articles.
- Notion API secret token.
- A blog by any static site generators.

## Setup

### install.sh

```bash
curl -sSf https://raw.githubusercontent.com/bonaysoft/notion-md-gen/master/install.sh | sh
```

### webi (not ready)

```bash
curl https://webinstall.dev/notion-md-gen | bash
```

### brew (not ready)

```bash
brew install notion-md-gen
```

## Usage

### CLI

```bash
cd your-blog-dir
notion-md-gen init
notion-md-gen
```

### Github Action

> The installation command tool is helpful for local debugging. If you do not want to debug locally, you can also copy the configuration file to your project and run it directly through GitHubAction. You can see the example config in [example/notion-md-gen.yaml](example/notion-md-gen.yaml).

To use it as a Github Action, you can follow the example of the repository
in [.github/worflows/notion.yml](.github/workflows/notion.yml).

## Contributing

See [CONTRIBUTING](CONTRIBUTING.md) for details on submitting patches and the contribution workflow.

## Special thanks

- [xzebra](https://github.com/xzebra)

I based this code
on [https://github.com/xzebra/notion-blog](https://github.com/xzebra/notion-blog/commit/7982bcf0445cfdca1efd250d1f76d9fee07fc975)

## License

notion-md-gen is under the MIT license. See the [LICENSE](/LICENSE) file for details.
