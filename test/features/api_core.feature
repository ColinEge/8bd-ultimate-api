@api
Feature: Core keyboard API contract
  As an API consumer
  I want reliable keyboard operations and explicit failures
  So that I can configure a device without protocol commands

  All scenarios use a scripted device provider, not USB or host input.

  Background:
    Given an 8BitDo keyboard is connected

  Scenario: Use independent built-in catalogs
    When I inspect and mutate copies of the public catalogs
    Then the catalogs should remain deterministic and unchanged
    And no packets should be written
    And exactly 1 keyboard connection should be closed

  Scenario Outline: Reject invalid changes before any write
    Given a batch containing "<invalid>"
    When I apply the prepared batch
    Then the operation should fail
    And the API error should be "<error>"
    And no packets should be written
    And exactly 1 keyboard connection should be closed

    Examples:
      | invalid             | error          |
      | invalid key         | invalid key    |
      | zero action         | invalid action |
      | unsupported action  | unsupported    |
      | invalid second key  | invalid key    |
      | duplicate key       | invalid key    |

  Scenario: Apply a batch in order without modifying its input
    Given a scripted batch with a "success" result
    When I apply the prepared batch
    Then the API operation should succeed
    And the supplied batch should remain unchanged
    And all scripted packets should be consumed
    And exactly 1 keyboard connection should be closed

  Scenario Outline: Report partial application without retrying
    Given a scripted batch with a "<failure>" result
    When I apply the prepared batch
    Then exactly the first binding should be acknowledged in the apply error
    And the API error should be "<error>"
    And the supplied batch should remain unchanged
    And all scripted packets should be consumed
    And exactly 1 keyboard connection should be closed

    Examples:
      | failure       | error       |
      | rejection     | rejected    |
      | write failure | transport   |
      | read failure  | transport   |

  Scenario Outline: Read typed binding states
    Given a scripted read-back of "<state>"
    When I read a binding through the public API
    Then its stored and effective actions should match "<state>"
    And exactly 1 keyboard connection should be closed

    Examples:
      | state    |
      | default  |
      | disabled |
      | action   |

  Scenario Outline: Reject malformed binding replies
    Given a scripted read-back of "<reply>"
    When I read a binding through the public API
    Then the API error should be "malformed"
    And no binding value should be returned
    And exactly 1 keyboard connection should be closed

    Examples:
      | reply        |
      | truncated    |
      | wrong key    |
      | wrong header |
      | oversized    |

  Scenario Outline: List stored overrides
    Given a scripted override list of "<kind>"
    When I list bindings through the public API
    Then the override list result should be "<kind>"
    And all scripted packets should be consumed
    And exactly 1 keyboard connection should be closed

    Examples:
      | kind      |
      | empty     |
      | mapped    |
      | malformed |

  Scenario: Reject a cancelled operation before I/O
    When I bind with an already cancelled context
    Then the API error should be "cancelled"
    And no packets should be written
    And exactly 1 keyboard connection should be closed

  Scenario Outline: Time out blocked command I/O
    When I run a command with a blocked "<direction>"
    Then the API error should be "deadline"
    And the blocked command should stop within its timeout budget

    Examples:
      | direction |
      | write     |
      | read      |

  Scenario: Serialize concurrent operations and cancel the waiting caller
    When a second binding waits behind a blocked write
    Then the waiting caller should time out without interleaving packets
    And the blocked caller should retain cancellation identity

  Scenario Outline: Preserve callback and cleanup errors
    When a keyboard callback ends with "<outcome>"
    Then the callback and cleanup error identities should be preserved
    And exactly 1 keyboard connection should be closed

    Examples:
      | outcome          |
      | callback error   |
      | panic            |
      | cleanup error    |
      | callback+cleanup |

  Scenario: Client cleanup owns open keyboards
    When I close the client with an open keyboard
    Then the client and keyboard should reject further operations
    And repeated close calls should not close resources again
