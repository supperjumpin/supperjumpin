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

  await act(() => {
    resolveFeed(Response.json(feedResponse([])));
  });

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
    expect(getByText('Judge this Jump')).toBeTruthy();
  });
});

test('Feed exposes performer attribution as a disabled profile stub target', async () => {
  const fetchSpy = mockPublicFetch();
  fetchSpy.mockImplementation(async () => Response.json(feedResponse([publicJump()])));

  const { getByLabelText } = await render(<FeedScreen onNavigateDetail={() => {}} />);

  await waitFor(() => {
    const performerTarget = getByLabelText('alice profile coming soon');
    expect(performerTarget.props.accessibilityRole).toBe('button');
    expect(performerTarget.props.accessibilityState).toEqual(expect.objectContaining({ disabled: true }));
    expect(performerTarget.props.style).toEqual(expect.objectContaining({ minHeight: 44 }));
  });
});

test('Feed card shows media placeholder when mediaObjectKey is empty', async () => {
  const fetchSpy = mockPublicFetch();
  fetchSpy.mockImplementation(async () => Response.json(feedResponse([publicJump()])));

  const { getByText } = await render(<FeedScreen onNavigateDetail={() => {}} />);

  await waitFor(() => {
    expect(getByText('📸')).toBeTruthy();
  });
});

test('signed-in Feed requests include bearer auth and render viewer states from the API', async () => {
  const fetchSpy = mockPublicFetch();
  const futureGracePeriod = new Date(Date.now() + 5 * 60 * 1000).toISOString();
  fetchSpy.mockImplementation(async () =>
    Response.json(
      feedResponse([
        publicJump({
          id: 'jump_self',
          food: 'Pizza',
          viewerContext: { canJudge: false, hasJudged: false, reason: 'self-judging' },
        }),
        publicJump({
          id: 'jump_grace',
          food: 'Burrito',
          gracePeriodExpiresAt: futureGracePeriod,
          viewerContext: { canJudge: false, hasJudged: false, reason: 'grace-period' },
        }),
        publicJump({
          id: 'jump_done',
          food: 'Pasta',
          viewerContext: { canJudge: false, hasJudged: true, reason: 'already-judged' },
        }),
        publicJump({
          id: 'jump_open',
          food: 'Salad',
          viewerContext: { canJudge: true, hasJudged: false },
        }),
      ])
    )
  );

  const { getAllByLabelText, getByText } = await render(
    <FeedScreen
      onNavigateDetail={() => {}}
      session={{ access_token: 'tok_feed', user: { id: 'u1' } }}
    />
  );

  await waitFor(() => {
    expect(fetchSpy).toHaveBeenCalledWith(
      'http://localhost:8080/v1/feed?limit=20',
      expect.objectContaining({
        headers: expect.objectContaining({
          Accept: 'application/json',
          Authorization: 'Bearer tok_feed',
        }),
      })
    );
    expect(getByText("You can't judge your own Jump")).toBeTruthy();
    expect(getByText(/Judging opens in/)).toBeTruthy();
    expect(getByText('You already judged this jump. Score submitted.')).toBeTruthy();
    expect(getByText('Judge this Jump')).toBeTruthy();
    expect(getAllByLabelText('You performed this jump. You cannot judge your own entry.').length).toBeGreaterThan(0);
    expect(getAllByLabelText(/Judging opens in .* Not yet available\./).length).toBeGreaterThan(0);
    expect(getAllByLabelText('You already judged this jump. Score submitted.').length).toBeGreaterThan(0);
    expect(getAllByLabelText('Judge this Jump').length).toBeGreaterThan(0);
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

test('guest Feed requests omit bearer auth', async () => {
  const fetchSpy = mockPublicFetch();
  fetchSpy.mockImplementation(async () => Response.json(feedResponse([])));

  await render(<FeedScreen onNavigateDetail={() => {}} />);

  await waitFor(() => {
    expect(fetchSpy).toHaveBeenCalledWith(
      'http://localhost:8080/v1/feed?limit=20',
      expect.objectContaining({
        headers: { Accept: 'application/json' },
      })
    );
  });
});

test('Feed key interactive targets meet minimum touch size', async () => {
  const fetchSpy = mockPublicFetch();
  fetchSpy.mockImplementation(async () => Response.json(feedResponse([publicJump()])));

  const { getAllByLabelText, getByLabelText } = await render(<FeedScreen onNavigateDetail={() => {}} onRequestAuth={() => {}} />);

  await waitFor(() => {
    expect(getByLabelText('Post a new Jump').props.style).toEqual(expect.objectContaining({ minHeight: 44 }));
    expect(getAllByLabelText('Judge this Jump')[0].props.style).toEqual(
      expect.arrayContaining([expect.objectContaining({ minHeight: 44 })])
    );
    expect(getByLabelText('alice profile coming soon').props.style).toEqual(expect.objectContaining({ minHeight: 44 }));
  });
});

test('Feed card renders media Image when mediaObjectKey is provided', async () => {
  const fetchSpy = mockPublicFetch();
  fetchSpy.mockImplementation(async () =>
    Response.json(feedResponse([publicJump({ mediaObjectKey: 'photos/has-key.jpg' })]))
  );

  const { getByText, queryByText } = await render(<FeedScreen onNavigateDetail={() => {}} />);

  await waitFor(() => {
    expect(queryByText('📸')).toBeNull();
    expect(getByText('Crunchwrap')).toBeTruthy();
    expect(getByText('Test caption')).toBeTruthy();
  });
});
