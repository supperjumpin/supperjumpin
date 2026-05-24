// Generated from apps/api/openapi.yaml by npm run generate:api-client.
export interface Account {
  id: string;
  email: string;
}

export interface Player {
  id: string;
  displayName: string;
}

export interface MeResponse {
  account: Account;
  player: Player;
}

export interface Group {
  id: string;
  name: string;
}

export interface GroupMembership {
  groupId: string;
  playerId: string;
  role: "Group Admin" | "Player";
}

export interface Invite {
  id: string;
  groupId: string;
  token: string;
  createdBy: string;
  expiresAt: string;
}

export interface Season {
  id: string;
  groupId: string;
  commissionerPlayerId: string;
  status: "Active" | "Judging Grace Period" | "Finalized";
}

export interface Stunt {
  id: string;
  groupId: string;
  playerId: string;
  seasonId: string | null;
  status: "Idea" | "Planned Stunt" | "Performed Stunt";
  source: string;
  destination: string;
  food: string;
  offSeason: boolean;
}

export interface EvidenceUploadAuthorization {
  id: string;
  stuntId: string;
  uploadUrl: string;
  uploadMethod: "PUT";
  uploadHeaders: Record<string, string>;
  mediaObjectKey: string;
  expiresAt: string;
}

export interface Evidence {
  id: string;
  stuntId: string;
  caption: string;
  mediaObjectKey: string;
  createdAt: string;
}

export interface EvidenceSubmission {
  stunt: Stunt;
  evidence: Evidence;
}

export interface PerformedStuntView {
  stunt: Stunt;
  performer: Player;
  evidence: Evidence;
}

export interface GroupHomeResponse {
  group: Group;
  membership: GroupMembership;
  activeSeason: Season | null;
  recentStunts: PerformedStuntView[];
  standings: unknown[];
}

export interface GroupMembershipSummary {
  group: Group;
  membership: GroupMembership;
}

export interface ListGroupsResponse {
  memberships: GroupMembershipSummary[];
}
