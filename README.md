# notion-md-gen

`notion-md-gen` allows you to use Notion as a CMS for pages built with any static site generators. You can use it as a cli or even automate your blog repo to update itself with the Github Action.

## Requisites

- Notion database for your articles.
- Notion API secret token.
- A blog by any static site generators.

## Setup(not ready)
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

## Contributing
See [CONTRIBUTING](CONTRIBUTING.md) for details on submitting patches and the contribution workflow.

## Special thanks
- [xzebra](https://github.com/xzebra)

I based this code on [https://github.com/xzebra/notion-blog](https://github.com/xzebra/notion-blog/commit/7982bcf0445cfdca1efd250d1f76d9fee07fc975)

## License
notion-md-gen is under the MIT license. See the [LICENSE](/LICENSE) file for details.