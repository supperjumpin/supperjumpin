# Explicit Backend Owns Game Rules

Supperjumpin will use an explicit backend API from day one rather than letting the mobile app write domain state directly to a backend-as-a-service. The backend owns Stunt lifecycle transitions, Group membership, Season state, judging eligibility, score aggregation, and House Rules boundaries so the mobile app can iterate quickly without becoming the source of truth for game behavior.
