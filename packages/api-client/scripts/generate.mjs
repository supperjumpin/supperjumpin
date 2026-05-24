import { readFile, writeFile } from "node:fs/promises";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const root = resolve(dirname(fileURLToPath(import.meta.url)), "../../..");
const contractPath = resolve(root, "apps/api/openapi.yaml");
const outputPath = resolve(root, "packages/api-client/src/generated.d.ts");

const contract = await readFile(contractPath, "utf8");
if (!contract.includes("operationId: getMe") || !contract.includes("MeResponse:")) {
  throw new Error("OpenAPI contract no longer exposes getMe MeResponse");
}

await writeFile(
  outputPath,
  `// Generated from apps/api/openapi.yaml by npm run generate:api-client.\n` +
    `export interface Account {\n` +
    `  id: string;\n` +
    `  email: string;\n` +
    `}\n\n` +
    `export interface Player {\n` +
    `  id: string;\n` +
    `  displayName: string;\n` +
    `}\n\n` +
    `export interface MeResponse {\n` +
    `  account: Account;\n` +
    `  player: Player;\n` +
    `}\n`,
);
