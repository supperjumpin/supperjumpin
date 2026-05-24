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
  "operationId: createInvite",
  "operationId: acceptInvite",
  "operationId: startSeason",
  "operationId: createIdea",
  "operationId: createPlannedStunt",
  "operationId: authorizeEvidenceUpload",
  "operationId: submitEvidence",
  "operationId: submitJudgment",
  "EvidenceUploadAuthorization:",
  "Evidence:",
  "EvidenceSubmission:",
  "Judgment:",
  "PerformedStuntView:",
  "GroupHomeResponse:",
  "Invite:",
  "Season:",
  "Stunt:",
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
    `export interface Invite {\n` +
    `  id: string;\n` +
    `  groupId: string;\n` +
    `  token: string;\n` +
    `  createdBy: string;\n` +
    `  expiresAt: string;\n` +
    `}\n\n` +
    `export interface Season {\n` +
    `  id: string;\n` +
    `  groupId: string;\n` +
    `  commissionerPlayerId: string;\n` +
    `  status: "Active" | "Judging Grace Period" | "Finalized";\n` +
    `}\n\n` +
    `export interface Stunt {\n` +
    `  id: string;\n` +
    `  groupId: string;\n` +
    `  playerId: string;\n` +
    `  seasonId: string | null;\n` +
    `  status: "Idea" | "Planned Stunt" | "Performed Stunt";\n` +
    `  source: string;\n` +
    `  destination: string;\n` +
    `  food: string;\n` +
    `  offSeason: boolean;\n` +
    `}\n\n` +
    `export interface EvidenceUploadAuthorization {\n` +
    `  id: string;\n` +
    `  stuntId: string;\n` +
    `  uploadUrl: string;\n` +
    `  uploadMethod: "PUT";\n` +
    `  uploadHeaders: Record<string, string>;\n` +
    `  mediaObjectKey: string;\n` +
    `  expiresAt: string;\n` +
    `}\n\n` +
    `export interface Evidence {\n` +
    `  id: string;\n` +
    `  stuntId: string;\n` +
    `  caption: string;\n` +
    `  mediaObjectKey: string;\n` +
    `  createdAt: string;\n` +
    `}\n\n` +
    `export interface EvidenceSubmission {\n` +
    `  stunt: Stunt;\n` +
    `  evidence: Evidence;\n` +
    `}\n\n` +
    `export interface Judgment {\n` +
    `  id: string;\n` +
    `  stuntId: string;\n` +
    `  playerId: string;\n` +
    `  difficulty: number;\n` +
    `  transgression: number;\n` +
    `  creativity: number;\n` +
    `  documentation: number;\n` +
    `}\n\n` +
    `export interface PerformedStuntView {\n` +
    `  stunt: Stunt;\n` +
    `  performer: Player;\n` +
    `  evidence: Evidence;\n` +
    `}\n\n` +
    `export interface GroupHomeResponse {\n` +
    `  group: Group;\n` +
    `  membership: GroupMembership;\n` +
    `  activeSeason: Season | null;\n` +
    `  recentStunts: PerformedStuntView[];\n` +
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
