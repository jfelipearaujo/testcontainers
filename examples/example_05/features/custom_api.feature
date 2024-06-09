Feature: Custom API
  In order to manage products
  As a user
  I want to be able to access a custom API

  Scenario: Get a product
    Given I have a product with name "Test Product"
    When I get the product
    Then the product should be returned