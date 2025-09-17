# up-npm

CLI tool written in Go to review and update your NPM dependencies, easy and fast.

![](https://i.imgur.com/8AUJFVb.png)

[![built with Codeium](https://codeium.com/badges/main)](https://codeium.com)


# Features

- 🔍 **Easily identify the update type** for each package, whether it's a patch, minor, or major update.
- 📃 Review the **release notes** for each package to see "what's new" before deciding whether to update.
- 🦘 Selectively **skip** updates for specific packages.
- 🛡️ **Back up** your `package.json` file before updating, ensuring you always have a fallback option if something goes wrong.
- 🔑 Supports .npmrc `_authToken` ([read more here](#npmrc-support))
- 🐞 Warns about versions released too recently


# Installation

```
npm install -g up-npm
```



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



# How to upgrade version

![image](https://github.com/Icaruk/up-npm/assets/10779469/80aa603c-af4e-4f68-8ed3-a754d8b366c1)

```
npm install -g up-npm
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



# .npmrc support

*from https://docs.npmjs.com/cli/v10/configuring-npm/npmrc*

Detects `_authToken` inside .npmrc file.
The four relevant files are:

- (✅ supported) per-project config file (/path/to/my/project/.npmrc)
- (✅ supported) per-user config file (~/.npmrc)
- (❌ unsupported) global config file ($PREFIX/etc/npmrc)
- (❌ unsupported) npm builtin config file (/path/to/npm/npmrc)

This feature allows to fetch private packages.



# Badge

![](https://img.shields.io/badge/up--npm-%20?style=flat&logo=rocket&logoColor=rgb(56%2C%20167%2C%20205)&label=updated%20with&color=rgb(74%2C%20100%2C%20206)&link=https%3A%2F%2Fgithub.com%2FIcaruk%2Fup-npm)

```markdown
![](https://img.shields.io/badge/up--npm-%20?style=flat&logo=rocket&logoColor=rgb(56%2C%20167%2C%20205)&label=updated%20with&color=rgb(74%2C%20100%2C%20206)&link=https%3A%2F%2Fgithub.com%2FIcaruk%2Fup-npm)
```
