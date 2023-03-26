const os = require("os");
const fs = require("fs");
const path = require("path");

const platform = os.platform();

const distPath = path.join(__dirname, "../dist");
const scriptsPath = path.join(__dirname, "../scripts");
const binPath = path.join(__dirname);

const appName = "up-npm";
const version = "0.0.1";

let distFilename = "";
let suffix = "";


switch (platform) {
	case "win32":
		distFilename = `${appName}-${version}-windows-amd64.exe`;
		suffix = ".exe";
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

const binFilename = `${appName}${suffix}`;

const source = path.join(distPath, distFilename);
const destination = path.join(binPath, binFilename);

if (!fs.existsSync(source)) {
	console.error(`${source} does not exist`);
	process.exit(1);
};


if (fs.existsSync(destination)) fs.rmSync(destination)
fs.copyFileSync(source, destination);


// Cleanup
// fs.rmdirSync(distPath, { recursive: true });
// fs.rmdirSync(scriptsPath, { recursive: true });
