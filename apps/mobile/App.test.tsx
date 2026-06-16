import { act } from 'react';
import { fireEvent, render, waitFor } from '@testing-library/react-native';
import App from './App';
import { mockPublicFetch, feedResponse, jumpDetailResponse } from './test/mockApi';

beforeEach(() => {
  jest.clearAllMocks();
});

test('renders feed screen with Post Jump button even when unauthenticated', async () => {
  mockPublicFetch().mockImplementation(async (url) => {
    if (url.toString().includes('/v1/feed')) {
      return Response.json(feedResponse([]));
    }
    return Response.json({});
  });

  const renderResult = await render(<App />);

  await waitFor(() => {
    expect(renderResult.getByText('Supperjumpin')).toBeTruthy();
    expect(renderResult.getByText('+ Jump')).toBeTruthy();
    expect(renderResult.getByText('No jumps yet. The feed is empty.')).toBeTruthy();
  }, { timeout: 10000 });
});

test('renders feed after public-read API succeeds', async () => {
  const fetchSpy = mockPublicFetch();
  fetchSpy.mockImplementation(async (url: RequestInfo | URL) => {
    if (url.toString().includes('/v1/feed')) {
      return Response.json(
        feedResponse([
          {
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
          },
        ])
      );
    }
    return Response.json({});
  });

  const { getByText } = await render(<App />);

  await waitFor(() => {
    expect(getByText('Supperjumpin')).toBeTruthy();
    expect(getByText('alice')).toBeTruthy();
    expect(getByText('Crunchwrap')).toBeTruthy();
  });
});

test('opens jump detail after tapping a Feed card', async () => {
  const fetchSpy = mockPublicFetch();
  fetchSpy.mockImplementation(async (url: RequestInfo | URL) => {
    if (url.toString().includes('/v1/feed')) {
      return Response.json(
        feedResponse([
          {
            id: 'jump_1',
            performerId: 'player_1',
            performerName: 'alice',
            source: 'Taco Bell',
            destination: 'Olive Garden parking lot',
            food: 'Crunchwrap',
            caption: 'Feed caption',
            mediaObjectKey: '',
            status: 'Performed Jump',
            gracePeriodExpiresAt: '2026-06-01T00:10:00Z',
            runningAverage: 3.5,
            judgmentCount: 4,
            createdAt: '2026-06-01T00:00:00Z',
            viewerContext: { canJudge: true, hasJudged: false },
          },
        ])
      );
    }
    if (url.toString().includes('/v1/jumps/jump_1')) {
      return Response.json(jumpDetailResponse({ caption: 'Detail caption' }));
    }
    return Response.json({});
  });

  const { getByRole, getByText } = await render(<App />);

  await waitFor(() => {
    expect(getByText('Crunchwrap')).toBeTruthy();
  });

  await act(async () => {
    fireEvent.press(getByRole('button', { name: /alice jumped Crunchwrap/ }));
  });

  await waitFor(() => {
    expect(getByText('Detail caption')).toBeTruthy();
    expect(getByText('Running Average')).toBeTruthy();
  });
});

test('signed-in players open Jump detail with bearer auth and viewer context', async () => {
  const fetchSpy = mockPublicFetch();
  fetchSpy.mockImplementation(async (url: RequestInfo | URL, init?: RequestInit) => {
    if (url.toString().includes('/v1/feed')) {
      return Response.json(
        feedResponse([
          {
            id: 'jump_1',
            performerId: 'player_1',
            performerName: 'alice',
            source: 'Taco Bell',
            destination: 'Olive Garden parking lot',
            food: 'Crunchwrap',
            caption: 'Feed caption',
            mediaObjectKey: '',
            status: 'Performed Jump',
            gracePeriodExpiresAt: '2026-06-01T00:10:00Z',
            runningAverage: 3.5,
            judgmentCount: 4,
            createdAt: '2026-06-01T00:00:00Z',
            viewerContext: { canJudge: true, hasJudged: false },
          },
        ])
      );
    }
    if (url.toString().includes('/v1/me')) {
      return Response.json({ player: { id: 'player_1', displayName: 'alice' } });
    }
    if (url.toString().includes('/v1/jumps/jump_1')) {
      expect(init).toEqual(
        expect.objectContaining({
          headers: expect.objectContaining({
            Accept: 'application/json',
            Authorization: 'Bearer dev-token',
          }),
        })
      );
      return Response.json(
        jumpDetailResponse({
          viewerContext: { canJudge: false, hasJudged: false, reason: 'self-judging' },
        })
      );
    }
    return Response.json({});
  });

  const { getByRole, getByText } = await render(<App />);

  await waitFor(() => {
    expect(getByText('Crunchwrap')).toBeTruthy();
  });

  await act(async () => {
    fireEvent.press(getByText('+ Jump'));
  });

  await waitFor(() => {
    expect(getByText('Post a Jump')).toBeTruthy();
  });

  await act(async () => {
    fireEvent.press(getByText('Back to Feed'));
  });

  await waitFor(() => {
    expect(getByText('Crunchwrap')).toBeTruthy();
  });

  await act(async () => {
    fireEvent.press(getByRole('button', { name: /alice jumped Crunchwrap/ }));
  });

  await waitFor(() => {
    expect(getByText("You can't judge your own Jump")).toBeTruthy();
  });
});

test('routes local dev sign-in without a display name through setup before creating a Jump', async () => {
  const fetchSpy = mockPublicFetch();
  fetchSpy.mockImplementation(async (url: RequestInfo | URL) => {
    if (url.toString().includes('/v1/feed')) {
      return Response.json(feedResponse([]));
    }
    if (url.toString().includes('/v1/me')) {
      return Response.json({ player: { id: 'player_1', displayName: '' } });
    }
    return Response.json({});
  });
  const { getByText } = await render(<App />);

  await waitFor(() => {
    expect(getByText('+ Jump')).toBeTruthy();
  });

  await act(async () => {
    fireEvent.press(getByText('+ Jump'));
  });

  await waitFor(() => {
    expect(getByText('Choose your display name')).toBeTruthy();
  });
});

test('routes unauthenticated Post Jump taps through sign in and into Create Jump', async () => {
  const fetchSpy = mockPublicFetch();
  fetchSpy.mockImplementation(async (url: RequestInfo | URL) => {
    if (url.toString().includes('/v1/feed')) {
      return Response.json(feedResponse([]));
    }
    if (url.toString().includes('/v1/me')) {
      return Response.json({ player: { id: 'player_1', displayName: 'alice' } });
    }
    return Response.json({});
  });

  const { getByText } = await render(<App />);

  await waitFor(() => {
    expect(getByText('+ Jump')).toBeTruthy();
  });

  await act(async () => {
    fireEvent.press(getByText('+ Jump'));
  });

  await waitFor(() => {
    expect(getByText('Post a Jump')).toBeTruthy();
  });
});

test('preserves a composed Jump draft when leaving and returning to Create Jump', async () => {
  const fetchSpy = mockPublicFetch();
  fetchSpy.mockImplementation(async (url: RequestInfo | URL) => {
    if (url.toString().includes('/v1/feed')) {
      return Response.json(feedResponse([]));
    }
    if (url.toString().includes('/v1/me')) {
      return Response.json({ player: { id: 'player_1', displayName: 'alice' } });
    }
    return Response.json({});
  });
  const { getByText, getByTestId } = await render(<App />);

  await waitFor(() => {
    expect(getByText('+ Jump')).toBeTruthy();
  });

  await act(async () => {
    fireEvent.press(getByText('+ Jump'));
  });

  await waitFor(() => {
    expect(getByText('Post a Jump')).toBeTruthy();
  });

  await act(async () => {
    fireEvent.changeText(getByTestId('source-input'), 'Taco Bell');
    fireEvent.changeText(getByTestId('destination-input'), 'Olive Garden');
    fireEvent.changeText(getByTestId('food-input'), 'Crunchwrap');
    fireEvent.changeText(getByTestId('caption-input'), 'Keep this draft');
    fireEvent.changeText(getByTestId('evidence-photo-input'), 'evidence/jump_1.jpg');
  });

  await act(async () => {
    fireEvent.press(getByText('Back to Feed'));
  });

  await waitFor(() => {
    expect(getByText('+ Jump')).toBeTruthy();
  });

  await act(async () => {
    fireEvent.press(getByText('+ Jump'));
  });

  await waitFor(() => {
    expect(getByText('Post a Jump')).toBeTruthy();
  });

  expect(getByTestId('source-input').props.value).toBe('Taco Bell');
  expect(getByTestId('caption-input').props.value).toBe('Keep this draft');
  expect(getByTestId('evidence-photo-input').props.value).toBe('evidence/jump_1.jpg');
});
