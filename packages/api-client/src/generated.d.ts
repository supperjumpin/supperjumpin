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
