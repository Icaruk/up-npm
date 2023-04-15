# up-npm

CLI tool written in Go to review and update your NPM dependencies, easy and fast.

![](https://i.imgur.com/8AUJFVb.png)

# Features

- 🔍 **Easily identify the update type** for each package, whether it's a patch, minor, or major update.
- 📃 Review the **release notes** for each package to see "what's new" before deciding whether to update.
- 🦘 Selectively **skip** updates for specific packages.
- 🛡️ **Back up** your `package.json` file before updating, ensuring you always have a fallback option if something goes wrong.


# Usage

Go where your `package.json` is located and run:

```bash
up-npm
```

or 

```bash
up-npm [flags]
```

| Flag              	| Description                                   |
|---------------------	|-----------------------------------------------|
| -d, --dev           	| Include dev dependencies                      |
| -f, --filter `string` | Filter dependencies by package name           |
| -h, --help          	| Display help information for up-npm           |
| -v, --version       	| Display the version number for up-npm         |



# Examples

```bash
# Update dependencies
npm-up

# Including dev dependencies
npm-up --dev
npm-up -d

# Update only packages containing "lint"
npm-up -filter lint
npm-up -f lint

```



# Build yourself

- Prerequisites:
  - [Go 1.20+](https://go.dev/doc/install)
  - [Node 18+](https://nodejs.org/en/download)
  - [Taskfile](https://taskfile.dev)
- Then run:
	```bash
	task buid
	```
- Binaries will be created in `/dist` folder.
