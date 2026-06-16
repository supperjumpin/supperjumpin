import { jest } from '@jest/globals';

process.env.EXPO_PUBLIC_MEDIA_BASE_URL = 'http://media.test';

const originalError = console.error.bind(console);
beforeAll(() => {
  console.error = (...args: unknown[]) => {
    const msg = typeof args[0] === 'string' ? args[0] : '';
    if (
      msg.includes('not configured to support act(') ||
      msg.includes('was not wrapped in act(')
    ) {
      return;
    }
    originalError(...args);
  };
});
afterAll(() => {
  console.error = originalError;
});

afterEach(() => {
  jest.restoreAllMocks();
});
