# GO-OpenAI-Bot-Discord

## Quick start

1. Make sure you have docker CLI and make installed

1. Rename `credentials_example.yaml` to `credentials.yaml`. Fill in the required fields (like discord token and OpenAI token)

1. Simply run `make execute`

    > ***Note:*** Your bot must have `Message Content Intent` permission enabled in the Discord dev portal. We need to read messages in the threads to have a proper AI conversation.
    
    > ***Note:*** Make sure you interact with the bot after adding it to a channel, don't DM the bot, the functionality won't work
    
    > ***Note:*** use this link to invite the bot to your workspace -> https://discord.com/api/oauth2/authorize?client_id=<your client ID>&permissions=8&scope=bot

1. `/info` in your server to list bot info such as version and commands
