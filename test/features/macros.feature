
@planned
Feature: Macro management
  As a user
  I want to create and assign macros
  So that I can automate sequences of keyboard inputs

  Background:
    Given an 8BitDo keyboard is connected

  Scenario: Create a macro
    When I create a macro named "Copy"
    Then the macro should exist

  Scenario: Rename a macro
    Given a macro named "Copy" exists
    When I rename the macro to "Paste"
    Then the macro should be named "Paste"

  Scenario: Delete a macro
    Given a macro named "Copy" exists
    When I delete the macro
    Then the macro should no longer exist

  Scenario: Record a macro
    Given I am recording a macro named "Copy"
    When I press Ctrl
    And I press C
    And I stop recording
    Then the macro should contain the recorded inputs in order

  Scenario: Preserve macro delays
    Given a macro contains multiple inputs with delays
    When I save the macro
    Then the delays should be preserved

  Scenario: Edit a macro
    Given a macro named "Copy" exists
    When I change its input sequence
    Then the updated sequence should be saved

  Scenario: Assign a macro to a key
    Given a macro named "Copy" exists
    When I assign the macro to a keyboard key
    Then the key should reference that macro

  Scenario: Remove a macro assignment
    Given a macro is assigned to a keyboard key
    When I remove the assignment
    Then the key should no longer reference the macro

  Scenario: Persist macros
    Given I have created and saved a macro
    When I restart the application
    Then the macro should still exist
