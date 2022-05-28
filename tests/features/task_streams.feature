Feature: Task Streaming
  Clients can connect to the streaming RPC to receive
  tasks as they become ready to process.

  # All scenarios start with a new and empty database.

  Scenario: Streaming endpoint accepts connections
    When I subscribe to tasks in the "campaigns" queue
    Then after 1 seconds I should receive 0 tasks

  Scenario: Stream returns created task when runtime reached
    Given I have created the task:
      """
      {
        "tts_seconds": 1,
        "queue": "campaigns"
      }
      """
    When I subscribe to tasks in the "campaigns" queue
    Then after 2 seconds I should receive the same task

  Scenario: Stream does not redeliver tasks
    Given I have created the task:
      """
      {
        "tts_seconds": 1,
        "queue": "campaigns"
      }
      """
    When I subscribe to tasks in the "campaigns" queue
    Then after 10 seconds I should receive 1 tasks

  Scenario: Stream returns nothing when task not ready
    Given I have created the task:
      """
      {
        "tts_seconds": 5,
        "queue": "campaigns"
      }
      """
    When I subscribe to tasks in the "campaigns" queue
    Then after 2 seconds I should receive 0 tasks

  Scenario: Stream returns task if ready before client connected
    Given I have created the task:
      """
      {
        "tts_seconds": 2,
        "queue": "campaigns"
      }
      """
    When I subscribe to tasks in the "campaigns" queue after a 5 second delay
    Then after 2 seconds I should receive 1 tasks

  Scenario: Stream does not return tasks from other queues
    Given I have created the task:
      """
      {
        "tts_seconds": 1,
        "queue": "campaigns"
      }
      """
    When I subscribe to tasks in the "jobs" queue
    Then after 2 seconds I should receive 0 tasks