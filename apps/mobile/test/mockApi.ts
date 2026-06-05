import { jest } from '@jest/globals';

/**
 * Reusable mocking pattern for @supperjumpin/api-client public-read calls.
 * The api-client functions default to global.fetch, so we intercept at the
 * transport layer. This avoids module-mocking ESM internals.
 */
export function mockPublicFetch() {
  return jest.spyOn(global, 'fetch').mockImplementation(async () =>
    Response.json({})
  );
}

/**
 * Feed response shape matching PublicFeedResponse.
 */
export function feedResponse(
  jumps: unknown[],
  nextCursor: string | null = null
): { jumps: unknown[]; nextCursor: string | null } {
  return { jumps, nextCursor };
}

/**
 * Jump detail response shape matching JumpDetail.
 */
export function jumpDetailResponse(overrides: Record<string, unknown> = {}): unknown {
  return {
    id: 'jump_1',
    performerId: 'player_1',
    performerName: 'alice',
    source: 'Taco Bell',
    destination: 'Olive Garden parking lot',
    food: 'Crunchwrap',
    caption: 'Test caption',
    mediaObjectKey: '',
    status: 'Performed Jump',
    gracePeriodExpiresAt: '2026-06-01T00:10:00Z',
    runningAverage: 3.5,
    judgmentCount: 4,
    createdAt: '2026-06-01T00:00:00Z',
    viewerContext: { canJudge: true, hasJudged: false },
    ...overrides,
  };
}
