# Timstack

A stack inspired by the amazing Tim.

## Getting Started

Install [air](https://github.com/air-verse/air) for hot reload if not done already.
Install [Task](https://taskfile.dev/installation) for task runner if not done already.

Grab the static tailwind binary from [here](https://github.com/tailwindlabs/tailwindcss/releases) and store it in your path - i.e:

```bash
$ mkdir -p ~/tailwind
$ cd ~/tailwind
$ curl -sLO https://github.com/tailwindlabs/tailwindcss/releases/download/v3.4.6/tailwindcss-macos-arm64
$ chmod +x tailwindcss-macos-x64
$ mv tailwindcss-macos-x64 tailwindcss
$ PATH=$PATH:`pwd`/tools
...
```

Note: _You will want the right binary for your OS/arch_

You should be ready to go - check your build:

```bash
go generate
go build -o mybin
./mybin
```

## Building Tailwind Templates

With the following you can go fast and generate CSS output CSS to play with.

```bash
$ cd templates
$ tailwindcss -i ./tailwind.css -o ./static/css/main.css --watch

Rebuilding...

Done in 264ms.
```

When happy with templates, compile and minify the CSS for production
`tailwindcss -i ./tailwind.css -o ./static/css/main.css --minify`

## Using taskfiles and air to make life easier

Rather than remembering the specifics, you can use the provided task file to make life easier as a starting point.

For dev, which will run with hotreload, via a smart proxy, on change of a go, js, tmpl, css, html file, you can use:
`task live`

For production, which will minify any outputs and be ready to run on a box:
`task prod`
and of course you can run with `task run-prod`.

Refer to the taskfile documentation for overrides of these two modes and run `task` on its own to see all options.

You can provide a .env file and the taskfile will take care of loading that - there's a simple example for the port.

## Database / SQL

### Create a new migration

Note, we use `-seq -digits 4` to use 0001 style format which makes SQLc work better lexographically.

`migrate create -seq -digits 4 -dir ./database/migrations/ -ext sql <your_migration>`

### Local Docker Container

`docker run -e POSTGRES_USER=local -e POSTGRES_PASSWORD=asecurepassword -e POSTGRES_DB=cylm -p 5003:5432 postgres:latest`

### Running a migration

`migrate -path ./database/migrations -database postgres://local:asecurepassword@0.0.0.0:5003/cylm?sslmode=disable up`

### Debugging with pgcli

`pgcli postgres://local:asecurepassword@0.0.0.0:5003/cylm?sslmode=disable`

## TODO

Have fun!
