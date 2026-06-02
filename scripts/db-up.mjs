import { spawnSync } from "node:child_process";

const result = spawnSync("docker", ["compose", "up", "-d", "postgres"], { stdio: "inherit" });

process.exit(result.status ?? 0);
