Feature: Task Streaming
  Clients can connect to the streaming RPC to receive
  tasks as they become ready to process.

  # All scenarios start with a new and empty database.

  Scenario: Streaming endpoint accepts connections
    When I begin streaming after a 0 second delay
    Then after 1 seconds I should receive 0 tasks

  Scenario: Stream returns created task when runtime reached
    Given I have created the task:
      """
      {
        "request_id": "task_streaming",
        "tts_seconds": 1,
        "namespace": "campaigns"
      }
      """
    When I begin streaming after a 0 second delay
    Then after 2 seconds I should receive the same task

  Scenario: Stream does not redeliver tasks
    Given I have created the task:
      """
      {
        "request_id": "task_streaming",
        "tts_seconds": 1,
        "namespace": "campaigns"
      }
      """
    When I begin streaming after a 0 second delay
    Then after 10 seconds I should receive 1 tasks

  Scenario: Stream returns nothing when task not ready
    Given I have created the task:
      """
      {
        "request_id": "task_streaming",
        "tts_seconds": 5,
        "namespace": "campaigns"
      }
      """
    When I begin streaming after a 0 second delay
    Then after 2 seconds I should receive 0 tasks

  Scenario: Stream returns task if ready before client connected
    Given I have created the task:
      """
      {
        "request_id": "task_streaming",
        "tts_seconds": 2,
        "namespace": "campaigns"
      }
      """
    When I begin streaming after a 5 second delay
    Then after 1 seconds I should receive 1 tasks