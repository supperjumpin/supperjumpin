     1|# Gesture-Driven Scoring Shortcuts
     2|
     3|## Context
     4|The core gameplay of Supperjumpin requires Players to judge Performed Stunts across four distinct factors: Commitment, Transgression, Creativity, and Documentation. While precision is important, the social and "absurd" nature of the game benefits from a fast-paced, tactile judging experience. The app must support both touch-screen devices (via Expo) and mouse-based interaction.
     5|
     6|## Decision
     7|We will implement a "Quick-Judge" gesture system that acts as a UX shortcut for populating the scoring sheet without bypassing the established game rules or backend constraints.
     8|
     9|### Interaction Model
    10|- **Gestures as Shortcuts**: Specific gestures (e.g., "Fire" swipes or directional drags) will populate suggested values across the four scoring factors. For example, a high-energy swipe might set all factors to 4/5.
    11|- **Confirmation Step**: Gestures do not trigger API calls. A **Judgment** is only submitted to the backend when the Player explicitly confirms the values on the detailed scoring sheet.
    12|- **Cross-Platform Compatibility**: Interaction patterns must be implemented using a "drag-and-drop" or "slide" logic that behaves identically for both touch-drag and mouse-click-and-drag.
    13|
    14|### Undo and Correction
    15|- **Unsubmitted Scores**: Any values populated by a gesture are local state and can be removed or changed via a "Clear" action or manual adjustment on the scoring sheet.
    16|- **Submitted Scores**: Once a **Judgment** is submitted, it follows the existing domain logic: the Judge may edit their submission until the **Judging Window** closes, or other players may raise a **Dispute**.
    17|
    18|## Consequences
    19|- **Pros**: Maintains the high-fidelity four-factor scoring system while providing a modern, "gamified" feel. Prevents accidental submissions. Requires zero changes to the Go backend or database schema.
    20|- **Cons**: Adds a layer of UI complexity to the mobile app. Requires careful mapping of gestures to score values to ensure the "shortcuts" feel intuitive.
    21|