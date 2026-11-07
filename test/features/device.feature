Feature: Device management
  As a user
  I want to manage my connected keyboard
  So that I can configure the correct device

  @api
  Scenario: Detect a connected keyboard
    Given an 8BitDo keyboard is connected
    When I request the connected devices
    Then the keyboard should appear in the device list

  @api
  Scenario: Identify a device model
    Given an 8BitDo keyboard is connected
    When I request its device information
    Then its model should be reported

  @planned
  Scenario: Report firmware version
    Given an 8BitDo keyboard is connected
    When I request its firmware version
    Then its firmware version should be reported

  @api
  Scenario: Report unsupported firmware querying
    Given an 8BitDo keyboard is connected
    When I request its firmware version
    Then firmware querying should report unsupported

  @api
  Scenario: Handle device disconnection
    Given an 8BitDo keyboard is connected
    When the keyboard is disconnected
    Then the device should no longer be available

  @api
  Scenario: Handle device reconnection
    Given an 8BitDo keyboard was previously connected
    And the keyboard has been disconnected
    When the keyboard is reconnected
    Then the device should become available again

  @api
  Scenario: Apply configuration to connected device
    Given an 8BitDo keyboard is connected
    And I have a valid configuration
    When I apply the configuration
    Then the configuration should be written to the keyboard

  @api
  Scenario: Reject configuration without a connected device
    Given no keyboard is connected
    When I attempt to apply a configuration
    Then the operation should fail
    And an appropriate error should be returned
    And the missing device error should be returned without writing
