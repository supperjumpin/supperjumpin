import { render, waitFor } from '@testing-library/react-native';
import App from './App';
import { mockPublicFetch, feedResponse } from './test/mockApi';

test('renders feed screen with Post Jump button even when unauthenticated', async () => {
  mockPublicFetch().mockImplementation(async (url) => {
    if (url.toString().includes('/v1/feed')) {
      return Response.json(feedResponse([]));
    }
    return Response.json({});
  });

  const renderResult = await render(<App />);

  await waitFor(() => {
    // Feed title is always visible
    expect(renderResult.getByText('Supperjumpin')).toBeTruthy();
    // + Jump button is always visible (triggers auth when tapped)
    expect(renderResult.getByText('+ Jump')).toBeTruthy();
  }, { timeout: 10000 });
}, 15000);