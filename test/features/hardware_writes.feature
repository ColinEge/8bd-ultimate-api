@hardware
Feature: Physical keyboard binding read-back
  Writes require a separate opt-in and a spare key.
  Each scenario restores and verifies the original mapping or default, including on failure.
  These scenarios verify configuration, not emitted keyboard input or host audio.

  Scenario Outline: Bind and read back supported actions
    When I bind the selected physical key to "<action>"
    Then physical read-back should report "<action>"

    Examples:
      | action      |
      | B           |
      | Ctrl+C      |
      | F13         |
      | disabled    |
      | volume up   |
      | volume down |
      | mute        |
      | B           |

  Scenario: Reset the selected key to its default action
    When I reset the selected physical key to default
    Then its effective action should be the factory default

  Scenario: Preserve a binding across API client recreation
    When I bind the selected physical key to "B"
    And I recreate the API client
    Then physical read-back should report "B"
