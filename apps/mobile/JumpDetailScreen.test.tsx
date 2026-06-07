import { fireEvent, render, waitFor } from '@testing-library/react-native';
import JumpDetailScreen from './JumpDetailScreen';
import { mockPublicFetch, jumpDetailResponse } from './test/mockApi';

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
