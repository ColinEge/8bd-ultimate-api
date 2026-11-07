
@planned
Feature: Profile management
  As a user
  I want to manage keyboard profiles locally
  So that I can maintain multiple configurations independently of the keyboard

  Scenario: Create a profile
    When I create a profile named "Gaming"
    Then a profile file named "Gaming" should exist

  Scenario: List profiles
    Given profiles named "Gaming" and "Work" exist locally
    When I request the available profiles
    Then "Gaming" should be listed
    And "Work" should be listed

  Scenario: Rename a profile
    Given a profile named "Gaming" exists locally
    When I rename the profile to "FPS"
    Then a profile named "FPS" should exist
    And a profile named "Gaming" should no longer exist

  Scenario: Delete a profile
    Given a profile named "Gaming" exists locally
    When I delete the profile
    Then the profile file should no longer exist

  Scenario: Edit a profile
    Given a profile named "Gaming" exists locally
    When I modify its key mappings
    Then the updated configuration should be saved locally

  Scenario: Persist profiles
    Given I have created multiple profiles
    When I restart the application
    Then all profiles should still exist
    And their configurations should be preserved

  Scenario: Read profile from file
    Given a profile named "Gaming" exists locally
    When I load the profile
    Then its configuration should be available in memory

  Scenario: Reject duplicate profile names
    Given a profile named "Gaming" exists locally
    When I create another profile named "Gaming"
    Then the operation should fail

  Scenario: Reject invalid profile data
    Given a profile file contains invalid configuration data
    When I attempt to load the profile
    Then the operation should fail
    And an appropriate error should be returned

  Scenario: Apply a local profile to the keyboard
    Given a profile named "Gaming" exists locally
    And an 8BitDo keyboard is connected
    When I apply the "Gaming" profile
    Then its configuration should be written to the keyboard

  Scenario: Read the keyboard configuration
    Given an 8BitDo keyboard is connected
    When I read its configuration
    Then the device configuration should be returned

  Scenario: Delete the keyboard configuration
    Given an 8BitDo keyboard is connected
    When I delete its stored configuration
    Then the device configuration should be deleted

  Scenario: Applying a profile does not modify the local profile
    Given a profile named "Gaming" exists locally
    And an 8BitDo keyboard is connected
    When I apply the "Gaming" profile
    Then the local profile should remain unchanged
