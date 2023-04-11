
const os = require("os");
const fs = require("fs");
const path = require("path");

const platform = os.platform();

const distPath = path.join(__dirname, "../dist2");

const appName = "up-npm";



(async() => {
	
	try {
		
		console.log( "Fetching latest version..." )
		const responseLatestVersion = await fetch("https://api.github.com/repos/Icaruk/up-npm/releases/latest");
		const json = await responseLatestVersion.json();
		
		const latestVersion = json.tag_name;
		console.log( "OK. Latest version found: ", latestVersion );
		
		let binaryName = "";
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
		
		binaryName = `${appName}-${latestVersion}-${platformName}-${archName}`;
		if (isWindows) binaryName += ".exe";
		
		const assets = json.assets ?? [];
		
		// https://regex101.com/r/rfH9Fe/1
		const regex = new RegExp(`(${platformName}-${archName})(\.exe)?$`, "gi")
		const foundAsset = assets.find( _asset => regex.test(_asset.name) );
		
		if (!foundAsset) {
			console.log(`Can't find binary asset for ${binaryName}`);
			process.exit(1);
		};
		
		const size = foundAsset.size ?? 0;
		const sizeMB = +(size / (1024 * 1024)).toFixed(2);
		console.log( `Found binary '${binaryName}' (${sizeMB} MB)` );
		
		console.log( `Downloading ${binaryName}...` );
		const downloadUrl = foundAsset.browser_download_url;
		const responseDownload = await fetch(downloadUrl);
		console.log( "OK" )
		
		const data = await responseDownload.arrayBuffer();
		const bufferView = new Uint8Array(data);
		
		if (!fs.existsSync(distPath)) fs.mkdirSync(distPath);
		
		fs.writeFileSync(path.join(distPath, binaryName), bufferView);
		
		if (!isWindows) {
			// "chmod +rwx up-npm"
			fs.chmodSync(destination, 0o700);
		}
		
	} catch (err) {
		console.error(err)
	};
	
})();



// Cleanup
// fs.rmdirSync(distPath, { recursive: true });
// fs.rmdirSync(scriptsPath, { recursive: true });
