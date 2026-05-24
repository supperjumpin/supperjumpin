import type {
  EvidenceSubmission,
  EvidenceUploadAuthorization,
  GroupHomeResponse,
  Invite,
  Judgment,
  ListGroupsResponse,
  MeResponse,
  PerformedStuntView,
  Stunt,
} from "./generated";

export type {
  Account,
  Group,
  GroupHomeResponse,
  GroupMembership,
  GroupMembershipSummary,
  Evidence,
  EvidenceSubmission,
  EvidenceUploadAuthorization,
  Invite,
  Judgment,
  ListGroupsResponse,
  MeResponse,
  PerformedStuntView,
  Player,
  Season,
  Stunt,
} from "./generated";

export function getMe(args: {
  baseUrl: string;
  accessToken: string;
  fetchImpl?: typeof fetch;
}): Promise<MeResponse>;

export function createGroup(args: {
  baseUrl: string;
  accessToken: string;
  name: string;
  fetchImpl?: typeof fetch;
}): Promise<GroupHomeResponse>;

export function listGroups(args: {
  baseUrl: string;
  accessToken: string;
  fetchImpl?: typeof fetch;
}): Promise<ListGroupsResponse>;

export function getGroupHome(args: {
  baseUrl: string;
  accessToken: string;
  groupId: string;
  fetchImpl?: typeof fetch;
}): Promise<GroupHomeResponse>;

export function createInvite(args: {
  baseUrl: string;
  accessToken: string;
  groupId: string;
  fetchImpl?: typeof fetch;
}): Promise<Invite>;

export function acceptInvite(args: {
  baseUrl: string;
  accessToken: string;
  token: string;
  fetchImpl?: typeof fetch;
}): Promise<GroupHomeResponse>;

export function startSeason(args: {
  baseUrl: string;
  accessToken: string;
  groupId: string;
  fetchImpl?: typeof fetch;
}): Promise<GroupHomeResponse>;

export function createIdea(args: {
  baseUrl: string;
  accessToken: string;
  groupId: string;
  source: string;
  destination: string;
  food: string;
  fetchImpl?: typeof fetch;
}): Promise<Stunt>;

export function createPlannedStunt(args: {
  baseUrl: string;
  accessToken: string;
  ideaId: string;
  offSeason?: boolean;
  fetchImpl?: typeof fetch;
}): Promise<Stunt>;

export function authorizeEvidenceUpload(args: {
  baseUrl: string;
  accessToken: string;
  stuntId: string;
  contentType: string;
  fetchImpl?: typeof fetch;
}): Promise<EvidenceUploadAuthorization>;

export function submitEvidence(args: {
  baseUrl: string;
  accessToken: string;
  stuntId: string;
  uploadAuthorizationId: string;
  caption: string;
  fetchImpl?: typeof fetch;
}): Promise<EvidenceSubmission>;

export function submitJudgment(args: {
  baseUrl: string;
  accessToken: string;
  stuntId: string;
  difficulty: number;
  transgression: number;
  creativity: number;
  documentation: number;
  fetchImpl?: typeof fetch;
}): Promise<Judgment>;
