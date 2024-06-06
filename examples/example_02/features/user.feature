Feature: User management
  In order to manage users
  As a manager
  I want to be able to create, read, update and delete users

  Scenario: Create an user
    Given I have entered "John" into the user name field
    And I have entered "john@example.com" into the user email field
    When I press "create"
    Then the user should be created

  Scenario: Read an user
    Given I have an existing user
    And I have entered "john@example.com" into the user email field
    When I press "read"
    Then the user should be read

  Scenario: Update an user
    Given I have an existing user
    And I have entered "John Doe" into the user name field
    And I have entered "john@example.com" into the user email field
    When I press "update"
    Then the user should be updated

  Scenario: Delete an user
    Given I have an existing user
    And I have entered "john@example.com" into the user email field
    When I press "delete"
    Then the user should be deleted