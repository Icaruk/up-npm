const os = require("os");
const fs = require("fs");
const path = require("path");

const platform = os.platform();
const binPath = path.join(__dirname, "dist");

switch (platform) {
	case "win32":
		fs.copyFileSync(path.join(binPath, "up-npm-v0.0.1-windows-amd64.exe"), path.join(__dirname, "up-npm.exe"));
		break;
	case "darwin":
		fs.copyFileSync(path.join(binPath, "up-npm-v0.0.1-darwin-amd64"), path.join(__dirname, "up-npm"));
		break;
	case "linux":
		fs.copyFileSync(path.join(binPath, "up-npm-v0.0.1-linux-amd64"), path.join(__dirname, "up-npm"));
		break;
	default:
		console.error("Unsupported platform:", platform);
		process.exit(1);
}
