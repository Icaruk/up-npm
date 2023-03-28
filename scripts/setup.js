require('dotenv').config();

const os = require("os");
const fs = require("fs");
const path = require("path");

const platform = os.platform();

const distPath = path.join(__dirname, "../dist");
const scriptsPath = path.join(__dirname, "../scripts");
const binPath = path.join(__dirname, "..");

const appName = process.env.APP_NAME;
const version = process.env.VERSION;

let distFilename = "";
let isWindows = false;


switch (platform) {
	case "win32":
		distFilename = `${appName}-${version}-windows-amd64.exe`;
		isWindows = true;
		break;
	case "darwin":
		distFilename = `${appName}-${version}-darwin-amd64`;
		break;
	case "linux":
		distFilename = `${appName}-${version}-linux-amd64`;
		break;
	default:
		console.error("Unsupported platform:", platform);
		process.exit(1);
}

const suffix = isWindows ? ".exe" : "";
const binFilename = `${appName}${suffix}`;

const source = path.join(distPath, distFilename);
const destination = path.join(binPath, binFilename);

if (!fs.existsSync(source)) {
	console.error(`${source} does not exist`);
	process.exit(1);
};


const emptyBinPath = path.join(binPath, appName);
if (fs.existsSync(emptyBinPath)) fs.rmSync(emptyBinPath)
fs.copyFileSync(source, destination);

if (!isWindows) {
	// "chmod +x up-npm"
	fs.chmodSync(destination, 0o100);
}

// Cleanup
fs.rmdirSync(distPath, { recursive: true });
fs.rmdirSync(scriptsPath, { recursive: true });
