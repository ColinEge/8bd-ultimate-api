
Feature: Audio controls
  As a keyboard user
  I want to control system audio using my keyboard
  So that I can adjust audio without leaving my current application

  Background:
    Given an 8BitDo keyboard is connected

  @planned
  Scenario: Increase system volume
    When I trigger the volume up action
    Then the system volume should increase

  @planned
  Scenario: Decrease system volume
    When I trigger the volume down action
    Then the system volume should decrease

  @planned
  Scenario: Mute system audio
    When I trigger the mute action
    Then system audio should be muted

  @planned
  Scenario: Unmute system audio
    Given system audio is muted
    When I trigger the mute action
    Then system audio should be unmuted

  @api
  Scenario: Apply audio configuration
    Given I have configured audio actions
    When I apply the configuration to the keyboard
    Then the keyboard should use the updated audio configuration
