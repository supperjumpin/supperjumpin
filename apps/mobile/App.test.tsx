import { render, waitFor } from '@testing-library/react-native';
import App from './App';

// App wraps content in AuthGate. When no SUPABASE_URL env var is set
// (test environment), GoTrueClient is null so there's no session,
// and AuthGate shows LoginScreen.
test('renders sign-in screen when unauthenticated', async () => {
  const renderResult = await render(<App />);

  await waitFor(() => {
    expect(renderResult.getByText('Supperjumpin')).toBeTruthy();
    expect(renderResult.getByText('Sign In')).toBeTruthy();
  }, { timeout: 10000 });
}, 15000);