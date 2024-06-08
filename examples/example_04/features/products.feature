Feature: Products
  In order to manage products
  As a user
  I want to be able to create, update and delete products

  Scenario: Create a product
    Given I have a product with name "Test Product"
    And I have a product with description "Test Product Description"
    When I create the product
    Then the product should be created

  Scenario: Update a product
    Given I have a product with name "Test Product"
    And I have a product with description "Test Product Description"
    And I create the product
    And I update the product name to "Updated Product Name"
    When I update the product
    Then the product should be updated

  Scenario: Delete a product
    Given I have a product with name "Test Product"
    And I have a product with description "Test Product Description"
    And I create the product
    When I delete the product
    Then the product should be deleted