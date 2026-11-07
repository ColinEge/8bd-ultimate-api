@hardware
Feature: Read-only physical keyboard API
  This suite only uses the Go API. It does not capture host input or audio.
  An explicit opt-in, device ID and firmware key ID are required.

  Scenario: Discover and identify the selected physical keyboard
    When I scan and open the selected physical keyboard
    Then its connection information should match discovery
    And firmware querying should explicitly report unsupported

  Scenario: Read a physical key binding consistently
    When I read the selected key twice through separate connections
    Then the physical binding state and actions should be consistent
