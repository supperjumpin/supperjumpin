Feature: Jump Lifecycle
  As a player
  I want to take an idea and prove I performed it
  So that it can be judged by my group

  Scenario: Off-season jump — promoted without active season
    Given a player "Bob" exists
    And a group "Midnight Snackers" exists and "Bob" is a member
    And an active season exists for the group
    When "Bob" creates an idea for "Late Night" at "24hr Diner" with "Nachos"
    And "Bob" promotes the idea to a planned jump
    And "Bob" authorizes an upload for "image/jpeg"
    And "Bob" submits evidence with caption "Made it past midnight"
    Then the jump status should be "Performed Jump"

  Scenario: Idea stays as idea — never performed without evidence
    Given a player "Carol" exists
    And a group "Waffle House Warriors" exists and "Carol" is a member
    When "Carol" creates an idea for "Waffle House" at "Waffle House" with "Hashbrowns"
    Then the idea status should be "Idea"

  Scenario: Cannot authorize upload for idea status
    Given a player "Dave" exists
    And a group "Late Night Crew" exists and "Dave" is a member
    When "Dave" creates an idea for "Denny's" at "Denny's" with "Moons Over My Hammy"
    Then the request to authorize upload should fail with status 404

  Scenario: Cannot submit evidence without authorization
    Given a player "Eve" exists
    And a group "Morning Crew" exists and "Eve" is a member
    When "Eve" creates an idea for "Breakfast" at "IHOP" with "Short Stack"
    And "Eve" promotes the idea to a planned jump
    Then the request to submit evidence without authorization should fail

  Scenario: Jump appears in group's recent jumps
    Given a player "Frank" exists
    And a group "Burger Bunch" exists and "Frank" is a member
    When "Frank" creates an idea for "Five Guys" at "Five Guys" with "Bacon Burger"
    And "Frank" promotes the idea to a planned jump
    And "Frank" authorizes an upload for "image/png"
    And "Frank" submits evidence with caption "Double patty proof"
    Then the group's recent jumps should include "Bacon Burger"