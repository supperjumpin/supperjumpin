import { jest } from '@jest/globals';

// React 19 requires the test environment to declare act support
(globalThis as any).IS_REACT_ACT_ENVIRONMENT = true;

afterEach(() => {
  jest.restoreAllMocks();
});
