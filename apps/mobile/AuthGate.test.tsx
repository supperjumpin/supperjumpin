import { render, waitFor } from '@testing-library/react-native';
import { Text } from 'react-native';
import AuthGate from './AuthGate';

test('renders LoginScreen when no session', async () => {
  const renderResult = await render(
    <AuthGate session={null} player={null}>
      <Text>Feed content</Text>
    </AuthGate>
  );

  await waitFor(() => {
    expect(renderResult.getByText('Sign In')).toBeTruthy();
    expect(renderResult.queryByText('Feed content')).toBeNull();
  });
});

test('renders DisplayNameSetupScreen when player has no display name', async () => {
  const session = { access_token: 'tok', user: { id: 'u1' } };
  const player = { id: 'p1', displayName: '' };

  const renderResult = await render(
    <AuthGate session={session} player={player}>
      <Text>Feed content</Text>
    </AuthGate>
  );

  await waitFor(() => {
    expect(renderResult.getByText('Choose your display name')).toBeTruthy();
    expect(renderResult.queryByText('Feed content')).toBeNull();
  });
});

test('renders children when session exists and display name is set', async () => {
  const session = { access_token: 'tok', user: { id: 'u1' } };
  const player = { id: 'p1', displayName: 'Bobby' };

  const renderResult = await render(
    <AuthGate session={session} player={player}>
      <Text>Feed content</Text>
    </AuthGate>
  );

  await waitFor(() => {
    expect(renderResult.getByText('Feed content')).toBeTruthy();
  });
});