Feature: Calculator
  In order to do math
  As a math student
  I want to be able to do operations with numbers

  Scenario: Add two numbers
    Given I have entered "50" into the calculator
    And I have entered "70" into the calculator
    When I press "add"
    Then the result should be "120" on the screen

  Scenario: Subtract two numbers
    Given I have entered "100" into the calculator
    And I have entered "40" into the calculator
    When I press "subtract"
    Then the result should be "60" on the screen

  Scenario: Multiply two numbers
    Given I have entered "5" into the calculator
    And I have entered "6" into the calculator
    When I press "multiply"
    Then the result should be "30" on the screen

  Scenario: Divide two numbers
    Given I have entered "10" into the calculator
    And I have entered "2" into the calculator
    When I press "divide"
    Then the result should be "5" on the screen