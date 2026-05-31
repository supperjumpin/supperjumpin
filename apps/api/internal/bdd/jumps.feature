Feature: Jump Lifecycle
  As a player
  I want to take an idea and prove I performed it
  So that it can be judged by my group

  Scenario: Successful Jump Performance
    Given a player "Alice" exists
    And a group "Taco Bell Daredevils" exists and "Alice" is a member
    And an active season exists for the group
    When "Alice" creates an idea for "Taco Bell" at "Olive Garden" with "Crunchwrap"
    And "Alice" promotes the idea to a planned jump
    And "Alice" authorizes an upload for "image/jpeg"
    And "Alice" submits evidence with caption "Security gave me a look"
    Then the jump status should be "Performed Jump"
