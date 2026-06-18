import { render, screen, waitFor } from '@testing-library/react-native';
import React from 'react';
import { Text } from 'react-native';
import { useLiveNow } from './useLiveNow';

function TestTicker({ enabled }: { enabled: boolean }) {
  const now = useLiveNow(enabled);
  return <Text testID="tick">{String(now)}</Text>;
}

test('returns current timestamp on initial render', async () => {
  const now = Date.now();
  await render(<TestTicker enabled={false} />);
  const value = Number(screen.getByTestId('tick').props.children);
  expect(value).toBeGreaterThanOrEqual(now);
});

test('returns updated timestamp when enabled', async () => {
  jest.useFakeTimers();
  jest.setSystemTime(new Date('2026-06-01T00:00:00Z'));

  await render(<TestTicker enabled={true} />);

  expect(screen.getByTestId('tick').props.children).toBe(String(Date.now()));

  jest.advanceTimersByTime(1000);
  await waitFor(() => {
    expect(screen.getByTestId('tick').props.children).toBe(String(Date.now()));
  });

  jest.useRealTimers();
});

test('stops updating when disabled', async () => {
  jest.useFakeTimers();
  jest.setSystemTime(new Date('2026-06-01T00:00:00Z'));

  const { rerender } = await render(<TestTicker enabled={true} />);

  jest.advanceTimersByTime(1000);
  await waitFor(() => {
    expect(screen.getByTestId('tick').props.children).toBe(String(Date.now()));
  });

  rerender(<TestTicker enabled={false} />);

  const frozen = screen.getByTestId('tick').props.children;

  jest.advanceTimersByTime(5000);
  await waitFor(() => {
    expect(screen.getByTestId('tick').props.children).toBe(frozen);
  });

  jest.useRealTimers();
});

test('does not update when disabled', async () => {
  jest.useFakeTimers();
  jest.setSystemTime(new Date('2026-06-01T00:00:00Z'));

  await render(<TestTicker enabled={false} />);

  const initial = screen.getByTestId('tick').props.children;

  jest.advanceTimersByTime(5000);

  expect(screen.getByTestId('tick').props.children).toBe(initial);

  jest.useRealTimers();
});
