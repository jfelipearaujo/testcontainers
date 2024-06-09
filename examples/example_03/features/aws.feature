Feature: AWS
  In order to manage AWS resources
  As a developer
  I want to be able to interact with AWS resources

  Scenario: Publish into SNS
    Given I have an SNS topic named "TestTopic"
    When I publish a message into the topic
    Then the message should be published into the topic

  Scenario: Publish into SQS
    Given I have an SQS queue named "TestQueue"
    When I publish a message into the queue
    Then the message should be published into the queue

  Scenario: Read from SQS
    Given I have an SQS queue named "TestQueue"
    When I publish a message into the queue
    Then the message should be published into the queue
    When I read the messages from the queue
    Then I should read "1" message