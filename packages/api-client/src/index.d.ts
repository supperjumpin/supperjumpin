import type { components } from "./generated";

export type Judgment = components["schemas"]["Judgment"];
export type MeResponse = components["schemas"]["MeResponse"];
export type Player = components["schemas"]["Player"];
export type Jump = components["schemas"]["Jump"];
export type JumpCard = components["schemas"]["JumpCard"];
export type JumpDetail = components["schemas"]["JumpDetail"];
export type JumpTombstone = components["schemas"]["JumpTombstone"];
export type PublicFeedResponse = components["schemas"]["PublicFeedResponse"];
export type UpdateDisplayNameResponse = components["schemas"]["UpdateDisplayNameResponse"];
export type ViewerContext = components["schemas"]["ViewerContext"];

export function getMe(args: {
  baseUrl: string;
  accessToken: string;
  fetchImpl?: typeof fetch;
}): Promise<MeResponse>;

export function updateDisplayName(args: {
  baseUrl: string;
  accessToken: string;
  displayName: string;
  fetchImpl?: typeof fetch;
}): Promise<UpdateDisplayNameResponse>;

export function createJump(args: {
  baseUrl: string;
  accessToken: string;
  source: string;
  destination: string;
  food: string;
  caption: string;
  mediaObjectKey: string;
  fetchImpl?: typeof fetch;
}): Promise<Jump>;

export function submitJudgment(args: {
  baseUrl: string;
  accessToken?: string;
  guestSessionId?: string;
  jumpId: string;
  commitment: number;
  transgression: number;
  creativity: number;
  presentation: number;
  fetchImpl?: typeof fetch;
}): Promise<Judgment>;

export function createGuestSession(args: {
  baseUrl: string;
  fetchImpl?: typeof fetch;
}): Promise<{ id: string }>;

export function getPublicFeed(args: {
  baseUrl: string;
  accessToken?: string;
  cursor?: string;
  limit?: number;
  fetchImpl?: typeof fetch;
}): Promise<PublicFeedResponse>;

export function getJumpDetail(args: {
  baseUrl: string;
  accessToken?: string;
  jumpId: string;
  fetchImpl?: typeof fetch;
}): Promise<JumpDetail | JumpTombstone>;
