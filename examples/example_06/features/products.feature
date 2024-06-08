Feature: Products
  In order to manage products
  As a product manager
  I want to be able to create and retrieve products

  Scenario: Create a product
    Given I have a product with the id "1"
    And I have a product with the name "product 1"
    When I create the product
    Then the product should be created

  Scenario: Retrieve a product
    Given I have a product with the id "1"
    And I have a product with the name "product 1"
    And I create the product
    When I retrieve the product with the id "1"
    Then the product should be retrieved

  Scenario: Retrieve a non-existent product
    Given I have a product with the id "1"
    And I have a product with the name "product 1"
    And I create the product
    When I retrieve the product with the id "2"
    Then the product should not be retrieved