import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";

export function dockerComposeArgsFor(action) {
  if (action === "up") {
    return ["compose", "up", "-d", "postgres"];
  }

  if (action === "down") {
    return ["compose", "stop", "postgres"];
  }

  throw new Error(`Unsupported db lifecycle action: ${action}`);
}

export function runLifecycle(action) {
  return spawnSync("docker", dockerComposeArgsFor(action), { stdio: "inherit" });
}

function main() {
  const action = process.argv[2];
  const result = runLifecycle(action);
  process.exit(result.status ?? 0);
}

if (process.argv[1] === fileURLToPath(import.meta.url)) {
  main();
}
