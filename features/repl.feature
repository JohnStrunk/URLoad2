Feature: urload2 REPL application

  Rule: When the REPL starts, the REPL shall display an interactive prompt.

    Scenario: Prompt is displayed upon starting
      Given the REPL application is initialized
      When the REPL is started with no input
      Then the output should contain "urload2> "

  Rule: While running, when the user inputs the exit command, the REPL shall terminate gracefully.

    Scenario: Exit command terminates the session
      Given the REPL application is initialized
      When the user enters "exit"
      Then the output should contain "Goodbye!"
      And the REPL session should terminate successfully

    Scenario: Quit command terminates the session
      Given the REPL application is initialized
      When the user enters "quit"
      Then the output should contain "Goodbye!"
      And the REPL session should terminate successfully

  Rule: While running, when the user inputs the help command, the REPL shall display available commands.

    Scenario: Help command displays command list
      Given the REPL application is initialized
      When the user enters "help"
      Then the output should contain "Available commands:"
      And the output should contain "help"
      And the output should contain "exit"

  Rule: While running, when the user inputs the "?" command, the REPL shall display available commands.

    Scenario: Question mark command displays command list
      Given the REPL application is initialized
      When the user enters "?"
      Then the output should contain "Available commands:"
      And the output should contain "help, ?"
      And the output should contain "exit"

  Rule: While the user is entering input, when completion is requested, the REPL shall return matching command candidates.

    Scenario: Tab completion returns matches for a prefix
      Given the REPL application is initialized
      When the user requests completion for "he"
      Then the completion candidates should include "help"

    Scenario: Tab completion returns all matches for empty input
      Given the REPL application is initialized
      When the user requests completion for ""
      Then the completion candidates should include "help"
      And the completion candidates should include "?"
      And the completion candidates should include "version"
      And the completion candidates should include "exit"
      And the completion candidates should include "quit"

  Rule: If an unrecognized command is entered, then the REPL shall display an error message and continue running.

    Scenario: Unrecognized command displays error
      Given the REPL application is initialized
      When the user enters "unknown_cmd"
      Then the output should report unknown command "unknown_cmd"
