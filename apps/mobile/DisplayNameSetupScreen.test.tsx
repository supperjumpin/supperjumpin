import { render, fireEvent, waitFor, act } from '@testing-library/react-native';
import DisplayNameSetupScreen from './DisplayNameSetupScreen';

test('renders text input and submit button', async () => {
  const renderResult = await render(
    <DisplayNameSetupScreen onSubmit={async () => {}} />
  );

  await waitFor(() => {
    expect(renderResult.getByPlaceholderText('Enter your display name')).toBeTruthy();
    expect(renderResult.getByText('Save')).toBeTruthy();
  });
});

test('shows validation error for empty name', async () => {
  const renderResult = await render(
    <DisplayNameSetupScreen onSubmit={async () => {}} />
  );

  const saveButton = renderResult.getByText('Save');
  await act(async () => {
    fireEvent.press(saveButton);
  });

  await waitFor(() => {
    expect(renderResult.getByText('Display name is required')).toBeTruthy();
  });
});

test('calls onSubmit with the entered display name', async () => {
  const onSubmit = jest.fn().mockResolvedValue(undefined);
  const renderResult = await render(
    <DisplayNameSetupScreen onSubmit={onSubmit} />
  );

  const input = renderResult.getByPlaceholderText('Enter your display name');
  await act(async () => {
    fireEvent.changeText(input, 'Bobby Cloutier');
  });

  const saveButton = renderResult.getByText('Save');
  await act(async () => {
    fireEvent.press(saveButton);
  });

  await waitFor(() => {
    expect(onSubmit).toHaveBeenCalledWith('Bobby Cloutier');
  });
});

test('shows error state on API failure', async () => {
  const onSubmit = jest.fn().mockRejectedValue(new Error('API error'));
  const renderResult = await render(
    <DisplayNameSetupScreen onSubmit={onSubmit} />
  );

  const input = renderResult.getByPlaceholderText('Enter your display name');
  await act(async () => {
    fireEvent.changeText(input, 'Bobby');
  });

  const saveButton = renderResult.getByText('Save');
  await act(async () => {
    fireEvent.press(saveButton);
  });

  await waitFor(() => {
    expect(renderResult.getByText('Could not save display name')).toBeTruthy();
  });
});