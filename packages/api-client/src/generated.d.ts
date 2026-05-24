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
  status: "Idea" | "Planned Stunt";
  source: string;
  destination: string;
  food: string;
  offSeason: boolean;
}

export interface GroupHomeResponse {
  group: Group;
  membership: GroupMembership;
  activeSeason: Season | null;
  recentStunts: unknown[];
  standings: unknown[];
}

export interface GroupMembershipSummary {
  group: Group;
  membership: GroupMembership;
}

export interface ListGroupsResponse {
  memberships: GroupMembershipSummary[];
}
