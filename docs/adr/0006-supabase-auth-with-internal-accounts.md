# Supabase Auth with Internal Accounts

Supperjumpin will use Supabase Auth for MVP authentication, including email magic links and social login providers where practical, while keeping an internal Account and Player model in the Go backend. External auth provider identities attach to Accounts; they do not define in-game Player identity, Group membership, Season history, or Stunt ownership, preserving a migration path to additional or replacement login methods later.
