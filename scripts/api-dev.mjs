import { spawn } from "node:child_process";

const databaseURL = process.env.DATABASE_URL ?? "postgres://postgres:postgres@localhost:5432/supperjumpin?sslmode=disable";
const devToken = process.env.SUPPERJUMPIN_DEV_AUTH_TOKEN ?? "dev-token";

console.log("Starting Supperjumpin API on http://localhost:8080");
console.log(`Demo bearer token: ${devToken}`);

const api = spawn("go", ["run", "./apps/api/cmd/api"], {
  env: {
    ...process.env,
    DATABASE_URL: databaseURL,
    SUPPERJUMPIN_DEV_AUTH_TOKEN: devToken,
  },
  stdio: "inherit",
});

api.on("exit", (code) => {
  process.exit(code ?? 0);
});
