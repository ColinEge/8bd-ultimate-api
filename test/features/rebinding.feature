
Feature: Key rebinding
  As a user
  I want to customise keyboard key mappings
  So that I can configure the keyboard to suit my needs

  Background:
    Given an 8BitDo keyboard is connected

  @planned
  Scenario: Rebind a key
    When I map the physical key "A" to "B"
    Then pressing the physical key "A" should produce "B"

  @planned
  Scenario: Reset a key mapping to default
    Given the physical key "A" is mapped to "B"
    When I reset its mapping to default
    Then pressing the physical key "A" should produce "A"

  @planned
  Scenario: Disable a key
    When I disable the physical key "A"
    Then pressing the physical key "A" should produce no input

  @planned
  Scenario: Map a key combination
    When I map the physical key "A" to "Ctrl+C"
    Then pressing the physical key "A" should produce "Ctrl+C"

  @planned
  Scenario: Map a special key
    When I map the physical key "A" to "F13"
    Then pressing the physical key "A" should produce "F13"

  @api
  Scenario: Reject an invalid mapping
    When I attempt to map a physical key to an unsupported key
    Then the mapping operation should fail
    And an appropriate error should be returned

  @planned
  Scenario: Persist key mappings
    Given the physical key "A" is mapped to "B"
    When I restart the application
    Then pressing the physical key "A" should produce "B"

  @planned
  Scenario: Apply a selected profile to the keyboard
    Given a profile named "Default" exists
    And "Default" is the selected profile
    And I have changed multiple key mappings
    When I apply the selected profile
    Then all updated mappings should be written to the keyboard

  @api
  Scenario Outline: Read back a key mapping
    When I map the physical key "A" to "<action>"
    Then the physical key "A" should read back "<action>"

    Examples:
      | action |
      | B      |
      | Ctrl+C |
      | F13    |

  @api
  Scenario: Read back a key reset to default
    Given the physical key "A" is mapped to "B"
    When I reset its mapping to default
    Then the physical key "A" should read back "A"

  @api
  Scenario: Read back a disabled key
    When I disable the physical key "A"
    Then the physical key "A" should read back disabled

  @api
  Scenario: Apply key mappings without a profile
    Given I have changed multiple key mappings
    When I apply the changed key mappings
    Then all updated mappings should be written to the keyboard

  @api
  Scenario: Read the stored keyboard bindings
    Given the physical key "A" is mapped to "B"
    When I read the keyboard bindings
    Then the returned bindings should contain the stored mapping
