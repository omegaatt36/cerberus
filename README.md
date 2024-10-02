# Cerberus - Emotional Intelligence Slack Bot

Cerberus is a Slack bot designed to help users track their emotions and receive personalized suggestions for maintaining or improving their mood. It uses AI-powered analysis to interpret emoji reactions and text descriptions, providing insightful and sometimes humorous responses.

## Features

- Emotion tracking with emoji and text descriptions
- AI-powered emotion score analysis
- Personalized task suggestions based on emotional state
- Daily emotional summaries
- Integration with Slack for seamless user interaction

## Prerequisites

Before you begin, ensure you have the following installed:
- Go 1.23
- Docker and Docker Compose (for development environment)
- Slack Bot Token and App Token
- Google Cloud Project with Gemini API enabled

## Setup

1. Clone the repository:
   ```
   git clone https://github.com/omegaatt36/cerberus.git
   cd cerberus
   ```

2. Set up environment variables:
   Create a `.env` file in the root directory and add the following:
   ```
   SLACK_BOT_TOKEN=your_slack_bot_token
   SLACK_APP_TOKEN=your_slack_app_token
   GEMINI_API_KEY=your_gemini_api_key
   GEMINI_MODEL=gemini-1.5-flash
   DB_DIALECT=postgres
   DB_HOST=localhost
   DB_PORT=5432
   DB_NAME=cerberus
   DB_USER=postgres
   DB_PASSWORD=postgres
   ```

3. Start the development database:
   ```
   docker-compose -f deploy/dev/docker-compose.yaml up -d
   ```

4. Run database migrations:
   ```
   go run cmd/cerberus.dbmigration/main.go
   ```

5. Build and run the Cerberus bot:
   ```
   go run cmd/cerberus/main.go
   ```

## Usage

Once the bot is running and added to your Slack workspace, you can interact with it using the `/emoji` slash command:

```
/emoji :happy: Feeling great today!
```

The bot will analyze your emotion and provide a personalized response with suggestions.

## Development

To contribute to Cerberus, please follow these steps:

1. Fork the repository
2. Create a new branch for your feature
3. Implement your changes
4. Write tests for your new functionality
5. Ensure all tests pass
6. Submit a pull request

## License

[MIT License](LICENSE)

## Acknowledgments

- Slack API
- Google Gemini AI
- GORM
- Urfave CLI
