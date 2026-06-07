import { act } from 'react';
import { fireEvent, render, waitFor } from '@testing-library/react-native';
import App from './App';
import { mockPublicFetch, feedResponse, jumpDetailResponse } from './test/mockApi';

test('renders empty feed when public-read API returns no jumps', async () => {
  const fetchSpy = mockPublicFetch();
  fetchSpy.mockImplementation(async (url: RequestInfo | URL) => {
    if (url.toString().includes('/v1/feed')) {
      return Response.json(feedResponse([]));
    }
    return Response.json({});
  });

  const { getByText } = await render(<App />);

  await waitFor(() => {
    expect(getByText('Supperjumpin')).toBeTruthy();
    expect(getByText('No jumps yet. The feed is empty.')).toBeTruthy();
  });
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
