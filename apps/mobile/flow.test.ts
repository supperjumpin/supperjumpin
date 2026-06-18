import { describe, test, expect } from '@jest/globals';
import { createAppFlow } from './flow';

describe('AppFlow', () => {
  test('initial state is feed screen with no session and empty draft', () => {
    const flow = createAppFlow();
    const snapshot = flow.getSnapshot();

    expect(snapshot.screen).toEqual({ name: 'feed' });
    expect(snapshot.session).toBeNull();
    expect(snapshot.player).toBeNull();
    expect(snapshot.showDisplayNameSetup).toBe(false);
    expect(snapshot.pendingCreate).toBe(false);
    expect(snapshot.draft).toEqual({
      source: '',
      destination: '',
      food: '',
      caption: '',
      mediaObjectKey: '',
    });
  });

  test('navigateToCreate when unauthenticated sets pendingCreate without navigating', () => {
    const flow = createAppFlow();
    flow.navigateToCreate();

    const snapshot = flow.getSnapshot();
    expect(snapshot.screen).toEqual({ name: 'feed' });
    expect(snapshot.pendingCreate).toBe(true);
    expect(snapshot.showDisplayNameSetup).toBe(false);
  });

  test('signIn + resolvePlayer with display name resolves pending create to create screen', () => {
    const flow = createAppFlow();
    flow.navigateToCreate();

    flow.signIn({ access_token: 'tok', user: { id: 'p1' } });
    flow.resolvePlayer({ id: 'p1', displayName: 'alice' });

    const snapshot = flow.getSnapshot();
    expect(snapshot.screen).toEqual({ name: 'create' });
    expect(snapshot.pendingCreate).toBe(false);
    expect(snapshot.session).toEqual({ access_token: 'tok', user: { id: 'p1' } });
    expect(snapshot.player).toEqual({ id: 'p1', displayName: 'alice' });
  });

  test('signIn + resolvePlayer without display name shows display name setup', () => {
    const flow = createAppFlow();
    flow.navigateToCreate();

    flow.signIn({ access_token: 'tok', user: { id: 'p1' } });
    flow.resolvePlayer({ id: 'p1', displayName: '' });

    const snapshot = flow.getSnapshot();
    expect(snapshot.screen).toEqual({ name: 'feed' });
    expect(snapshot.pendingCreate).toBe(true);
    expect(snapshot.showDisplayNameSetup).toBe(true);
    expect(snapshot.player).toEqual({ id: 'p1', displayName: '' });
  });

  test('setDisplayName clears setup screen and navigates to create when pending', () => {
    const flow = createAppFlow();
    flow.navigateToCreate();
    flow.signIn({ access_token: 'tok', user: { id: 'p1' } });
    flow.resolvePlayer({ id: 'p1', displayName: '' });

    flow.setDisplayName('alice');

    const snapshot = flow.getSnapshot();
    expect(snapshot.showDisplayNameSetup).toBe(false);
    expect(snapshot.pendingCreate).toBe(false);
    expect(snapshot.screen).toEqual({ name: 'create' });
    expect(snapshot.player).toEqual({ id: 'p1', displayName: 'alice' });
  });

  test('navigateToDetail sets screen to detail with jumpId', () => {
    const flow = createAppFlow();
    flow.navigateToDetail('jump_42');

    const snapshot = flow.getSnapshot();
    expect(snapshot.screen).toEqual({ name: 'detail', jumpId: 'jump_42' });
  });

  test('navigateBack returns to feed from any screen', () => {
    const flow = createAppFlow();
    flow.navigateToDetail('jump_42');
    flow.navigateBack();

    expect(flow.getSnapshot().screen).toEqual({ name: 'feed' });
  });

  test('draft is preserved across create → back → create roundtrip', () => {
    const flow = createAppFlow();
    flow.signIn({ access_token: 'tok', user: { id: 'p1' } });
    flow.resolvePlayer({ id: 'p1', displayName: 'alice' });
    flow.navigateToCreate();

    flow.changeDraft({
      source: 'Taco Bell',
      destination: 'Olive Garden',
      food: 'Crunchwrap',
      caption: 'Keep this',
      mediaObjectKey: 'evidence/1.jpg',
    });

    flow.navigateBack();
    expect(flow.getSnapshot().screen).toEqual({ name: 'feed' });

    flow.navigateToCreate();
    expect(flow.getSnapshot().screen).toEqual({ name: 'create' });

    const snapshot = flow.getSnapshot();
    expect(snapshot.draft).toEqual({
      source: 'Taco Bell',
      destination: 'Olive Garden',
      food: 'Crunchwrap',
      caption: 'Keep this',
      mediaObjectKey: 'evidence/1.jpg',
    });
  });

  test('submitDraft clears the draft', () => {
    const flow = createAppFlow();
    flow.signIn({ access_token: 'tok', user: { id: 'p1' } });
    flow.resolvePlayer({ id: 'p1', displayName: 'alice' });
    flow.navigateToCreate();

    flow.changeDraft({
      source: 'Taco Bell',
      destination: 'Olive Garden',
      food: 'Crunchwrap',
      caption: 'Great jump',
      mediaObjectKey: 'evidence/1.jpg',
    });

    flow.submitDraft();

    expect(flow.getSnapshot().draft).toEqual({
      source: '',
      destination: '',
      food: '',
      caption: '',
      mediaObjectKey: '',
    });
  });

  test('rejectPlayer clears session', () => {
    const flow = createAppFlow();
    flow.signIn({ access_token: 'tok', user: { id: 'p1' } });
    flow.rejectPlayer();

    const snapshot = flow.getSnapshot();
    expect(snapshot.session).toBeNull();
    expect(snapshot.player).toBeNull();
  });

  test('navigateToCreate when signed in with display name navigates directly', () => {
    const flow = createAppFlow();
    flow.signIn({ access_token: 'tok', user: { id: 'p1' } });
    flow.resolvePlayer({ id: 'p1', displayName: 'alice' });
    flow.navigateToCreate();

    const snapshot = flow.getSnapshot();
    expect(snapshot.screen).toEqual({ name: 'create' });
    expect(snapshot.pendingCreate).toBe(false);
  });

  test('navigateToCreate when signed in but no player sets pendingCreate', () => {
    const flow = createAppFlow();
    flow.signIn({ access_token: 'tok', user: { id: 'p1' } });
    // player not yet resolved
    flow.navigateToCreate();

    const snapshot = flow.getSnapshot();
    expect(snapshot.screen).toEqual({ name: 'feed' });
    expect(snapshot.pendingCreate).toBe(true);
  });

  test('navigateToCreate when signed in but no display name shows setup', () => {
    const flow = createAppFlow();
    flow.signIn({ access_token: 'tok', user: { id: 'p1' } });
    flow.resolvePlayer({ id: 'p1', displayName: '' });
    flow.navigateToCreate();

    const snapshot = flow.getSnapshot();
    expect(snapshot.screen).toEqual({ name: 'feed' });
    expect(snapshot.pendingCreate).toBe(true);
    expect(snapshot.showDisplayNameSetup).toBe(true);
  });
});
