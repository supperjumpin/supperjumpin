# Supperjumpin: Judge-to-Performer Transition Analysis

## Executive Summary

The central challenge for Supperjumpin is converting passive Judges (consumers) into active Jump performers (creators) as the platform scales from First Playable (10-20 seed Players) to MVP (100-500 Players). Research across social apps, UGC platforms, and game ecosystems reveals that **the transition is not automatic—it must be architected through specific mechanics that lower friction, create reciprocity obligations, and provide status rewards for creation.**

At small scale (100-500 users), the 90-9-1 rule relaxes to approximately **70-25-5** (70% lurk, 25% contribute occasionally, 5% create most content). This is Supperjumpin's target conversion window. The key insight: **Judging is already a form of micro-contribution.** Users who rate Jumps are not pure lurkers—they're in the 9% occasional contributor category. The path to performer is shorter than it appears, but it requires deliberate bridge mechanics.

---

## 1. The Judge-to-Performer Funnel

### Current State Analysis

| Stage | User Type | Action | Friction | Conversion Rate |
|-------|-----------|--------|----------|-----------------|
| 1 | Guest Judge | Views Jump, taps to Judge | Very Low | 100% (core loop) |
| 2 | Guest Judge | Submits Judgment scores | Low | ~80% of viewers |
| 3 | Account Holder | Creates Account (MVP) | Medium | ~30-40% of engaged Judges |
| 4 | First-Time Performer | Performs first Jump | High | ~5-15% of Account Holders |
| 5 | Repeat Performer | Performs 2+ Jumps | High | ~30-40% of first-timers |

**Critical Drop-off:** Stage 3→4. This is where most platforms lose users. The gap between "I rate content" and "I create content" feels enormous to users.

### The Bridge: Micro-Contribution Ladder

Research on community activation loops shows that users who make micro-contributions are **3-4x more likely** to make larger contributions within 30 days. For Supperjumpin, the Judging action itself is a micro-contribution. The task is to **escalate the commitment gradually**:

1. **Judge** (existing) → 2. **Comment on Jump** → 3. **Suggest Jump idea** → 4. **Co-perform Jump** → 5. **Solo Jump**

---

## 2. Concrete Mechanics & Patterns

### Pattern A: Post-to-View / Give-to-Get (BeReal Model)

**Mechanic:** Users must contribute to consume. On BeReal, you can't see friends' posts until you post your own. This creates a reciprocity norm that eliminates pure lurking.

**Supperjumpin Application:**
- **"Judge to Unlock":** Guest Judges get 3 free Judgments per session. To see more Jumps or detailed Judgment breakdowns, they must perform a Jump or invite a friend.
- **"The Reveal":** After Judging, show the Judge how their scores compare to others—but only if they've performed at least one Jump themselves. This creates information asymmetry that motivates creation.
- **Reciprocity Dashboard:** Show users who Judged their Jumps. Prompt: "3 people judged your Jump. Judge theirs to see what they thought of yours."

**Why it works at small scale:** At 100-500 users, social accountability is high. Users know their friends are watching. The "everyone participates" norm is enforceable because the community is small enough that non-participation is visible.

### Pattern B: Foot-in-the-Door Escalation

**Mechanic:** Start with trivially easy creation actions, then gradually increase the ask. Research shows reaction-based asks see **5-8x higher response rates** than open-ended prompts.

**Supperjumpin Application:**
- **"Caption This Jump":** Before full Judging, ask users to suggest a funny caption for a Jump photo. One text field, no stakes.
- **"Rate the Location":** Ask users to tag where they think the Jump happened. Low effort, builds investment.
- **"Jump Challenge":** Weekly prompt: "Take a photo of food somewhere it doesn't belong—any food, any place." The prompt removes the creative burden of "what should I do?"
- **"Duet Jump":** Allow users to perform a Jump with a friend. Shared effort, shared credit, lower individual friction.

**Conversion data:** Communities using weekly micro-challenges see **12-18% weekly active user growth** and **35-50% recycling rates** (first-time contributors making a second contribution within 30 days).

### Pattern C: Recognition & Status Incentives

**Mechanic:** Publicly highlight contributions to create social proof and status competition. Research on BoardGameGeek shows that **peer rewards (likes, tips) lead to more, longer, and higher-quality content** than platform rewards (badges, money).

**Supperjumpin Application:**
- **"Judge of the Week":** Highlight the user whose Judgments most closely matched the community average (wisdom of crowds). Display on leaderboard.
- **"Jump Critic" Badges:** Award titles based on Judging volume and consistency: "Apprentice Critic" (10 Judgments), "Master Critic" (100 Judgments), "Grand Arbiter" (500 Judgments).
- **"Spotlight Performer":** When a user performs their first Jump, feature it prominently with a "First Jump" badge. Make a big deal of it.
- **"The Open" Standings (MVP):** Monthly competition with public leaderboards. Categories: Best Jump, Funniest Jump, Most Creative Jump. Winners get profile badges and featured placement.

**Critical insight:** Recognition is most effective when it's **specific and timely**. "Great Jump!" is weak. "Your Commitment score of 4.0 was the highest this week—here's why it resonated" is strong.

### Pattern D: Social Accountability & Peer Pressure

**Mechanic:** Make participation visible to friends. Users are **3-4x more likely** to contribute when asked within 24 hours of consuming value, especially if friends will see their contribution (or lack thereof).

**Supperjumpin Application:**
- **"Your Friends Are Judging":** Show users which of their friends have Judged recently. Social proof that "everyone's doing it."
- **"Jump Streaks":** Track consecutive weeks with at least one Jump. Display streak prominently on profile. Loss aversion is powerful.
- **"Group Challenges":** "Your friend group has performed 12 Jumps this month. Can you get to 20?" Collective goals with individual contribution visibility.
- **"The Open" Teams (MVP):** Allow users to form teams for monthly competitions. Team members can see who hasn't contributed yet. Gentle peer pressure.

### Pattern E: Lowering the Creation Barrier

**Mechanic:** Reduce the real-world effort required to perform a Jump. The research on BeReal shows that **constraints increase participation** by removing decision paralysis.

**Supperjumpin Application:**
- **"Jump Kits":** Provide suggested combinations: "Take a banana to the gym," "Bring coffee to a meeting," "Eat pizza in a library." Remove the "what should I do?" friction.
- **"Template Jumps":** Pre-defined Jump formats with clear instructions. Users just execute.
- **"Jump Reminders":** Push notification: "It's been 5 days since your last Jump. Your friends are waiting." (Use sparingly—weekly max.)
- **"Low-Stakes First Jump":** Explicitly label the first Jump as "practice" or "warm-up." It doesn't count toward standings. Removes performance anxiety.

### Pattern F: Attention Bartering (Reciprocal Engagement)

**Mechanic:** Research on social media economics shows that users engage in "attention bartering"—mutual following and engagement to build audience. Platforms that facilitate this see higher production rates.

**Supperjumpin Application:**
- **"Judge Back":** When someone Judges your Jump, prompt you to Judge theirs. Creates reciprocal obligation.
- **"Jump Exchange":** "I'll Judge your Jump if you Judge mine." Formalized bartering system.
- **"Mentor Matches":** Pair experienced Jumpers with new users. Experienced users get status; new users get guidance and social connection.

---

## 3. Application to Supperjumpin's Two Stages

### First Playable Loop (Guest Judges, No Accounts)

**Goal:** ≥1.0 Judgments per Jump within 7 days.

**Judge-to-Performer Mechanics (Stage 1):**

1. **Seeded Content + "Judge to Unlock"**
   - Seed 20-30 Jumps from founders/friends.
   - Guest Judges get 3 free Judgments per session.
   - To see more Jumps or "community average" scores, they must perform a Jump (guest submission with optional email for notification).
   - This creates the give-to-get dynamic even without accounts.

2. **"Caption This" Micro-Contributions**
   - On every Jump, add a "Suggest a caption" field below the Judging panel.
   - Zero friction, builds psychological investment.
   - Top captions get highlighted on the Jump page.

3. **"Jump Challenge of the Week"**
   - Weekly prompt displayed prominently: "This week: Take breakfast food to a place of work."
   - Removes creative burden.
   - Guest submissions accepted with simple form (photo + location + optional email).

4. **Social Proof Dashboard**
   - Show real-time counter: "47 people have Judged this Jump."
   - Show "Most Active Judges this week" list.
   - Creates FOMO and norm of participation.

**Success Metrics:**
- Guest-to-submitter conversion rate: Target 5-10% of Guest Judges submit at least one Jump.
- Micro-contribution rate: Target 30% of Judges leave a caption or comment.

### MVP (Authentication, Accounts, "The Open")

**Goal:** 100-500 Players with sustainable performer conversion.

**Judge-to-Performer Mechanics (Stage 2):**

1. **Account Creation Triggers**
   - Prompt account creation after 5th Judgment: "Create an account to save your Judging history and track your Jump streak."
   - Prompt after first Jump submission: "Create an account to see how people rated your Jump and get notified of new Jumps."
   - One-tap social login (as planned) removes friction.

2. **"The Open" as Conversion Engine**
   - Monthly competition with **multiple entry categories**:
     - Best Overall Jump
     - Funniest Jump
     - Most Creative Location
     - Best Presentation
     - "Rookie of the Month" (first-time Jumpers only)
   - **Critical:** "Rookie of the Month" category lowers the barrier for first-time performers. They compete against other rookies, not veterans.
   - Public leaderboards with profile badges.
   - Winners featured in app launch screen and push notifications.

3. **Gamification Layer**
   - **Levels/XP:** Judge Jumps → earn XP → level up → unlock profile customization, early access to new features, "Judge" titles.
   - **Streaks:** Daily/weekly Jump streaks with escalating rewards.
   - **Collections:** "You've Judged 50 Jumps. Collect all 100 to earn 'Centurion' badge."
   - **Team Competitions:** Users form teams of 3-5. Team with most Jumps in a month wins. Peer pressure + collective identity.

4. **"Jump Mentor" Program**
   - Identify top 10% of Jumpers (by volume and average Judgment scores).
   - Offer them "Mentor" status: exclusive badge, early feature access, ability to feature mentee Jumps.
   - Mentors get matched with new users who haven't performed a Jump yet.
   - Mentors earn XP for each mentee's first Jump.
   - **Why:** Top performers often enjoy teaching. New users get social connection + guidance.

5. **Reciprocity Mechanics**
   - "You Judged 5 Jumps this week. 3 of those Jumpers haven't been Judged much—Judging them gives 2x XP."
   - "Your Jump got 12 Judgments. Return the favor and Judge 12 Jumps to earn 'Reciprocal' badge."
   - "Friend Challenge: You and [Friend] have both Judged each other's Jumps. Keep it going for a 7-day streak!"

6. **Friction Reduction**
   - "Jump Kits" with suggested ideas (as above).
   - "Quick Jump" mode: Pre-selected food + location combo. User just takes photo.
   - "Jump Reminders" based on user's past behavior (not generic broadcast).
   - "Draft Jumps": Save a Jump idea and come back to it. Reduces "I don't have time" abandonment.

**Success Metrics:**
- Account creation rate: 30-40% of engaged Guest Judges.
- First Jump rate: 15-25% of Account Holders.
- Second Jump rate: 35-50% of first-time Jumpers.
- Monthly active Jumpers: 20-30% of total user base (small community target).

---

## 4. Addressing the Judging Labor Ceiling

**The Tension:** In First Playable, users only Judge. In MVP, users need to perform Jumps. But if performers outnumber active Judges, Jumps go un-Judged, performers get no feedback, and the loop breaks.

### The Math

At 500 users with 70-25-5 distribution:
- ~350 lurkers (mostly pure consumers, some may Judge occasionally)
- ~125 occasional contributors (Judges)
- ~25 heavy creators (performers)

If each performer does 2 Jumps/week = 50 Jumps/week.
If each Judge does 10 Judgments/session, 3 sessions/week = 3,750 Judgments/week.
**Judgment capacity: 3,750. Judgment demand: 50 Jumps × 4 factors × ~10 Judges per Jump = 2,000.**

**At 500 users, the labor ceiling is not the immediate problem.** The problem is **distribution**—ensuring each Jump gets enough Judgments to feel meaningful.

### Solutions

1. **Judgment Quotas & Incentives**
   - To perform a Jump, users must first Judge 5 Jumps. This ensures labor supply before demand.
   - "You have 3 Jumps pending Judgment. Judge them to unlock your next Jump submission."
   - This is the BeReal model applied to labor economics.

2. **Algorithmic Distribution (MVP)**
   - New Jumps get distributed to active Judges first.
   - "Jump Roulette": Randomly assign un-Judged Jumps to users who haven't Judged recently.
   - "Bounty Jumps": Jumps that haven't been Judged in 48 hours offer 2x XP for Judging.

3. **Reduce Judgments per Jump**
   - Instead of requiring 10+ Judgments for meaningful average, show preliminary scores after 3 Judgments.
   - "Early scores" update in real-time as more Judgments come in.
   - Performers get faster feedback, reducing the "my Jump is ignored" abandonment.

4. **NPC/Seed Judgments (First Playable)**
   - Founders/team manually Judge every Jump in First Playable.
   - Ensures no Jump goes un-Judged.
   - Sets the standard for what good Judging looks like.

5. **"The Open" as Labor Redistribution**
   - During "The Open," all users are encouraged to Judge more (competition entries need Judgments).
   - Offer bonus XP for Judging "Open" entries.
   - Creates periodic spikes in Judging activity that match spikes in Jump creation.

---

## 5. Recommended Implementation Roadmap

### Phase 1: First Playable (Weeks 1-4)

**Week 1-2: Core Loop Validation**
- Implement "Judge to Unlock" (3 free Judgments, then prompt to submit Jump).
- Add "Caption This" micro-contribution.
- Seed 20-30 Jumps manually.
- Track: Guest-to-submitter conversion rate.

**Week 3-4: Social Proof & Challenges**
- Add "Jump Challenge of the Week" prompt.
- Add "Most Active Judges" leaderboard.
- Add real-time Judgment counters on Jumps.
- Track: Micro-contribution rate, Judgments per Jump.

### Phase 2: MVP Launch (Weeks 5-8)

**Week 5-6: Account Creation & Onboarding**
- Implement one-tap social login.
- Add account creation triggers (after 5th Judgment, after first Jump submission).
- Add XP/Level system for Judging.
- Track: Account creation rate, first Jump rate.

**Week 7-8: "The Open" & Gamification**
- Launch first "Open" competition with Rookie category.
- Implement streaks, badges, team competitions.
- Add "Judgment Quota" (must Judge 5 to submit 1).
- Track: Monthly active Jumpers, second Jump rate, Judgments per Jump.

### Phase 3: Optimization (Weeks 9-12)

**Week 9-10: Reciprocity & Mentorship**
- Add "Judge Back" prompts.
- Launch "Jump Mentor" program.
- Add "Jump Kits" and "Quick Jump" mode.
- Track: Reciprocal Judging rate, mentor-mentee conversion.

**Week 11-12: Labor Ceiling Management**
- Implement algorithmic Jump distribution.
- Add "Bounty Jumps" for un-Judged content.
- Reduce minimum Judgments for preliminary scores.
- Track: Average Judgments per Jump, time-to-first-Judgment, performer retention.

---

## 6. Key Risks & Mitigations

| Risk | Evidence | Mitigation |
|------|----------|------------|
| **Judging feels like work** | BeReal's daily notification became "homework" for some users | Keep Judging sessions short (<2 mins). Don't require Judgments—make them optional but rewarding. |
| **Performers outnumber Judges** | Open Question #3 in roadmap | Implement Judgment Quota. Use algorithmic distribution. Ensure founders Judge everything in First Playable. |
| **First Jump anxiety** | Research shows "fear of posting" is #1 lurker barrier | Label first Jump as "practice." Offer Jump Kits. Enable co-Jumps with friends. |
| **Recognition fatigue** | Over-indexing on 1% creates community calcification | Rotate spotlights. Feature new contributors weekly. Celebrate "most improved" not just "best." |
| **Monetary rewards reduce quality** | BoardGameGeek study: platform rewards → shorter, less complex content | Use peer rewards (likes, visibility) not platform rewards (money) for Jumpers. Status > cash at small scale. |

---

## 7. Summary: The Core Insight

**Supperjumpin's Judges are not lurkers—they're already contributors.** The platform's core loop (Judging) is a micro-contribution that builds psychological investment. The transition to performer is not about motivating disengaged users; it's about **removing the perceived gap between "I rate things" and "I make things."**

The most effective pattern is **BeReal's "post-to-view" mechanic**: make consumption contingent on creation. At small scale, this is socially enforceable because the community is tight-knit. At MVP scale, gamification ("The Open," streaks, badges) and reciprocity ("Judge Back," quotas) maintain the balance.

**The Judging labor ceiling is solvable** at 100-500 users through quotas, algorithmic distribution, and founder seeding. The bigger risk is **performer abandonment due to lack of feedback**—which is solved by faster preliminary scores, guaranteed founder Judgments, and social proof.

**Bottom line:** Don't treat Judges and Performers as separate user types. Treat them as **stages on a ladder**, and build mechanics that make each rung feel like the natural next step.
