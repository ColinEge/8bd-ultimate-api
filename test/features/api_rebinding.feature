@api
Feature: Public rebinding API
  As an API consumer
  I want acknowledged bindings and accurate read-back
  So that I can handle success and failure reliably

  Rebinding uses a scripted keyboard, never physical USB hardware.

  Scenario Outline: Bind A to B through a short-lived session
    Given a scripted keyboard with a binding response of "<response>"
    When I bind A to B through the public API
    Then the binding result should be "<result>"
    And the binding session should be closed

    Examples:
      | response      | result                         |
      | ok            | success                        |
      | bad           | device responded bad           |
      | malformed     | unexpected status payload      |
      | write failure | BindAction write failed        |
      | read failure  | BindAction read failed         |

  Scenario: Bind, read, and restore A in one session
    Given a scripted keyboard for a binding round-trip
    When I bind, read, and restore A through the public API
    Then the binding result should be "success"
    And the binding session should be closed

  Scenario Outline: Distinguish an absent override from a binding
    Given a scripted action response of "<response>"
    When I read the scripted action through the public API
    Then the action presence should be "<presence>"
    And the binding session should be closed

    Examples:
      | response | presence |
      | unbound  | absent   |
      | disabled | mapped   |
      | bound    | mapped   |

  Scenario Outline: Configure supported keyboard actions
    Given a scripted keyboard for the action "<action>"
    When I bind the selected action through the public API
    Then the binding result should be "success"
    And the binding session should be closed

    Examples:
      | action       |
      | disabled     |
      | Ctrl+C       |
      | F13          |
      | volume up    |
      | volume down  |
      | mute         |
      | Z            |
      | 0            |
      | F24          |
      | numpad enter |
      | left ctrl    |
      | right meta   |
      | minus        |
      | underscore   |
      | double quote |
      | scroll left  |
      | next track   |

  Scenario Outline: Read and restore physical modifier defaults
    Given a scripted default for modifier key "<key>" at "<id>" with action "<action>"
    When I read and restore the modifier through the public API
    Then the binding result should be "success"
    And the binding session should be closed

    Examples:
      | key         | id   | action   |
      | left ctrl   | 0x64 | 07-e0-00 |
      | left shift  | 0x65 | 07-e1-00 |
      | left alt    | 0x66 | 07-e2-00 |
      | left meta   | 0x67 | 07-e3-00 |
      | right ctrl  | 0x68 | 07-e4-00 |
      | right shift | 0x69 | 07-e5-00 |
      | right alt   | 0x6a | 07-e6-00 |
