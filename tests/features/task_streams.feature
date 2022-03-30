Feature: Task Streaming

  Scenario: Streaming endpoint accepts connections
    When I begin streaming after a 0 second delay
    Then after 1 seconds I should receive 0 tasks

  Scenario: Stream returns created task when runtime reached
    Given I have created the task:
      """
      {
        "request_id": "task_streaming_process_task_when_ready",
        "tts_seconds": 2,
        "namespace": "campaigns"
      }
      """
    When I begin streaming after a 0 second delay
    Then after 5 seconds I should receive the same task
    And after 10 seconds I should receive 1 tasks