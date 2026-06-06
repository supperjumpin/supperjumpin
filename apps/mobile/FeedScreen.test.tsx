import { render, waitFor } from '@testing-library/react-native';
import FeedScreen from './FeedScreen';
import { mockPublicFetch, feedResponse } from './test/mockApi';

test('renders jumps from public-read API', async () => {
  const fetchSpy = mockPublicFetch();
  fetchSpy.mockImplementation(async (url) => {
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

  const { getByText } = await render(<FeedScreen onNavigateDetail={() => {}} />);

  await waitFor(
    () => {
      expect(getByText('alice')).toBeTruthy();
      expect(getByText('Crunchwrap')).toBeTruthy();
      expect(getByText('Taco Bell → Olive Garden parking lot')).toBeTruthy();
    },
    { timeout: 10000 }
  );
}, 15000);

test('shows Post Jump button when session is provided', async () => {
  const fetchSpy = mockPublicFetch();
  fetchSpy.mockImplementation(async (url) => {
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
}, 15000);

test('shows Post Jump button when no session', async () => {
  const fetchSpy = mockPublicFetch();
  fetchSpy.mockImplementation(async (url) => {
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
}, 15000);