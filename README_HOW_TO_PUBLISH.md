# How to Publish a New Version

- Verify that `.version` is set to `0.0.0`.
  - This will be overridden by `scripts/setup.js` during installation.
- Verify that `bin/up-npm` is empty or contains only placeholder text like `"i'm empty"`.
  - This is required because `package.json` has a a field called `bin.up-npm` that must detect a file to create the symlink.
  - Windows `.exe` is omitted because npm handles it automatically.
- Run `npm version patch/minor/major`.
- Run `task build`.
  - This will place the binaries with checksums in the `dist` folder.
 Run `npm publish`.
- Create a new release at https://github.com/Icaruk/up-npm/releases
  - Upload the `dist` files as release assets.
  - These will be downloaded automatically by users during installation.
