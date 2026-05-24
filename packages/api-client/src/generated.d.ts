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

export interface GroupHomeResponse {
  group: Group;
  membership: GroupMembership;
  activeSeason: null;
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
