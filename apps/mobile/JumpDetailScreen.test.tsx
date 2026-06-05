import { render, waitFor } from '@testing-library/react-native';
import JumpDetailScreen from './JumpDetailScreen';
import { mockPublicFetch, jumpDetailResponse } from './test/mockApi';

test('renders jump detail from public-read API', async () => {
  const fetchSpy = mockPublicFetch();
  fetchSpy.mockImplementation(async (url: RequestInfo | URL) => {
    if (url.toString().includes('/v1/jumps/jump_1')) {
      return Response.json(jumpDetailResponse());
    }
    return Response.json({});
  });

  const { getByText } = await render(
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
  });
});
