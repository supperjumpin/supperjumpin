import { act } from 'react';
import { fireEvent, render, waitFor } from '@testing-library/react-native';
import FeedScreen from './FeedScreen';
import { mockPublicFetch, feedResponse } from './test/mockApi';

function publicJump(overrides: Record<string, unknown> = {}) {
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

test('shows initial loading state while Feed request is pending', async () => {
  const fetchSpy = mockPublicFetch();
  let resolveFeed: (response: Response) => void = () => {};
  fetchSpy.mockImplementation(
    () => new Promise<Response>((resolve) => { resolveFeed = resolve; })
  );

  const { getByText } = await render(<FeedScreen onNavigateDetail={() => {}} />);

  expect(getByText('Loading jumps...')).toBeTruthy();

  resolveFeed(Response.json(feedResponse([])));

  await waitFor(() => {
    expect(getByText('No jumps yet. The feed is empty.')).toBeTruthy();
  });
});

test('shows empty Feed state when public-read API returns no jumps', async () => {
  const fetchSpy = mockPublicFetch();
  fetchSpy.mockImplementation(async (url: RequestInfo | URL) => {
    if (url.toString().includes('/v1/feed')) {
      return Response.json(feedResponse([]));
    }
    return Response.json({});
  });

  const { getByText } = await render(<FeedScreen onNavigateDetail={() => {}} />);

  await waitFor(() => {
    expect(getByText('No jumps yet. The feed is empty.')).toBeTruthy();
  });
});

test('shows Feed error state and retries after a failed request', async () => {
  const fetchSpy = mockPublicFetch();
  fetchSpy
    .mockRejectedValueOnce(new Error('Network error'))
    .mockImplementationOnce(async () => Response.json(feedResponse([publicJump()])));

  const { getByText } = await render(<FeedScreen onNavigateDetail={() => {}} />);

  await waitFor(() => {
    expect(getByText('Could not load jumps. Network error.')).toBeTruthy();
  });

  fireEvent.press(getByText('Retry'));

  await waitFor(() => {
    expect(getByText('Crunchwrap')).toBeTruthy();
  });
});

test('renders Jump card anatomy from public Feed data', async () => {
  const fetchSpy = mockPublicFetch();
  fetchSpy.mockImplementation(async (url) => {
    if (url.toString().includes('/v1/feed')) {
      return Response.json(
        feedResponse([
          publicJump(),
        ])
      );
    }
    return Response.json({});
  });

  const { getByText } = await render(<FeedScreen onNavigateDetail={() => {}} />);

  await waitFor(() => {
    expect(getByText('alice')).toBeTruthy();
    expect(getByText('Crunchwrap')).toBeTruthy();
    expect(getByText('Taco Bell → Olive Garden parking lot')).toBeTruthy();
    expect(getByText('Test caption')).toBeTruthy();
    expect(getByText('★ 3.5 (4)')).toBeTruthy();
    expect(getByText('Judge')).toBeTruthy();
  });
});

test('pull-to-refresh replaces existing Feed contents', async () => {
  const fetchSpy = mockPublicFetch();
  fetchSpy
    .mockImplementationOnce(async () => Response.json(feedResponse([
      publicJump({ id: 'jump_old', food: 'Crunchwrap' }),
    ])))
    .mockImplementationOnce(async () => Response.json(feedResponse([
      publicJump({ id: 'jump_new', performerName: 'bob', food: 'Lasagna' }),
    ])));

  const { getByTestId, getByText, queryByText } = await render(
    <FeedScreen onNavigateDetail={() => {}} />
  );

  await waitFor(() => {
    expect(getByText('Crunchwrap')).toBeTruthy();
  });

  await act(async () => {
    await getByTestId('feed-list').props.refreshControl.props.onRefresh();
  });

  await waitFor(() => {
    expect(getByText('Lasagna')).toBeTruthy();
    expect(queryByText('Crunchwrap')).toBeNull();
  });
});

test('load-more appends Feed cards without dropping existing cards', async () => {
  const fetchSpy = mockPublicFetch();
  fetchSpy
    .mockImplementationOnce(async () => Response.json(feedResponse([
      publicJump({ id: 'jump_1', food: 'Crunchwrap' }),
    ], 'cursor_2')))
    .mockImplementationOnce(async () => Response.json(feedResponse([
      publicJump({ id: 'jump_2', performerName: 'bob', food: 'Lasagna' }),
    ])));

  const { getByTestId, getByText } = await render(
    <FeedScreen onNavigateDetail={() => {}} />
  );

  await waitFor(() => {
    expect(getByText('Crunchwrap')).toBeTruthy();
  });

  fireEvent(getByTestId('feed-list'), 'endReached');

  await waitFor(() => {
    expect(getByText('Crunchwrap')).toBeTruthy();
    expect(getByText('Lasagna')).toBeTruthy();
  });
});

test('shows Post Jump button when session is provided', async () => {
  const fetchSpy = mockPublicFetch();
  fetchSpy.mockImplementation(async (url: RequestInfo | URL) => {
    if (url.toString().includes('/v1/feed')) {
      return Response.json(feedResponse([]));
    }
    return Response.json({});
  });

  const { getByText } = await render(
    <FeedScreen
      onNavigateDetail={() => {}}
      onRequestAuth={() => {}}
      session={{ access_token: 'tok', user: { id: 'u1' } }}
    />
  );

  await waitFor(
    () => {
      expect(getByText('+ Jump')).toBeTruthy();
    },
    { timeout: 10000 }
  );
});

test('shows Post Jump button when no session', async () => {
  const fetchSpy = mockPublicFetch();
  fetchSpy.mockImplementation(async (url: RequestInfo | URL) => {
    if (url.toString().includes('/v1/feed')) {
      return Response.json(feedResponse([]));
    }
    return Response.json({});
  });

  const { getByText } = await render(<FeedScreen onNavigateDetail={() => {}} onRequestAuth={() => {}} />);

  await waitFor(
    () => {
      expect(getByText('+ Jump')).toBeTruthy();
    },
    { timeout: 10000 }
  );
});
