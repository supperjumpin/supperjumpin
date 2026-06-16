import { act } from 'react';
import { cleanup, fireEvent, render, waitFor } from '@testing-library/react-native';
import { createJump } from '@supperjumpin/api-client';
import CreateJumpScreen from './CreateJumpScreen';

jest.mock('@supperjumpin/api-client', () => ({
  createJump: jest.fn(),
}));

const mockCreateJump = createJump as jest.MockedFunction<typeof createJump>;

beforeEach(() => {
  mockCreateJump.mockReset();
});

afterEach(() => {
  cleanup();
});

test('keeps Submit Jump disabled until source, destination, food, caption, and a photo are present, then submits once', async () => {
  const onBack = jest.fn();
  mockCreateJump.mockResolvedValue({
    id: 'jump_123',
    playerId: 'player_1',
    status: 'Performed Jump',
    source: 'Taco Bell',
    destination: 'Olive Garden',
    food: 'Crunchwrap',
  } as any);

  const renderResult = await render(
    <CreateJumpScreen session={{ access_token: 'tok_create', user: { id: 'u1' } }} onBack={onBack} />
  );

  const submitButton = renderResult.getByTestId('submit-jump-button');
  expect(submitButton.props.accessibilityState.disabled).toBe(true);

  await act(async () => {
    fireEvent.changeText(renderResult.getByTestId('source-input'), 'Taco Bell');
    fireEvent.changeText(renderResult.getByTestId('destination-input'), 'Olive Garden');
    fireEvent.changeText(renderResult.getByTestId('food-input'), 'Crunchwrap');
    fireEvent.changeText(renderResult.getByTestId('caption-input'), 'Best jump ever');
  });
  expect(renderResult.getByTestId('submit-jump-button').props.accessibilityState.disabled).toBe(true);

  await act(async () => {
    fireEvent.changeText(renderResult.getByTestId('evidence-photo-input'), 'evidence/jump_123.jpg');
  });

  await waitFor(() => {
    expect(renderResult.getByTestId('submit-jump-button').props.accessibilityState.disabled).toBe(false);
  });

  await act(async () => {
    fireEvent.press(renderResult.getByTestId('submit-jump-button'));
  });

  await waitFor(() => {
    expect(onBack).toHaveBeenCalledTimes(1);
  });

  expect(mockCreateJump).toHaveBeenCalledWith(
    expect.objectContaining({
      baseUrl: 'http://localhost:8080',
      accessToken: 'tok_create',
      source: 'Taco Bell',
      destination: 'Olive Garden',
      food: 'Crunchwrap',
      caption: 'Best jump ever',
      mediaObjectKey: 'evidence/jump_123.jpg',
    })
  );
});

test('does not submit device-local evidence photo URIs as persisted object keys', async () => {
  const onBack = jest.fn();
  const renderResult = await render(
    <CreateJumpScreen session={{ access_token: 'tok_create', user: { id: 'u1' } }} onBack={onBack} />
  );

  await act(async () => {
    fireEvent.changeText(renderResult.getByTestId('source-input'), 'Taco Bell');
    fireEvent.changeText(renderResult.getByTestId('destination-input'), 'Olive Garden');
    fireEvent.changeText(renderResult.getByTestId('food-input'), 'Crunchwrap');
    fireEvent.changeText(renderResult.getByTestId('caption-input'), 'Best jump ever');
    fireEvent.changeText(renderResult.getByTestId('evidence-photo-input'), 'file:///var/mobile/photo.jpg');
  });

  expect(renderResult.getByTestId('submit-jump-button').props.accessibilityState.disabled).toBe(true);

  await act(async () => {
    fireEvent.press(renderResult.getByTestId('submit-jump-button'));
  });

  expect(mockCreateJump).not.toHaveBeenCalled();
  expect(onBack).not.toHaveBeenCalled();
});

test('renders the create form in a scrollable container', async () => {
  const renderResult = await render(
    <CreateJumpScreen session={{ access_token: 'tok_create', user: { id: 'u1' } }} onBack={jest.fn()} />
  );

  expect(renderResult.getByTestId('create-jump-scroll')).toBeTruthy();
});
