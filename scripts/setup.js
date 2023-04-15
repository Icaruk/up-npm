
const os = require("os");
const fs = require("fs");
const path = require("path");

const platform = os.platform();

const currentVersionPath = ".version";
const binPath = path.join(__dirname, "../bin");

const appName = "up-npm";



/**
 * Compare two semantic version strings and return 1 if the first version is
 * greater, -1 if the second version is greater, or 0 if they are equal.
 *
 * @param {string} versionA The first version string to compare.
 * @param {string} versionB The second version string to compare.
 * @returns {-1 | 0 | 1} 1 if versionA > versionB, -1 if versionB > versionA, or 0 if they are equal.
 */
function compareSemver(versionA, versionB) {
	const arr1 = versionA.split(".").map(n => parseInt(n));
	const arr2 = versionB.split(".").map(n => parseInt(n));

	for (let i = 0; i < arr1.length; i++) {
		if (arr1[i] > arr2[i]) {
			return 1;
		} else if (arr2[i] > arr1[i]) {
			return -1;
		}
	}

	return 0;
};

async function downloadBinary() {
	try {
		// Get current version
		let currentVersion = "0.0.0";
		
		if (fs.existsSync(currentVersionPath)) {
			currentVersion = fs.readFileSync(currentVersionPath).toString();
			console.log( `Current version: ${currentVersion}` )
		};
		
		console.log( "Fetching latest version..." )
		const responseLatestVersion = await fetch("https://api.github.com/repos/Icaruk/up-npm/releases/latest");
		const json = await responseLatestVersion.json();
		
		const latestVersion = json.tag_name;
		console.log( "OK. Latest version found: ", latestVersion );
		
		
		const semverCompareResult = compareSemver(latestVersion, currentVersion);
		// 1 must upgrade
		// 0 up to date
		// -1 (impossible) local version is greater than remote one
		
		if (semverCompareResult === 0) {
			console.log( "You are already up-to-date!" );
			process.exit(1);
		};
		
		let remoteBinaryName = "";
		let localBinaryName = appName;
		let isWindows = false;
		let platformName = "";
		let archName = "amd64";
		
		
		switch (platform) {
			case "win32":
				platformName = `windows`;
				isWindows = true;
				break;
			case "darwin":
				platformName = `darwin`;
				break;
			case "linux":
				platformName = `linux`;
				break;
			default:
				console.error("Unsupported platform:", platform);
				process.exit(1);
		};
		
		remoteBinaryName = `${appName}-${latestVersion}-${platformName}-${archName}`;
		
		if (isWindows) {
			remoteBinaryName += ".exe";
			localBinaryName += ".exe";
		};
		
		const assets = json.assets ?? [];
		
		// https://regex101.com/r/rfH9Fe/1
		const regex = new RegExp(`(${platformName}-${archName})(\.exe)?$`, "gi")
		const foundAsset = assets.find( _asset => regex.test(_asset.name) );
		
		if (!foundAsset) {
			console.log(`Can't find binary asset for ${remoteBinaryName}`);
			process.exit(1);
		};
		
		const size = foundAsset.size ?? 0;
		const sizeMB = +(size / (1024 * 1024)).toFixed(2);
		console.log( `Found binary '${remoteBinaryName}' (${sizeMB} MB)` );
		
		console.log( `Downloading ${remoteBinaryName}...` );
		const downloadUrl = foundAsset.browser_download_url;
		const responseDownload = await fetch(downloadUrl);
		console.log( "OK" )
		
		const data = await responseDownload.arrayBuffer();
		const bufferView = new Uint8Array(data);
		
		// Delete all contents of folder or create if it doesn't exists
		if (fs.existsSync(binPath)) {
			for (const _file of fs.readdirSync(binPath)) {
				fs.unlinkSync(path.join(binPath, _file));
			}
		} else {
			fs.mkdirSync(binPath);
		}
		
		fs.writeFileSync(path.join(binPath, localBinaryName), bufferView);
		
		if (!isWindows) {
			// "chmod +rwx up-npm"
			fs.chmodSync(destination, 0o700);
		}
		
		fs.writeFileSync(currentVersionPath, latestVersion);
		
	} catch (err) {
		console.error(err)
	};
};


downloadBinary();