import type { GroupHomeResponse, ListGroupsResponse, MeResponse } from "./generated";

export type {
  Account,
  Group,
  GroupHomeResponse,
  GroupMembership,
  GroupMembershipSummary,
  ListGroupsResponse,
  MeResponse,
  Player,
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
