# up-npm

CLI tool written in Go to review and update your NPM dependencies, easy and fast.

![](https://i.imgur.com/8AUJFVb.png)


# Features

- Update each package one by one
- Check version update before updating it: patch, minor or major
- Review what's new on each package before updating it


# Usage

Go where your `package.json` is located and run:

```bash
up-npm [flags]
```

Flags:

- -h, --help      help
- -v, --version   version
- -d, --dev       Update dev dependencies



# Examples

- Update dependencies:
`npm-up`

- Update dependencies including _devDependencies:_
`npm-up --dev` or `npm-up -d`



# Build yourself

- Prerequisites:
  - [Go 1.20](https://go.dev/doc/install)
  - [Node 18](https://nodejs.org/en/download)
  - [Taskfile](https://taskfile.dev)
- Then run:
	```bash
	task buid
	```
- Binaries will be created in `/dist` folder.
