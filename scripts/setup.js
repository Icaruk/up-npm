const os = require("os");
const fs = require("fs");
const path = require("path");
const https = require("https");

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
	const arr1 = versionA.split(".").map((n) => parseInt(n));
	const arr2 = versionB.split(".").map((n) => parseInt(n));

	for (let i = 0; i < arr1.length; i++) {
		if (arr1[i] > arr2[i]) {
			return 1;
		}
		if (arr2[i] > arr1[i]) {
			return -1;
		}
	}

	return 0;
}

function getCurrentVersion() {
	let currentVersion = "0.0.0";

	if (fs.existsSync(currentVersionPath)) {
		currentVersion = fs.readFileSync(currentVersionPath).toString();
	}

	return currentVersion;
}

/**
 * @returns {Promise<{latestVersion: string, assets: any[]}>}
 */
async function fetchLatestVersion() {
	try {
		const options = {
			hostname: "api.github.com",
			path: "/repos/Icaruk/up-npm/releases/latest",
			headers: {
				"User-Agent": "Node.js",
			},
		};

		return new Promise((resolve, reject) => {
			const req = https.get(options, (res) => {
				let body = "";
				res.on("data", (chunk) => {
					body += chunk;
				});
				res.on("end", () => {
					const json = JSON.parse(body);

					resolve({
						latestVersion: json.tag_name,
						assets: json.assets ?? [],
					});
				});
			});

			req.on("error", (err) => reject(err));
		});
	} catch (err) {
		console.error(err);
		return null;
	}
}

/**
 * @returns {Promise<{isAvailableUpdate: boolean, latestVersion: string, assets: any[]}>}
 */
async function checkIsAvailableUpdate() {
	const currentVersion = getCurrentVersion();
	console.log(`Current version: ${currentVersion}`);

	console.log("Fetching latest version...");
	const { latestVersion, assets } = await fetchLatestVersion();

	if (!latestVersion) {
		console.log("No latest version found");
		return {
			isAvailableUpdate: false,
			latestVersion: null,
		};
	}
	console.log("OK. Latest version found: ", latestVersion);

	const semverCompareResult = compareSemver(latestVersion, currentVersion);
	// 1 must upgrade
	// 0 up to date
	// -1 (impossible) local version is greater than remote one

	if (semverCompareResult === 0) {
		return {
			isAvailableUpdate: false,
			latestVersion: null,
			assets: [],
		};
	}

	return {
		isAvailableUpdate: true,
		latestVersion: latestVersion,
		assets: assets,
	};
}

/**
 * @param {string} url The URL of the binary file to download.
 * @returns {Promise<Uint8Array | null>} A promise that resolves with the binary file data as a Uint8Array.
 */
async function downloadBinary(url, fileName) {
	// https://github.com/Icaruk/up-npm/releases/download/3.1.0/up-npm-3.1.0-windows-amd64.exe

	const urlObj = new URL(url);
	const hostname = urlObj.hostname;
	const path = urlObj.pathname;

	try {
		return new Promise((resolve, reject) => {
			https
				.get(url, (response) => {
					if (response.statusCode === 302) {
						downloadBinary(response.headers.location, fileName);
					}

					const fileStream = fs.createWriteStream(fileName);

					fileStream.on("error", (error) => {
						console.error(`Error saving the file: ${error.message}`);
						fs.unlink(fileName, () => {});
					});

					response.pipe(fileStream);

					fileStream.on("finish", () => {
						fileStream.close();
						resolve(true);
					});
				})
				.on("error", (error) => {
					console.error(`Error downloading file: ${error.message}`);
					reject(error);
				});
		});
	} catch (err) {
		console.error(err);
		return null;
	}
}

async function init() {
	try {
		const { isAvailableUpdate, latestVersion, assets } =
			await checkIsAvailableUpdate();

		if (!isAvailableUpdate) {
			console.log("You are already up-to-date!");
			process.exit(1);
		}

		let remoteBinaryName = "";
		let localBinaryName = appName;
		let isWindows = false;
		let platformName = "";
		const archName = "amd64";

		switch (platform) {
			case "win32":
				platformName = "windows";
				isWindows = true;
				break;
			case "darwin":
				platformName = "darwin";
				break;
			case "linux":
				platformName = "linux";
				break;
			default:
				console.error("Unsupported platform:", platform);
				process.exit(1);
		}

		remoteBinaryName = `${appName}-${latestVersion}-${platformName}-${archName}`;

		if (isWindows) {
			remoteBinaryName += ".exe";
			localBinaryName += ".exe";
		}

		// https://regex101.com/r/rfH9Fe/1
		const regex = new RegExp(`(${platformName}-${archName})(\.exe)?$`, "gi");
		const foundAsset = assets.find((_asset) => regex.test(_asset.name));

		if (!foundAsset) {
			console.log(`Can't find binary asset for ${remoteBinaryName}`);
			process.exit(1);
		}

		const size = foundAsset.size ?? 0;
		const sizeMB = +(size / (1024 * 1024)).toFixed(2);
		console.log(`Found binary '${remoteBinaryName}' (${sizeMB} MB)`);

		const destination = path.join(binPath, localBinaryName);

		// Delete all contents of folder or create if it doesn't exists
		if (fs.existsSync(binPath)) {
			for (const _file of fs.readdirSync(binPath)) {
				fs.unlinkSync(path.join(binPath, _file));
			}
		} else {
			fs.mkdirSync(binPath);
		}

		// Download file
		console.log(`Downloading '${remoteBinaryName}'...`);
		const downloadUrl = foundAsset.browser_download_url;
		const bufferView = await downloadBinary(downloadUrl, destination);
		console.log("OK");

		fs.writeFileSync(currentVersionPath, latestVersion);

		if (!isWindows) {
			// "chmod +rwx up-npm"
			fs.chmodSync(destination, 0o700);

			// "chmod +rw .version"
			fs.chmodSync(currentVersionPath, 0o600);
		}
	} catch (err) {
		console.error(err);
	}
}

init();
