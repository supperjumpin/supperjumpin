import { readFile, writeFile } from "node:fs/promises";
import { dirname, resolve } from "node:path";
import { fileURLToPath } from "node:url";

const root = resolve(dirname(fileURLToPath(import.meta.url)), "../../..");
const contractPath = resolve(root, "apps/api/openapi.yaml");
const outputPath = resolve(root, "packages/api-client/src/generated.d.ts");

const contract = await readFile(contractPath, "utf8");
for (const required of [
  "operationId: getMe",
  "operationId: createGroup",
  "operationId: listGroups",
  "operationId: getGroupHome",
  "GroupHomeResponse:",
]) {
  if (!contract.includes(required)) {
    throw new Error(`OpenAPI contract no longer exposes ${required}`);
  }
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
    `}\n\n` +
    `export interface Group {\n` +
    `  id: string;\n` +
    `  name: string;\n` +
    `}\n\n` +
    `export interface GroupMembership {\n` +
    `  groupId: string;\n` +
    `  playerId: string;\n` +
    `  role: "Group Admin" | "Player";\n` +
    `}\n\n` +
    `export interface GroupHomeResponse {\n` +
    `  group: Group;\n` +
    `  membership: GroupMembership;\n` +
    `  activeSeason: null;\n` +
    `  recentStunts: unknown[];\n` +
    `  standings: unknown[];\n` +
    `}\n\n` +
    `export interface GroupMembershipSummary {\n` +
    `  group: Group;\n` +
    `  membership: GroupMembership;\n` +
    `}\n\n` +
    `export interface ListGroupsResponse {\n` +
    `  memberships: GroupMembershipSummary[];\n` +
    `}\n`,
);
