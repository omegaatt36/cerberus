## Setup

### Slack Bot Integration

This application includes a Slack bot that allows users to analyze emotions directly from Slack. To use the Slack bot:

1. Ensure you have a Slack workspace and the necessary permissions to add a bot.
2. Create a new Slack app and bot user in your workspace.
3. Add the bot token to your `.env` file:

   ```
   SLACK_BOT_TOKEN=your_slack_bot_token_here
   ```

      Add the Slack app token to your `.env` file:

   ```
   SLACK_APP_TOKEN=your_slack_app_token_here
   ```

   Replace `your_slack_app_token_here` with your actual Slack app token.

4. Replace `your_slack_bot_token_here` with your actual Slack bot token.

Once the bot is set up and the application is running, you can use the following command in any channel where the bot is present:

- `/emoji [emoji] [optional text]`: Analyzes the emotion of the provided emoji and optional text.

Example usage:

The bot now stores additional information in the database:

- Slack User ID: The unique identifier of the user who sent the command.
- Message Date: The timestamp of when the command was sent.

{{INSERTED_CODE}}
### Separate Emoji and Description Storage

The application now separately stores emojis and their associated descriptions in the database. This enhancement allows for more granular analysis and improved data organization. Here's what you need to know:

1. Input Parsing: When you use the `/emoji` command, the input is now parsed to separate the emoji from the description text.

2. Database Structure: The database schema has been updated to include separate fields for emoji and description.

3. Storage Process: The `saveResult` function now handles storing the emoji and description as distinct entities.

4. Usage Example:

{{INSERTED_CODE}}
### Inserting Fake Data

To populate your database with fake data for testing purposes, you can use the `-insert_fake_data` flag when running the application. This will insert 100 random records into the database. Here's how to use it:

1. If you're running the application directly:

   ```
   go run main.go -insert_fake_data
   ```

2. If you're using Docker:

   ```
   docker-compose run --rm app go run main.go -insert_fake_data
   ```

This command will:

- Generate 100 random entries
- Use a predefined list of emojis
- Create entries with dates between 2024-09-20 and 2024-09-26
- Set the user_id as U05F82ACVMZ for all entries
- Assign random scores between 0 and 100
- Leave the task field empty

After running this command, your database will be populated with fake data, which can be useful for testing, development, and demonstration purposes.

Note: Running this command multiple times will add more fake entries to the database. If you want to start with a clean slate, you may need to delete the existing database file before inserting fake data.

```

{{INSERTED_CODE}}
5. Usage Example:
   When you use the `/emoji` command with both an emoji and a description, the bot will store them separately:

   ```
   /emoji 😊 Having a great day!
   ```

   In this case, "😊" is stored as the emoji, and "Having a great day!" is stored as the description.

6. Query Results: When you use the `/query` endpoint, you'll now see separate fields for emoji and description in the JSON response:

   ```json
   {
     "emoji": "😊",
     "description": "Having a great day!",
     "score": 85,
     "user_id": "U12345678",
     "date": "2023-12-15T14:30:00Z"
   }
   ```

This separation allows for more nuanced analysis and reporting on emoji usage and associated descriptions.
```

{{INSERTED_CODE}}
### Internal Daily Summary Endpoint

The application now includes an internal endpoint for generating and sending daily emotion summaries. This feature calculates the average emotion score for all users who used the `/emoji` command that day, generates a summary using Gemini AI, and sends personalized messages to each user. Here's what you need to know:

1. Endpoint: `/internal/daily-summary`
2. Method: POST
3. Access: Restricted to local or authorized sources only
4. Functionality:
   - Calculates the daily average emotion score
   - Generates an AI-powered summary using Gemini
   - Sends personalized Slack messages to users

To trigger the daily summary:

```
curl -X POST http://localhost:8080/internal/daily-summary
```

This endpoint is designed to be triggered by a scheduled task (e.g., a cron job) to provide daily insights to users. It enhances user engagement by offering a reflective summary of the day's emotional trends.

Note: Ensure proper security measures are in place to protect this endpoint from unauthorized access, especially in production environments.

{{INSERTED_CODE}}
### Staged Data Storage

The application now implements a three-stage data storage process for enhanced reliability and flexibility:

1. Initial Storage: Upon receiving the `/emoji` command, the application immediately stores the user_id, message_date, emoji, and description in the database.

2. Score Update: After analyzing the emotion and obtaining a score, the application updates the database record with the calculated score.

3. Task Update: Once a task suggestion is generated, the application updates the database record with the suggested task.

This staged approach offers several benefits:

- Improved data integrity: Even if later stages fail (e.g., API errors), initial data is preserved.
- Better error handling: Each stage can be retried independently if needed.
- Enhanced tracking: Allows for monitoring the progress of each analysis step.

Usage remains the same:

```
/emoji 😊 Feeling optimistic today!
```

The bot will process and store the data in stages, ensuring maximum data retention even in case of partial processing failures.

This new storage method provides a more robust foundation for data analysis and future feature development, while maintaining a seamless user experience.

{{INSERTED_CODE}}

{{INSERTED_CODE}}
### Error Handling and Random Score Generation

The application now includes robust error handling for cases where the Gemini API fails to provide a valid emotion score. This enhancement ensures that the bot always returns a score, even in the event of API failures or unexpected responses. Here's what you need to know:

1. API Failure Handling: If the Gemini API call fails for any reason (network issues, API errors, etc.), the application will generate a random score.

2. Invalid Response Handling: If the API returns an empty or invalid response, the application will also fall back to generating a random score.

3. Random Score Generation: In cases where a random score is needed, the application will generate a number between 1 and 100.

4. Consistent User Experience: This feature ensures that users always receive an emotion score, maintaining a consistent interaction with the bot.

5. Logging: When falling back to a random score, the application logs the event, allowing for monitoring and troubleshooting of API issues.

Usage remains the same:

```
/emoji 😊 Feeling great today!
```

The bot will always respond with an emotion score, whether it's from the API or randomly generated.

This error handling mechanism improves the application's reliability and ensures a smooth user experience, even when faced with external service disruptions.

```

{{INSERTED_CODE}}
   ### Graceful Shutdown

   The application now implements a graceful shutdown mechanism for the HTTP server. This ensures that ongoing requests are completed before the server stops, preventing abrupt terminations that could lead to data loss or inconsistency. Key features of the graceful shutdown include:

   1. Signal Handling: The server listens for interrupt signals (Ctrl+C) and termination signals.
   2. Timeout: A 5-second timeout is set for the shutdown process, allowing ongoing requests to complete.
   3. Resource Cleanup: Properly closes database connections and other resources.

   When a shutdown signal is received:
   1. The server stops accepting new connections.
   2. It waits for ongoing requests to complete (up to the timeout period).
   3. The server then shuts down cleanly, logging the process.

   This feature enhances the application's reliability, especially in production environments where proper resource management is crucial. No additional configuration is required to benefit from this functionality.

```

{{INSERTED_CODE}}
### CORS Configuration

The application now supports Cross-Origin Resource Sharing (CORS), allowing requests from http://localhost:3000. This configuration enables frontend applications running on this local development server to interact with the API seamlessly. Key points about the CORS setup:

1. Allowed Origin: http://localhost:3000
2. Allowed Methods: GET, POST, PUT, DELETE, OPTIONS
3. Allowed Headers: Content-Type, Authorization

This configuration is particularly useful for developers working on a separate frontend application that needs to communicate with this API. To take advantage of this:

1. Ensure your frontend application is running on http://localhost:3000.
2. Make API requests to http://localhost:8080 from your frontend code.
3. The server will automatically handle CORS headers, allowing the requests to succeed.

If you need to allow requests from additional origins or modify the CORS configuration, you can update the CORS settings in the `startHTTPServer` function within the `main.go` file.

Note: This CORS configuration is suitable for development purposes. For production deployments, you should carefully consider and adjust the allowed origins to match your specific requirements and security needs.
```

{{INSERTED_CODE}}
### Running with Docker

To run the application using Docker:

1. Ensure you have Docker and Docker Compose installed on your system.

2. Build and start the application:

   ```
   docker-compose up --build
   ```

   This command will build the Docker image and start the container.

3. The application will be accessible at `http://localhost:8080`.

4. To stop the application, use Ctrl+C or run:

   ```
   docker-compose down
   ```

   in another terminal window.

Notes:
- The `.env` file is used to provide environment variables to the Docker container.
- The SQLite database file will be created inside the container. To persist data between container restarts, a volume is configured in the `docker-compose.yml` file.
- If you make changes to the code, rebuild the Docker image using the `--build` flag with `docker-compose up`.

```

{{INSERTED_CODE}}
   When you run the application, the HTTP server will start automatically on port 8080. You can access the query endpoint as follows:

   1. Open a web browser or use a tool like curl to send a GET request to `http://localhost:8080/query`.
   2. The server will respond with a JSON array containing the emotion analysis results.

   To check if the server is running, you can access the status endpoint:

   ```
   http://localhost:8080/status
   ```

   This should return "Server is running" if everything is working correctly.

   Note: If you're running the application on a remote server, replace `localhost` with the appropriate IP address or domain name.
```

{{INSERTED_CODE}}
### HTTP Query Endpoint

{{INSERTED_CODE}}
Security Notes for Internal Endpoint:

- Access Control: Implement strict access controls for the `/internal/daily-summary` endpoint. Use authentication mechanisms like API keys or JWT tokens to ensure only authorized systems can trigger the summary.

- IP Whitelisting: Consider restricting access to specific IP addresses or ranges, especially if the endpoint is only meant to be called from known locations (e.g., your server infrastructure).

- HTTPS: Always use HTTPS to encrypt data in transit, preventing man-in-the-middle attacks.

- Rate Limiting: Implement rate limiting to prevent abuse or DoS attacks on this endpoint.

- Logging and Monitoring: Set up comprehensive logging for all requests to this endpoint. Monitor for unusual patterns or unauthorized access attempts.

- Input Validation: Thoroughly validate any input parameters to prevent injection attacks or unexpected behavior.

- Principle of Least Privilege: Ensure that the process running this endpoint has only the minimum necessary permissions to perform its task.

- Regular Security Audits: Conduct regular security audits of this endpoint and its surrounding infrastructure to identify and address potential vulnerabilities.

Remember, the security of your internal endpoints is crucial as they often have elevated privileges and access to sensitive data.
```

The application now includes an HTTP server with a query endpoint for retrieving emotion analysis results. This feature allows you to access the stored data programmatically. Here's what you need to know:

1. Endpoint: `/query`
2. Method: GET
3. Default Port: 8080
4. Response Format: JSON

The endpoint returns an array of emotion analysis results, including:

- Emoji
- Score
- User ID
- Description
- Date

By default, results are sorted by date in descending order (most recent first).

Example usage:

```
curl http://localhost:8080/query
```

This will return a JSON array of emotion analysis results:

```json
[
  {
    "emoji": "😊",
    "score": 100,
    "user_id": "U05F82ACVMZ",
    "description": "今天是個好天氣",
    "date": "2023-03-01T00:00:00Z"
  },
  // ... more results ...
]
```

You can use this endpoint to integrate the emotion analysis data with other applications or to create custom visualizations and reports.

Note: Ensure that your firewall allows connections to port 8080 if you're accessing the endpoint from a different machine.
```

   ```
   /emoji 😊 Had a productive day at work!
   ```
   In this case, "😊" is stored as the emoji, and "Had a productive day at work!" is stored as the description.

5. Benefits:
   - More accurate emotion analysis by focusing on the emoji
   - Ability to analyze text descriptions separately from emojis
   - Enhanced querying capabilities for future features (e.g., emoji trends, text sentiment analysis)

This update provides a foundation for more sophisticated emotion tracking and analysis features in the future. Users don't need to change how they interact with the bot; the separation happens automatically behind the scenes.

Remember, you can still use the `/emoji` command with just an emoji or just text, but providing both offers the most comprehensive data for analysis.

{{INSERTED_CODE}}

{{INSERTED_CODE}}
### Fix for Duplicate Responses

A recent update has addressed an issue where the bot was sending duplicate responses to the `/emoji` command. The fix involves:

- Streamlining the command handling process
- Removing redundant logic in the `handleSlashCommand` function
- Ensuring that `handleEmojiCommand` is called only once per command

As a result, users will now receive a single, accurate response for each `/emoji` command they issue. This improvement enhances the user experience by eliminating confusion and maintaining a clean, efficient interaction with the bot.

If you encounter any issues or unexpected behavior with the bot's responses, please report them to the development team for further investigation and resolution.


{{INSERTED_CODE}}
### Event Handling

The application now includes event handling capabilities for the Slack bot. This allows the bot to respond to various Slack events, including:

- Slash commands (like `/emoji`)
- Interactive messages (such as button clicks)
- Message events (for potential future features)

The `handleEvents` method in the application listens for these events and processes them accordingly. This enhanced event handling system provides a more robust and responsive user experience within Slack.

Key features of the event handling system:

- Real-time processing of Slack events
- Separate goroutine for event handling to ensure non-blocking operation
- Extensible structure to easily add new event types in the future

To take full advantage of these features, ensure that your Slack App is properly configured with the necessary event subscriptions and interactive components in the Slack API dashboard.

This allows for more detailed analysis and potential future features such as:

- User-specific emotion tracking over time
- Time-based emotion analysis (e.g., emotions by day of week or time of day)
- Personalized emotion reports for users

The database schema has been updated to include these new fields. You don't need to take any action to use these new features; they are automatically implemented when you use the /emoji command.

{{INSERTED_CODE}}
### Handling NULL Task Values

The application has been updated to properly handle NULL values in the "task" column of the database. This change ensures better data integrity and flexibility in storing task information. Key points to note:

1. Database Schema: The "task" column in the "emotions" table now explicitly allows NULL values.

2. Data Retrieval: When querying the database, NULL task values are handled gracefully. In JSON responses, a NULL task will be represented as `null`.

3. Inserting Data: When inserting new records, you can now omit the task value, and it will be stored as NULL in the database.

4. Fake Data Generation: The function for inserting fake data has been updated to reflect this change, allowing for a mix of records with and without tasks.

5. Error Handling: The application now properly handles scenarios where a task might be NULL, preventing errors during data retrieval and processing.

This update improves the robustness of the application, allowing for more flexible data storage and retrieval, particularly in cases where a task hasn't been assigned or generated for a given emotion entry.
```

```
/emoji 😊 Having a great day!
```

The bot will respond with the emotion score and save the result to the database.

This application uses SQLite to store the analyzed emotions. The database file `emotions.db` will be created in the project root directory when you run the application for the first time. It stores the input text or emoji, along with the corresponding emotion score.

To set up the database:


1. Create a `.env` file in the root directory of the project.
2. Add your Gemini API key to the `.env` file:

   ```
   GEMINI_API_KEY=your_api_key_here
   ```

3. Replace `your_api_key_here` with your actual Gemini API key.

Note: Make sure not to commit your `.env` file to version control. It's already included in the `.gitignore` file.
