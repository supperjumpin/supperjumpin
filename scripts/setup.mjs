export function parseSemver(versionString) {
  const cleaned = versionString.replace(/^v/, "");
  const parts = cleaned.split(".").map(Number);
  if (parts.length !== 3 || parts.some((n) => Number.isNaN(n))) {
    throw new Error(`Invalid semver string: ${versionString}`);
  }
  return { major: parts[0], minor: parts[1], patch: parts[2] };
}

function normalizeVersionString(str) {
  const parts = str.split(".");
  while (parts.length < 3) {
    parts.push("0");
  }
  return parts.join(".");
}

export function checkVersionRange(version, range) {
  const v = parseSemver(version);
  const minMatch = range.match(/>=([\d.]+)/);
  const maxMatch = range.match(/<([\d.]+)/);

  if (minMatch) {
    const min = parseSemver(normalizeVersionString(minMatch[1]));
    if (v.major < min.major || (v.major === min.major && v.minor < min.minor) || (v.major === min.major && v.minor === min.minor && v.patch < min.patch)) {
      return false;
    }
  }

  if (maxMatch) {
    const max = parseSemver(normalizeVersionString(maxMatch[1]));
    if (v.major > max.major || (v.major === max.major && v.minor > max.minor) || (v.major === max.major && v.minor === max.minor && v.patch >= max.patch)) {
      return false;
    }
  }

  return true;
}

export function formatFailureMessage(tool, current, requirement) {
  return `${tool} version ${current} does not satisfy requirement ${requirement}`;
}

import { spawnSync } from "node:child_process";

function capture(command, args) {
  return spawnSync(command, args, { encoding: "utf8" });
}

function checkTool(name, command, args, extractVersion, requirement, installHint) {
  const result = capture(command, args);
  if (result.status !== 0) {
    return { ok: false, message: `${name} is required but not found. ${installHint}` };
  }
  const version = extractVersion(result.stdout);
  if (!checkVersionRange(version, requirement)) {
    return { ok: false, message: formatFailureMessage(name, version, requirement) + `. ${installHint}` };
  }
  return { ok: true, message: `${name} ${version} ✅` };
}

function extractNodeVersion(stdout) {
  return stdout.trim().replace(/^v/, "");
}

function extractGoVersion(stdout) {
  // "go version go1.26.3 linux/amd64"
  const match = stdout.trim().match(/go([\d.]+)/);
  if (!match) throw new Error(`Could not parse Go version from: ${stdout}`);
  return match[1];
}

function extractDockerComposeVersion(stdout) {
  // "Docker Compose version v2.32.4" or similar
  const match = stdout.trim().match(/v?([\d.]+)/);
  if (!match) throw new Error(`Could not parse Docker Compose version from: ${stdout}`);
  return match[1];
}

if (import.meta.url === `file://${process.argv[1]}`) {
  const checks = [
    checkTool(
      "Node",
      "node",
      ["--version"],
      extractNodeVersion,
      ">=24.16.0 <25",
      "Install Node 24.16.0+ via your version manager (e.g. nvm install 24.16.0) or from https://nodejs.org/"
    ),
    checkTool(
      "Go",
      "go",
      ["version"],
      extractGoVersion,
      ">=1.26.3 <1.27",
      "Install Go 1.26.3+ from https://go.dev/dl/ or via your package manager."
    ),
    checkTool(
      "Docker Compose",
      "docker",
      ["compose", "version"],
      extractDockerComposeVersion,
      ">=2.0.0",
      "Install Docker Desktop/Engine with Compose from https://docs.docker.com/compose/install/"
    ),
  ];

  let allOk = true;
  for (const check of checks) {
    console.log(check.message);
    if (!check.ok) allOk = false;
  }

  if (!allOk) {
    console.error("\nSetup aborted due to missing or outdated host tools.");
    process.exit(1);
  }

  console.log("\nRunning npm ci for reproducible dependency installation...");
  const npmCi = capture("npm", ["ci"]);
  if (npmCi.status !== 0) {
    console.error("npm ci failed:");
    console.error(npmCi.stderr || "");
    process.exit(1);
  }
  console.log("npm ci completed ✅");
  console.log("\nSetup complete. You can now run:");
  console.log("  npm run api:dev       # Start the API dev server");
  console.log("  npm run demo:api      # Start API with Docker Compose Postgres");
  console.log("  npm run demo:mobile   # Start the Expo mobile app");
}
