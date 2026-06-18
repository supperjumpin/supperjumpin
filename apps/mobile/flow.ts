export interface Session {
  access_token: string;
  user: { id: string };
}

export interface Player {
  id: string;
  displayName: string;
}

export interface JumpDraft {
  source: string;
  destination: string;
  food: string;
  caption: string;
  mediaObjectKey: string;
}

type Screen =
  | { name: 'feed' }
  | { name: 'detail'; jumpId: string }
  | { name: 'create' };

export interface FlowSnapshot {
  screen: Screen;
  session: Session | null;
  player: Player | null;
  showDisplayNameSetup: boolean;
  pendingCreate: boolean;
  draft: JumpDraft;
}

function emptyDraft(): JumpDraft {
  return {
    source: '',
    destination: '',
    food: '',
    caption: '',
    mediaObjectKey: '',
  };
}

export function createAppFlow() {
  let screen: Screen = { name: 'feed' };
  let session: Session | null = null;
  let player: Player | null = null;
  let showDisplayNameSetup = false;
  let pendingCreate = false;
  let draft = emptyDraft();

  return {
    getSnapshot(): FlowSnapshot {
      return {
        screen,
        session,
        player,
        showDisplayNameSetup,
        pendingCreate,
        draft,
      };
    },

    signIn(s: Session): void {
      session = s;
    },

    resolvePlayer(p: Player): void {
      player = p;
      if (pendingCreate) {
        if (p.displayName) {
          pendingCreate = false;
          screen = { name: 'create' };
        } else {
          showDisplayNameSetup = true;
        }
      }
    },

    setDisplayName(displayName: string): void {
      if (!session || !player) return;
      player = { ...player, displayName };
      showDisplayNameSetup = false;
      if (pendingCreate) {
        pendingCreate = false;
        screen = { name: 'create' };
      }
    },

    navigateToDetail(jumpId: string): void {
      screen = { name: 'detail', jumpId };
    },

    navigateBack(): void {
      screen = { name: 'feed' };
    },

    navigateToCreate(): void {
      if (!session) {
        pendingCreate = true;
        return;
      }
      if (!player) {
        pendingCreate = true;
        return;
      }
      if (!player.displayName) {
        pendingCreate = true;
        showDisplayNameSetup = true;
        return;
      }
      screen = { name: 'create' };
    },

    changeDraft(d: JumpDraft): void {
      draft = d;
    },

    submitDraft(): void {
      draft = emptyDraft();
    },

    rejectPlayer(): void {
      session = null;
      player = null;
    },
  };
}
