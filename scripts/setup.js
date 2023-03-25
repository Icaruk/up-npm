const os = require("os");
const fs = require("fs");
const path = require("path");

const platform = os.platform();
const binPath = path.join(__dirname, "../dist");
const basePath = path.join(__dirname, "..");

switch (platform) {
	case "win32":
		fs.copyFileSync(path.join(binPath, "up-npm-0.0.1-windows-amd64.exe"), path.join(basePath, "up-npm.exe"));
		break;
	case "darwin":
		fs.copyFileSync(path.join(binPath, "up-npm-0.0.1-darwin-amd64"), path.join(basePath, "up-npm"));
		break;
	case "linux":
		fs.copyFileSync(path.join(binPath, "up-npm-0.0.1-linux-amd64"), path.join(basePath, "up-npm"));
		break;
	default:
		console.error("Unsupported platform:", platform);
		process.exit(1);
}
