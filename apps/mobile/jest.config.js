/** @type {import('jest').Config} */
module.exports = {
  preset: 'jest-expo',
  setupFilesAfterEnv: ['<rootDir>/jest.setup.ts'],
  transformIgnorePatterns: [
    'node_modules/(?!((jest-)?react-native|@react-native(-community)?)|expo(nent)?|@expo(nent)?/.*|@expo-google-fonts/.*|react-navigation|@react-navigation/.*|@sentry/react-native|native-base|react-native-svg|@supperjumpin/.*)',
  ],
  collectCoverageFrom: [
    '**/*.{ts,tsx}',
    '!**/*.test.{ts,tsx}',
    '!jest.setup.ts',
    '!types.ts',
    '!test/**',
    '!**/node_modules/**',
  ],
  coverageDirectory: '../../coverage/jest',
  coverageReporters: ['json-summary', 'text'],
};