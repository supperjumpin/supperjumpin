import type {
  Judgment,
  MeResponse,
  PublicFeedResponse,
  JumpDetail,
  JumpTombstone,
} from "./generated";

export type {
  Judgment,
  MeResponse,
  Player,
  Jump,
  JumpCard,
  JumpDetail,
  JumpTombstone,
  PublicFeedResponse,
  ViewerContext,
} from "./generated";

export function getMe(args: {
  baseUrl: string;
  accessToken: string;
  fetchImpl?: typeof fetch;
}): Promise<MeResponse>;

export function submitJudgment(args: {
  baseUrl: string;
  accessToken: string;
  jumpId: string;
  commitment: number;
  transgression: number;
  creativity: number;
  presentation: number;
  fetchImpl?: typeof fetch;
}): Promise<Judgment>;

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
