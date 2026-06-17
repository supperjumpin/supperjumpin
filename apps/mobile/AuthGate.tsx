import React from 'react';
import LoginScreen from './LoginScreen';
import DisplayNameSetupScreen from './DisplayNameSetupScreen';

interface Session {
  access_token: string;
  user: { id: string };
}

interface Player {
  id: string;
  displayName: string;
}

interface AuthGateProps {
  session: Session | null;
  player: Player | null;
  onSignIn?: () => Promise<void>;
  onDisplayNameSet?: (displayName: string) => Promise<void>;
  children: React.ReactNode;
}

export default function AuthGate({ session, player, onSignIn, onDisplayNameSet, children }: AuthGateProps) {
  if (!session) {
    return <LoginScreen onSignIn={onSignIn || (async () => {})} />;
  }

  if (!player || !player.displayName) {
    return <DisplayNameSetupScreen onSubmit={onDisplayNameSet || (async () => {})} />;
  }

  return <>{children}</>;
}