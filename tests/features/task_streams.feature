Feature: Task Streaming
  Scenario: Process task when ready
    Given I have created the task:
      """
      {
        "request_id": "task_streaming_process_task_when_ready",
        "tts_seconds": 5,
        "namespace": "campaigns"
      }
      """
    When I begin streaming after a 1 second delay
    Then after 6 seconds I should receive the same task
    And after 10 seconds I should receive 1 tasks