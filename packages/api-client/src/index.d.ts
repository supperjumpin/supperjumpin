import type { MeResponse } from "./generated";

export type { Account, MeResponse, Player } from "./generated";

export function getMe(args: {
  baseUrl: string;
  accessToken: string;
  fetchImpl?: typeof fetch;
}): Promise<MeResponse>;
