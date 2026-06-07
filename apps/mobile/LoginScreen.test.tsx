import { render, fireEvent, waitFor } from '@testing-library/react-native';
import LoginScreen from './LoginScreen';

test('renders sign-in button', async () => {
  const renderResult = await render(<LoginScreen onSignIn={async () => {}} />);
  await waitFor(() => {
    expect(renderResult.getByText('Sign In')).toBeTruthy();
  });
});

test('calls onSignIn when button pressed', async () => {
  const onSignIn = jest.fn().mockResolvedValue(undefined);
  const renderResult = await render(<LoginScreen onSignIn={onSignIn} />);

  await waitFor(() => {
    const button = renderResult.getByText('Sign In');
    fireEvent.press(button);
  });

  expect(onSignIn).toHaveBeenCalledTimes(1);
});