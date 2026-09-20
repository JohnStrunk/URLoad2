Feature: urload2 REPL application

  Rule: While running, when displaying the prompt, the REPL shall include the directory id and the current length of the URL list.

    Scenario: Prompt displays initial count of zero
      Given the REPL application is initialized
      When the REPL is started with no input
      Then the output should contain "URLoad2 [0000] (0)> "

    Scenario: Prompt updates dynamically as URLs are added
      Given the REPL application is initialized
      When the user enters "add http://example.com/1"
      And the user enters "add http://example.com/2"
      Then the output should contain "URLoad2 [0000] (1)> "
      And the output should contain "URLoad2 [0000] (2)> "

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
      And the output should contain "add <url>"
      And the output should contain "get"
      And the output should contain "head <n>"
      And the output should contain "tail <n>"
      And the output should contain "list"
      And the output should contain "clear"
      And the output should contain "help"
      And the output should contain "href"
      And the output should contain "img"
      And the output should contain "sort"
      And the output should contain "uniq"
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
      And the completion candidates should include "head"

    Scenario: Tab completion returns all matches for empty input
      Given the REPL application is initialized
      When the user requests completion for ""
      Then the completion candidates should include "add"
      And the completion candidates should include "clear"
      And the completion candidates should include "exit"
      And the completion candidates should include "get"
      And the completion candidates should include "head"
      And the completion candidates should include "help"
      And the completion candidates should include "href"
      And the completion candidates should include "img"
      And the completion candidates should include "list"
      And the completion candidates should include "quit"
      And the completion candidates should include "sort"
      And the completion candidates should include "tail"
      And the completion candidates should include "uniq"
      And the completion candidates should include "version"
      And the completion candidates should include "?"

  Rule: If an unrecognized command is entered, then the REPL shall display an error message and continue running.

    Scenario: Unrecognized command displays error
      Given the REPL application is initialized
      When the user enters "unknown_cmd"
      Then the output should report unknown command "unknown_cmd"

  Rule: When the user enters the add command with a URL, the REPL shall append the URL to the end of the URL list.

    Scenario: Add URL appends to list
      Given the REPL application is initialized
      When the user enters "add http://example.com/first"
      And the user enters "add http://example.com/second"
      And the user enters "list"
      Then the output should contain "http://example.com/first"
      And the output should contain "http://example.com/second"

  Rule: If the add command is entered without a URL, then the REPL shall display an error message and continue running.

    Scenario: Add command without argument displays error
      Given the REPL application is initialized
      When the user enters "add"
      Then the output should contain "error: add requires a URL"

  Rule: When the user enters the list command, the REPL shall display the current list of URLs.

    Scenario: List command displays all stored URLs
      Given the REPL application is initialized
      When the user enters "add http://example.com/item1"
      And the user enters "list"
      Then the output should contain "http://example.com/item1"

  Rule: When the user enters the clear command, the REPL shall remove all URLs from the URL list.

    Scenario: Clear command removes all URLs from list
      Given the REPL application is initialized
      When the user enters "add http://example.com/item1"
      And the user enters "clear"
      Then the output should contain "URLoad2 [0000] (0)> "

  Rule: When the user enters the head command with a count, the REPL shall keep the first n URLs in the list and discard the rest.

    Scenario: Head command keeps first n URLs
      Given the REPL application is initialized
      When the user enters "add http://example.com/1"
      And the user enters "add http://example.com/2"
      And the user enters "add http://example.com/3"
      And the user enters "head 2"
      And the user enters "list"
      Then the output should contain "http://example.com/1"
      And the output should contain "http://example.com/2"
      And the output should not contain "http://example.com/3"

  Rule: If the head command is entered without a valid count, then the REPL shall display an error message and continue running.

    Scenario: Head command without count displays error
      Given the REPL application is initialized
      When the user enters "head"
      Then the output should contain "error: head requires a count"

  Rule: When the user enters the tail command with a count, the REPL shall keep the last n URLs in the list and discard the rest.

    Scenario: Tail command keeps last n URLs
      Given the REPL application is initialized
      When the user enters "add http://example.com/1"
      And the user enters "add http://example.com/2"
      And the user enters "add http://example.com/3"
      And the user enters "tail 2"
      And the user enters "list"
      Then the output should not contain "http://example.com/1"
      And the output should contain "http://example.com/2"
      And the output should contain "http://example.com/3"

  Rule: If the tail command is entered without a valid count, then the REPL shall display an error message and continue running.

    Scenario: Tail command without count displays error
      Given the REPL application is initialized
      When the user enters "tail"
      Then the output should contain "error: tail requires a count"

  Rule: When the REPL starts, the REPL shall determine the download target directory as one greater than the highest numbered directory starting at 0000.

    Scenario: Target directory is determined upon startup
      Given a working directory with directory "0002"
      When the REPL application is initialized in that working directory
      Then the download target directory name should be "0003"

  Rule: While running without executing the get command, the REPL shall not create the download target directory.

    Scenario: Download target directory is not created before get
      Given the REPL application is initialized
      When the user enters "add http://example.com/1"
      Then the download target directory should not exist

  Rule: When the user enters the get command, the REPL shall create the target directory and save each downloaded URL as a file.

    Scenario: Get command downloads URLs and saves them to target directory
      Given a test web server serving "hello world" at "/page.html"
      And the REPL application is initialized
      When the user enters "add <server>/page.html"
      And the user enters "get"
      Then the download target directory should exist
      And the file "page.html" in the target directory should contain "hello world"

  Rule: When the user enters the sort command, the REPL shall sort the URLs in the list alphabetically.

    Scenario: Sort command orders URLs alphabetically
      Given the REPL application is initialized
      When the user enters "add http://example.com/gamma"
      And the user enters "add http://example.com/alpha"
      And the user enters "add http://example.com/beta"
      And the user enters "sort"
      And the user enters "list"
      Then the output should contain "http://example.com/alpha"
      And the output should contain "http://example.com/beta"
      And the output should contain "http://example.com/gamma"

  Rule: When the user enters the uniq command, the REPL shall remove duplicate URLs from the list while preserving their original order.

    Scenario: Uniq command removes duplicates preserving first occurrence
      Given the REPL application is initialized
      When the user enters "add http://example.com/one"
      And the user enters "add http://example.com/two"
      And the user enters "add http://example.com/one"
      And the user enters "add http://example.com/three"
      And the user enters "add http://example.com/two"
      And the user enters "uniq"
      And the user enters "list"
      Then the output should contain "http://example.com/one"
      And the output should contain "http://example.com/two"
      And the output should contain "http://example.com/three"
      And the output should contain "URLoad2 [0000] (3)> "

  Rule: When the user enters the href command, the REPL shall retrieve each URL in the list, extract all anchor href targets as absolute URLs, append them to the list, and remove the original URLs.

    Scenario: Href command extracts link targets and replaces original URLs
      Given a test web server serving "<a href='/about'>About</a><a href='https://external.example.com/link'>External</a>" at "/index.html"
      And the REPL application is initialized
      When the user enters "add <server>/index.html"
      And the user enters "href"
      And the user enters "list"
      Then the output should not contain "<server>/index.html"
      And the output should contain "<server>/about"
      And the output should contain "https://external.example.com/link"

  Rule: If retrieving a URL fails during the href command, then the REPL shall display an error message and continue processing remaining URLs.

    Scenario: Href command reports error on failed URL retrieval
      Given the REPL application is initialized
      When the user enters "add http://127.0.0.1:1/nonexistent"
      And the user enters "href"
      Then the output should contain "error extracting hrefs from http://127.0.0.1:1/nonexistent"

  Rule: When the user enters the img command, the REPL shall retrieve each URL in the list, extract all image src targets as absolute URLs, append them to the list, and remove the original URLs.

    Scenario: Img command extracts image targets and replaces original URLs
      Given a test web server serving "<img src='/images/pic.png'><img src='https://cdn.example.com/logo.jpg'>" at "/gallery.html"
      And the REPL application is initialized
      When the user enters "add <server>/gallery.html"
      And the user enters "img"
      And the user enters "list"
      Then the output should not contain "<server>/gallery.html"
      And the output should contain "<server>/images/pic.png"
      And the output should contain "https://cdn.example.com/logo.jpg"

  Rule: If retrieving a URL fails during the img command, then the REPL shall display an error message and continue processing remaining URLs.

    Scenario: Img command reports error on failed URL retrieval
      Given the REPL application is initialized
      When the user enters "add http://127.0.0.1:1/nonexistent"
      And the user enters "img"
      Then the output should contain "error extracting images from http://127.0.0.1:1/nonexistent"
