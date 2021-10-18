module notion-blog

go 1.17

// replace github.com/jomei/notionapi => github.com/xzebra/notionapi v1.5.1-0.20211017174639-1af1d92f2914

replace github.com/jomei/notionapi => ../notionapi

require (
	github.com/itzg/go-flagsfiller v1.5.0
	github.com/joho/godotenv v1.4.0
	github.com/jomei/notionapi v1.5.3-0.20211015052055-e3eaeddd589f
	github.com/otiai10/opengraph v1.1.3
)

require (
	github.com/iancoleman/strcase v0.0.0-20191112232945-16388991a334 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	golang.org/x/net v0.0.0-20211015210444-4f30a5c0130f // indirect
)
