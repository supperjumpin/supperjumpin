# Gesture-Driven Scoring Shortcuts

## Context
The core gameplay of Supperjumpin requires Players to judge Performed Stunts across four distinct factors: Difficulty, Transgression, Creativity, and Documentation. While precision is important, the social and "absurd" nature of the game benefits from a fast-paced, tactile judging experience. The app must support both touch-screen devices (via Expo) and mouse-based interaction.

## Decision
We will implement a "Quick-Judge" gesture system that acts as a UX shortcut for populating the scoring sheet without bypassing the established game rules or backend constraints.

### Interaction Model
- **Gestures as Shortcuts**: Specific gestures (e.g., "Fire" swipes or directional drags) will populate suggested values across the four scoring factors. For example, a high-energy swipe might set all factors to 4/5.
- **Confirmation Step**: Gestures do not trigger API calls. A **Judgment** is only submitted to the backend when the Player explicitly confirms the values on the detailed scoring sheet.
- **Cross-Platform Compatibility**: Interaction patterns must be implemented using a "drag-and-drop" or "slide" logic that behaves identically for both touch-drag and mouse-click-and-drag.

### Undo and Correction
- **Unsubmitted Scores**: Any values populated by a gesture are local state and can be removed or changed via a "Clear" action or manual adjustment on the scoring sheet.
- **Submitted Scores**: Once a **Judgment** is submitted, it follows the existing domain logic: the Judge may edit their submission until the **Judging Window** closes, or other players may raise a **Dispute**.

## Consequences
- **Pros**: Maintains the high-fidelity four-factor scoring system while providing a modern, "gamified" feel. Prevents accidental submissions. Requires zero changes to the Go backend or database schema.
- **Cons**: Adds a layer of UI complexity to the mobile app. Requires careful mapping of gestures to score values to ensure the "shortcuts" feel intuitive.
