# up-npm

CLI tool written in Go to review and update your NPM dependencies, easy and fast.

![](https://i.imgur.com/8AUJFVb.png)

[![built with Codeium](https://codeium.com/badges/main)](https://codeium.com)

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

| Flag              	| Description                                   				|
|---------------------	|-------------------------------------------------------------  |
| -h, --help          	| Display help information for up-npm.           				|
| --allow-downgrade     | Allows downgrading a if latest version is older than current.	|
| --file `string`     	| Default `package.json`.										|
| -f, --filter `string` | Filter dependencies by package name           				|
| --no-dev           	| Exclude dev dependencies. Default `false`.   					|
| --update-patches     	| Update patch versions automatically. Default `false`.  		|
| -v, --version       	| Display the version number for up-npm.         				|



# Examples

```bash
# Update dependencies
npm-up

# Excluding dev dependencies
npm-up --no-dev

# Update only packages containing "lint"
npm-up --filter lint
npm-up -f lint

# Update some specific .json
npm-up --file my-project/package.json

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
