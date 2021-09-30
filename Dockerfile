FROM golang:1.17-alpine

RUN apk add --no-cache git

LABEL "com.github.actions.name"="notion-blog"
LABEL "com.github.actions.description"="Notion blog articles database to hugo-style markdown."
LABEL "repository"="https://github.com/xzebra/notion-blog"
LABEL "maintainer"="xzebra <zebrv.apps@gmail.com>"

WORKDIR /usr/src/app

# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

# Build the Go app
RUN go build -o ./bin/cli cmd/cli/main.go

ENTRYPOINT ["/usr/src/app/bin/cli"]
