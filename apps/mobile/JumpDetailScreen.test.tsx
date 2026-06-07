import { act } from 'react';
import { fireEvent, render, waitFor } from '@testing-library/react-native';
import JumpDetailScreen from './JumpDetailScreen';
import { mockPublicFetch, jumpDetailResponse } from './test/mockApi';

test('Jump detail exposes loading, performer stub, and back actions accessibly', async () => {
  const fetchSpy = mockPublicFetch();
  let resolveDetail: (response: Response) => void = () => {};
  fetchSpy.mockImplementation(
    () => new Promise<Response>((resolve) => { resolveDetail = resolve; })
  );

  const { getByLabelText, queryByLabelText } = await render(
    <JumpDetailScreen
      jumpId="jump_1"
      onBack={() => {}}
      onBrowseFeed={() => {}}
    />
  );

  expect(getByLabelText('Loading jump detail')).toBeTruthy();
  expect(queryByLabelText('Back to Feed')).toBeNull();

  resolveDetail(Response.json(jumpDetailResponse()));

  await waitFor(() => {
    const backButton = getByLabelText('Back to Feed');
    expect(backButton.props.accessibilityRole).toBe('button');
    expect(backButton.props.style).toEqual(expect.objectContaining({ minHeight: 44 }));

    const performerTarget = getByLabelText('alice profile coming soon');
    expect(performerTarget.props.accessibilityRole).toBe('button');
    expect(performerTarget.props.accessibilityState).toEqual(expect.objectContaining({ disabled: true }));
    expect(performerTarget.props.style).toEqual(expect.objectContaining({ minHeight: 44 }));
  });
});

test.each([
  {
    name: 'self-judging',
    viewerContext: { canJudge: false, hasJudged: false, reason: 'self-judging' },
    expectedText: "You can't judge your own Jump",
  },
  {
    name: 'grace-period',
    viewerContext: { canJudge: false, hasJudged: false, reason: 'grace-period' },
    expectedText: /Judging opens in/,
    gracePeriodExpiresAt: new Date(Date.now() + 5 * 60 * 1000).toISOString(),
  },
  {
    name: 'already-judged',
    viewerContext: { canJudge: false, hasJudged: true, reason: 'already-judged' },
    expectedText: 'You already judged this jump. Score submitted.',
  },
  {
    name: 'can-judge',
    viewerContext: { canJudge: true, hasJudged: false },
    expectedText: 'Judge this Jump',
  },
])('signed-in Jump detail requests include bearer auth and render $name viewer state', async ({ viewerContext, expectedText, gracePeriodExpiresAt }) => {
  const fetchSpy = mockPublicFetch();
  fetchSpy.mockImplementation(async (_url: RequestInfo | URL, init?: RequestInit) => {
    expect(init).toEqual(
      expect.objectContaining({
        headers: expect.objectContaining({
          Accept: 'application/json',
          Authorization: 'Bearer tok_detail',
        }),
      })
    );

    return Response.json(jumpDetailResponse({ viewerContext, gracePeriodExpiresAt }));
  });

  const { getByText, getByLabelText } = await render(
    <JumpDetailScreen
      jumpId="jump_1"
      onBack={() => {}}
      onBrowseFeed={() => {}}
      session={{ access_token: 'tok_detail', user: { id: 'u1' } }}
    />
  );

  await waitFor(() => {
    if (expectedText instanceof RegExp) {
      expect(getByText(expectedText)).toBeTruthy();
      return;
    }

    expect(getByText(expectedText)).toBeTruthy();
    expect(getByLabelText(expectedText)).toBeTruthy();
  });
});

test('renders normal jump detail from public-read API data', async () => {
  const fetchSpy = mockPublicFetch();
  fetchSpy.mockImplementation(async (url: RequestInfo | URL) => {
    if (url.toString().includes('/v1/jumps/jump_1')) {
      return Response.json(jumpDetailResponse({
        caption: 'Carried the Crunchwrap cleanly across the lot',
        runningAverage: 3.25,
        judgmentCount: 8,
      }));
    }
    return Response.json({});
  });

  const { getByLabelText, getByText } = await render(
    <JumpDetailScreen
      jumpId="jump_1"
      onBack={() => {}}
      onBrowseFeed={() => {}}
    />
  );

  await waitFor(() => {
    expect(getByText('alice')).toBeTruthy();
    expect(getByText('Crunchwrap')).toBeTruthy();
    expect(getByText('Taco Bell → Olive Garden parking lot')).toBeTruthy();
    expect(getByText('Carried the Crunchwrap cleanly across the lot')).toBeTruthy();
    expect(getByText('3.3')).toBeTruthy();
    expect(getByText('from 8 judgments')).toBeTruthy();
    expect(getByLabelText('Judge this Jump')).toBeTruthy();
  });
});

test('renders tombstone detail and Browse Feed returns to Feed', async () => {
  const onBrowseFeed = jest.fn();
  const fetchSpy = mockPublicFetch();
  fetchSpy.mockImplementation(async (url: RequestInfo | URL) => {
    if (url.toString().includes('/v1/jumps/jump_1')) {
      return Response.json(jumpDetailResponse({ status: 'Removed Jump' }));
    }
    return Response.json({});
  });

  const { getByText } = await render(
    <JumpDetailScreen
      jumpId="jump_1"
      onBack={() => {}}
      onBrowseFeed={onBrowseFeed}
    />
  );

  await waitFor(() => {
    expect(getByText('This Jump is no longer available')).toBeTruthy();
  });

  fireEvent.press(getByText('Browse Feed'));

  expect(onBrowseFeed).toHaveBeenCalledTimes(1);
});

test('shows error state when jump detail cannot be loaded', async () => {
  const onBack = jest.fn();
  const fetchSpy = mockPublicFetch();
  fetchSpy.mockRejectedValueOnce(new Error('Jump not found'));

  const { getByText } = await render(
    <JumpDetailScreen
      jumpId="missing_jump"
      onBack={onBack}
      onBrowseFeed={() => {}}
    />
  );

  await waitFor(() => {
    expect(getByText('Jump not found')).toBeTruthy();
  });

  fireEvent.press(getByText('Back to Feed'));

  expect(onBack).toHaveBeenCalledTimes(1);
});

test('guest Jump detail requests omit bearer auth', async () => {
  const fetchSpy = mockPublicFetch();
  fetchSpy.mockImplementation(async (_url: RequestInfo | URL, init?: RequestInit) => {
    expect(init).toEqual(
      expect.objectContaining({
        headers: { Accept: 'application/json' },
      })
    );

    return Response.json(jumpDetailResponse());
  });

  const { getByLabelText } = await render(
    <JumpDetailScreen
      jumpId="jump_1"
      onBack={() => {}}
      onBrowseFeed={() => {}}
    />
  );

  await waitFor(() => {
    expect(getByLabelText('Judge this Jump')).toBeTruthy();
  });
});

test('grace-period countdown updates live on Jump detail', async () => {
  jest.useFakeTimers();
  jest.setSystemTime(new Date('2026-06-01T00:00:00Z'));

  const fetchSpy = mockPublicFetch();
  fetchSpy.mockImplementation(async () =>
    Response.json(
      jumpDetailResponse({
        gracePeriodExpiresAt: '2026-06-01T00:00:05Z',
        viewerContext: { canJudge: false, hasJudged: false, reason: 'grace-period' },
      })
    )
  );

  const { getByText } = await render(
    <JumpDetailScreen
      jumpId="jump_1"
      onBack={() => {}}
      onBrowseFeed={() => {}}
    />
  );

  await waitFor(() => {
    expect(getByText('Judging opens in 0m 5s')).toBeTruthy();
  });

  await act(async () => {
    jest.advanceTimersByTime(1000);
  });

  expect(getByText('Judging opens in 0m 4s')).toBeTruthy();

  jest.useRealTimers();
});
