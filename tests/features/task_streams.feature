Feature: Task Streaming
  Scenario: Process task when ready
    Given I have created the task:
      """
      {
        "request_id": "task_streaming_process_task_when_ready",
        "tts_seconds": 2,
        "namespace": "campaigns"
      }
      """
    When I begin streaming after a 1 second delay
    Then after 1 second I should receive the same task
    And after 5 seconds I should receive 0 tasks