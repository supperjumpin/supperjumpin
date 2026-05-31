Feature: Judging
  As a judge
  I want to score a performed jump
  So that it can be ranked against other jumps

  Scenario: Judge can score a performed jump
    Given a player "Gina" exists
    And a group "Score Seekers" exists and "Gina" is a member
    And an active season exists for the group
    When "Gina" creates an idea for "Chick-fil-A" at "Chick-fil-A" with "Spicy Deluxe"
    And "Gina" promotes the idea to a planned jump
    And "Gina" authorizes an upload for "image/jpeg"
    And "Gina" submits evidence with caption "Spicy indeed"
    Then the jump status should be "Performed Jump"

  Scenario: Judgment scores must be between 0 and 10
    Given a player "Hank" exists
    And a group "Rules Crew" exists and "Hank" is a member
    And an active season exists for the group
    When "Hank" creates an idea for "Burger King" at "Burger King" with "Whopper"
    And "Hank" promotes the idea to a planned jump
    And "Hank" authorizes an upload for "image/jpeg"
    And "Hank" submits evidence with caption "Big meal"
    Then the request to submit judgment with commitment 15 should fail with status 400